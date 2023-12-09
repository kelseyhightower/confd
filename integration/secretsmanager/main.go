package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var db map[string]string

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("Authorization") == "" {
		log.Println("Unauthorized request")
		return
	}

	switch t := r.Header.Get("X-Amz-Target"); t {
	case "secretsmanager.CreateSecret":
		defer r.Body.Close()
		var b secretsmanager.CreateSecretInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		log.Printf("DB: Setting key=%s value=%s", *b.Name, *b.SecretString)
		db[*b.Name] = *b.SecretString
		return
	case "secretsmanager.GetSecretValue":
		defer r.Body.Close()
		var b secretsmanager.GetSecretValueInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		log.Printf("DB: Getting key=%s", *b.SecretId)
		GetSecretValueOutput := &secretsmanager.GetSecretValueOutput{
			Name:         aws.String(*b.SecretId),
			SecretString: aws.String(db[*b.SecretId]),
			VersionId:    aws.String("abcd"),
		}
		resp, err := jsonutil.BuildJSON(GetSecretValueOutput)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(w, string(resp))
		return
	case "secretsmanager.ListSecrets":
		defer r.Body.Close()
		var b secretsmanager.ListSecretsInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		log.Printf("DB: Getting All key")

		var secretList []*secretsmanager.SecretListEntry

		for k, _ := range db {
			entry := secretsmanager.SecretListEntry{
				Name: aws.String(k),
			}
			secretList = append(secretList, &entry)
		}
		ListSecretsOutput := &secretsmanager.ListSecretsOutput{
			//NextToken:  aws.String(""),
			SecretList: secretList,
		}
		log.Printf("Entries:%v ", ListSecretsOutput)
		resp, err := jsonutil.BuildJSON(ListSecretsOutput)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(w, string(resp))
		return
	default:
		log.Println(r.Header.Get("X-Amz-Target"))
		return
	}
}

func main() {
	db = make(map[string]string)
	http.HandleFunc("/", handler)
	log.Println("Starting AWS Secrets Manager HTTP mocking server")
	http.ListenAndServe(":8002", nil)
}
