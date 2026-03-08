package auth

import "errors"

// ErrNoToken is returned when no PAT can be found from any source.
var ErrNoToken = errors.New("no Azure DevOps PAT configured; run 'lazado init' to set up authentication")

// Provider supplies an Azure DevOps Personal Access Token.
type Provider interface {
	Token() (string, error)
}

// Chain tries multiple providers in order, returning the first successful token.
type Chain struct {
	providers []Provider
}

// NewChain creates a provider that tries each source in order:
// 1. Environment variable
// 2. System keyring
// 3. Credentials file
func NewChain() *Chain {
	return &Chain{
		providers: []Provider{
			&EnvProvider{},
			&KeyringProvider{},
		},
	}
}

func (c *Chain) Token() (string, error) {
	for _, p := range c.providers {
		token, err := p.Token()
		if err == nil && token != "" {
			return token, nil
		}
	}
	return "", ErrNoToken
}
