package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var db map[string]string

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("Authorization") == "" {
		log.Println("Unauthorized request")
		return
	}
	switch t := r.Header.Get("X-Amz-Target"); t {
	case "AmazonSSM.PutParameter":
		defer r.Body.Close()
		var b ssm.PutParameterInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		log.Printf("DB: Setting key=%s value=%s", *b.Name, *b.Value)
		db[*b.Name] = *b.Value
		return
	case "AmazonSSM.GetParameters":
		defer r.Body.Close()
		var b ssm.GetParametersInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		var getParametersOutput ssm.GetParametersOutput
		var parameters []*ssm.Parameter
		for _, name := range b.Names {
			log.Printf("DB: Getting key=%s value=%s", *name, db[*name])
			parameters = append(parameters, &ssm.Parameter{
				Name:  aws.String(*name),
				Type:  aws.String("String"),
				Value: aws.String(db[*name]),
			})
		}
		getParametersOutput.SetParameters(parameters)
		resp, err := jsonutil.BuildJSON(getParametersOutput)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(w, string(resp))
		return
	case "AmazonSSM.DescribeParameters":
		defer r.Body.Close()
		var b ssm.DescribeParametersInput
		err := jsonutil.UnmarshalJSON(&b, r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("Body=%#v\n", b)
		var describeParametersOutput ssm.DescribeParametersOutput
		var parameters []*ssm.ParameterMetadata
		filter := b.Filters[0].Values[0]
		for k := range db {
			if strings.HasPrefix(k, *filter) == false {
				continue
			}
			log.Printf("DB: Getting key=%s", k)
			parameters = append(parameters, &ssm.ParameterMetadata{
				Name: aws.String(k),
				Type: aws.String("String"),
			})
		}
		describeParametersOutput.SetParameters(parameters)
		resp, err := jsonutil.BuildJSON(describeParametersOutput)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(w, string(resp))
		return
	default:
		log.Println("Unknown target")
		return
	}
	return
}

func main() {
	db = make(map[string]string)
	http.HandleFunc("/", handler)
	log.Println("Starting AWS SSM HTTP mocking server")
	http.ListenAndServe(":8001", nil)
}
