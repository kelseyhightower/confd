package clconf

import (
	"encoding/base64"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestDecryptPaths(t *testing.T) {
	config, err := NewTestConfig()
	if err != nil {
		t.Errorf("Unable to create config for TestDecryptPaths: %v", err)
	}
	secretAgent, err := NewTestSecretAgent()
	if err != nil {
		t.Errorf("Unable to create secret agent for TestDecryptPaths: %v", err)
	}

	err = secretAgent.DecryptPaths(config, "/foo")
	if err == nil {
		t.Error("DecryptPaths invalid path should have failed")
	}

	err = secretAgent.DecryptPaths(config, "/app/db/port")
	if err == nil {
		t.Error("DecryptPaths not a string should have failed")
	}

	err = secretAgent.DecryptPaths(config, "/app/db/password-plaintext")
	if err == nil {
		t.Error("DecryptPaths not an encrypted value should have failed")
	}

	err = secretAgent.DecryptPaths(config, "/app/db/username", "/app/db/password")
	if err != nil {
		t.Errorf("DecryptPaths failed: %v", err)
	}

	if !ValuesAtPathsAreEqual(config, "/app/db/username", "/app/db/username-plaintext") ||
		!ValuesAtPathsAreEqual(config, "/app/db/password", "/app/db/password-plaintext") {
		t.Errorf("DecryptPaths do not match plaintext paths: %v", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "SECRET"
	secretAgent, err := NewTestSecretAgent()
	if err != nil {
		t.Errorf("Unable to create secret agent: %v", err)
	}
	ciphertext, err := secretAgent.Encrypt(plaintext)
	if err != nil {
		t.Errorf("Unable to encrypt: %v", err)
	}
	decrypted, err := secretAgent.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Unable to decrypt: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Decrypted doesnt match plaintext: %v", decrypted)
	}
}

func TestNewSecretAgent(t *testing.T) {
	expected, err := ioutil.ReadFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("Unable to read key file: %v", err)
	}

	secretAgent, err := NewSecretAgentFromFile(NewTestKeysFile())
	if err != nil {
		t.Errorf("Unable to create secret agent from file: %v", err)
	}
	if !reflect.DeepEqual(expected, secretAgent.key) {
		t.Errorf("Unable to create secret agent from file: %v", err)
	}

	secretAgent, err = NewSecretAgentFromBase64(base64.StdEncoding.EncodeToString(expected))
	if err != nil {
		t.Errorf("Unable to create secret agent from base64: %v", err)
	}
	if !reflect.DeepEqual(expected, secretAgent.key) {
		t.Errorf("Unable to create secret agent from base64: %v", err)
	}
}
