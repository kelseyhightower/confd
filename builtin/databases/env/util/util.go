package util

import "strings"

func Transform(key string) string {
	replacer := strings.NewReplacer("/", "_")
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}

func Clean(key string) string {
	cleanReplacer := strings.NewReplacer("_", "/")
	newKey := "/" + key
	return cleanReplacer.Replace(strings.ToLower(newKey))
}
