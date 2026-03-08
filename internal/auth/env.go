package auth

import "os"

// EnvProvider reads the PAT from the LAZADO_PAT environment variable.
type EnvProvider struct{}

func (e *EnvProvider) Token() (string, error) {
	token := os.Getenv("LAZADO_PAT")
	if token == "" {
		return "", ErrNoToken
	}
	return token, nil
}
