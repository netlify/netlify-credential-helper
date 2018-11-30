package credentials

import (
	"fmt"
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
)

const (
	netlifyEnvAccessToken  = "NETLIFY_ACCESS_TOKEN"
	netlifyEnvClientID     = "NETLIFY_CLIENT_ID"
	netlifyServerName      = "https://api.netlify.com"
	netlifyAccessTokenUser = "access-token"
	netlifyDefaultClientID = "5edad8f69d47ae8923d0cf0b4ab95ba1415e67492b5af26ad97f4709160bb31b"
)

var (
	Version = "static-binary-version"
	SHA     = "static-binary-sha"
)

type NetlifyCredentials struct {
	accessToken string
	clientID    string
}

func NewNetlifyCredentials() NetlifyCredentials {
	accessToken := os.Getenv(netlifyEnvAccessToken)
	clientID := os.Getenv(netlifyEnvClientID)
	if clientID == "" {
		clientID = netlifyDefaultClientID
	}
	return NetlifyCredentials{accessToken, clientID}
}

func (h NetlifyCredentials) Add(creds *credentials.Credentials) error {
	return SaveAccessToken(creds.Secret)
}

func (h NetlifyCredentials) Delete(serverURL string) error {
	return DeleteAccessToken()
}

func (h NetlifyCredentials) Get(serverURL string) (string, string, error) {
	accessToken := LoadAccessToken()
	if accessToken != "" {
		return netlifyAccessTokenUser, accessToken, nil
	}

	accessToken, err := Login(h.clientID)
	if err != nil {
		return "", "", err
	}
	return netlifyAccessTokenUser, accessToken, nil

	return "", "", nil
}

func (h NetlifyCredentials) List() (map[string]string, error) {
	return map[string]string{
		netlifyServerName: netlifyAccessTokenUser,
	}, nil
}

func (h NetlifyCredentials) Version() string {
	return fmt.Sprintf("Version: %s\nGit SHA: %s", Version, SHA)
}
