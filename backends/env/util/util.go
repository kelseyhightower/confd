package util

import "strings"

var replacer = strings.NewReplacer("/", "_")

func Transform(key string) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}

var cleanReplacer = strings.NewReplacer("_", "/")

func Clean(key string) string {
	newKey := "/" + key
	return cleanReplacer.Replace(strings.ToLower(newKey))
}
