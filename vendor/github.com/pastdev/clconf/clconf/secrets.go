package clconf

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/xordataexchange/crypt/encoding/secconf"
)

// SecretAgent loads and holds a keypair needed for
// encryption/decryption
type SecretAgent struct {
	key []byte
}

// Decrypt will return the decrypted value represented by encrypted
func (secretAgent *SecretAgent) Decrypt(encrypted string) (string, error) {
	if secretAgent.key == nil {
		return "", errors.New("SecretAgent missing key")
	}
	b, err := secconf.Decode(
		[]byte(encrypted),
		bytes.NewBuffer(secretAgent.key))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// DecryptPaths will will replace the values at the indicated paths with thier
// decrypted values
func (secretAgent *SecretAgent) DecryptPaths(config interface{}, encryptedPaths ...string) error {
	for _, encryptedPath := range encryptedPaths {
		value, ok := GetValue(config, encryptedPath)
		if !ok {
			return fmt.Errorf("%v not found", encryptedPath)
		}
		stringValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v not a string", encryptedPath)
		}
		decrypted, err := secretAgent.Decrypt(stringValue)
		if err != nil {
			return err
		}
		SetValue(config, encryptedPath, decrypted)
	}
	return nil
}

// Encrypt will return the encrypted value represented by decrypted
func (secretAgent *SecretAgent) Encrypt(decrypted string) (string, error) {
	if secretAgent.key == nil {
		return "", errors.New("SecretAgent missing key")
	}
	b, err := secconf.Encode(
		[]byte(decrypted),
		bytes.NewBuffer(secretAgent.key))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func newSecretAgent(key []byte, err error) (*SecretAgent, error) {
	if err != nil {
		return nil, err
	}
	return NewSecretAgent(key), nil
}

// NewSecretAgent will return a new SecretAgent with the provided
// key.
func NewSecretAgent(key []byte) *SecretAgent {
	return &SecretAgent{
		key: key,
	}
}

// NewSecretAgentFromFile loads from keyFile
func NewSecretAgentFromFile(keyFile string) (*SecretAgent, error) {
	return newSecretAgent(ioutil.ReadFile(keyFile))
}

// NewSecretAgentFromBase64 loads from keyBase64
func NewSecretAgentFromBase64(keyBase64 string) (*SecretAgent, error) {
	return newSecretAgent(base64.StdEncoding.DecodeString(keyBase64))
}
