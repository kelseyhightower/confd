package aac

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"time"
)

type connectionString struct {
	Endpoint string
	ID       string
	Secret   string
}

type kvItem struct {
	Etag         string            `json:"etag" yaml:"etag"`
	Key          string            `json:"key" yaml:"key"`
	Label        string            `json:"label" yaml:"label"`
	ContentType  string            `json:"content_type" yaml:"content_type"`
	Value        string            `json:"value" yaml:"value"`
	Tags         map[string]string `json:"tags" yaml:"tags"`
	Locked       bool              `json:"locked" yaml:"locked"`
	LastModified time.Time         `json:"last_modified" yaml:"last_modified"`
}

func (c connectionString) getSecret() []byte {
	val, _ := base64.StdEncoding.DecodeString(c.Secret)
	return val
}

func parseConnectionString(connStr string) (connectionString, error) {
	re := regexp.MustCompile("Endpoint=([^;]+);Id=([^;]+);Secret=([^;]+)")
	match := re.FindStringSubmatch(connStr)

	if len(match) != 4 {
		return connectionString{}, fmt.Errorf("Invalid connection string")
	}

	return connectionString{Endpoint: match[1], ID: match[2], Secret: match[3]}, nil
}
