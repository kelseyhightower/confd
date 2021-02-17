# Vault Backend

The vault backend enables `confd` to pull configuration parameters from Hashicorp Vault

## Configuration

### Authentication



### Environment Variables

Environment variables can be used to provide the required configurations to
`confd`. They will override configurations set in the config and credentials
files.

```
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-2
```

### Config and Credentials Files

AWS credentials and configuration can be stored in the standard AWS CLI config
files. These may be set up manually or via `aws configure`

\~/.aws/credentials

```
[default]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

\~/.aws/config

```
[default]
region=us-east-2
```

### IAM Role for EC2

An IAM role can be used to grant `confd` permissions to SSM. When used you will
not need to set `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY`. When `confd` is
executed on an EC2 instance it will acquire the AWS Region setting from EC2
Metadata.

Setup of IAM roles for EC2 instances is well documented in the AWS User Guides.


## Options



## Basic Example



## Advanced Example