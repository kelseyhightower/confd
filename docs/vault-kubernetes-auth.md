These are steps to get vault with Kubernetes auth working on minikube.

*Do not use default service account in prod instead create dedicated acount for Vault auth.*

- Deploy Helm
  ```
  # Install Helm
  Use the correct method for your OS from https://docs.helm.sh/using_helm/#installing-the-helm-client
  # Deploy tiller into the cluster
  helm init

- Install Vault in dev mode
  ```
  # Add Vault chart
  helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
  # Install Vault
  # We need at least Vault 0.8.3
  helm install incubator/vault --name vault --set vault.dev=true --set image.tag="0.9.5"
  ```

- Enable Kubernetes backend
  ```
  # Get Vault pod name
  export POD_NAME=$(kubectl get pods --namespace default -l "app=vault" -o jsonpath="{.items[0].metadata.name}")
  # Get inside pod
  kubectl exec -i -t ${POD_NAME} sh
  # Set env vars for Vault client
  export VAULT_TOKEN=$(cat /root/.vault-token)
  # Set Vault host URL (do this everytime you exec back into container)
  export VAULT_ADDR=http://127.0.0.1:8200
  # Enable Kube auth backend
  vault auth enable kubernetes
  # Configure Kube auth backend
  vault write auth/kubernetes/config \
    kubernetes_host=https://kubernetes \
    kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  # Create Vault policy for testing
  vault policy write test -<<EOF
  path "secret/*" {
    capabilities = ["create"]
  }

  path "secret/foo" {
    capabilities = ["read"]
  }
  EOF

  # Create role for confd
  vault write auth/kubernetes/role/confd \
    bound_service_account_names=vault-auth \
    bound_service_account_namespaces=default \
    policies=test \
    ttl=1h
  # Write test secret
  vault write secret/foo value=bar

  # Now exit vault container
  exit
  ```
  
- Create RBAC (if used) rule to allow acccess to TokenReview API
  ```
  kubectl create -f - <<EOF
  apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: role-tokenreview-binding
    namespace: default
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: system:auth-delegator
  subjects:
  - kind: ServiceAccount
    name: vault-auth
    namespace: default
  EOF
  ```

- Start a pod with confd and get a secret
  ```
  # Create service account for kube auth
  kubectl create serviceaccount vault-auth
  # Start pod
  kubectl run test -i -t --image=quay.io/stepanstipl/test:confd-v7 \
    --overrides='{ "apiVersion": "v1", "spec": {"serviceAccount": "vault-auth", "serviceAccountName": "vault-auth"} }' \
    --restart=Never -- sh
  # Inside the pod
  # Create confd config
  mkdir -p /etc/confd/conf.d /etc/confd/templates
  echo '[template]
  src = "test.conf.tmpl"
  dest = "/tmp/test.conf"
  keys = [
      "/secret/foo",
  ]' > /etc/confd/conf.d/test.toml
  # And template
  echo '{{getv "/secret/foo"}}' > /etc/confd/templates/test.conf.tmpl
  # and finally run confd
  confd -onetime -backend vault -auth-type kubernetes -role confd -node http://vault-vault:8200 -log-level debug
  ```

- Check `/tmp/test.conf`, it should contain your secret
  ```
  cat /tmp/test.conf
  ```
