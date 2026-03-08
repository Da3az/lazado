package auth

import "github.com/zalando/go-keyring"

const (
	serviceName = "lazado"
	accountName = "azure-devops-pat"
)

// KeyringProvider stores and retrieves the PAT from the system keyring.
type KeyringProvider struct{}

func (k *KeyringProvider) Token() (string, error) {
	token, err := keyring.Get(serviceName, accountName)
	if err != nil {
		return "", ErrNoToken
	}
	return token, nil
}

// Store saves a PAT to the system keyring.
func Store(token string) error {
	return keyring.Set(serviceName, accountName, token)
}

// Delete removes the PAT from the system keyring.
func Delete() error {
	return keyring.Delete(serviceName, accountName)
}
