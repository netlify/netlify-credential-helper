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
	loginMessage           = "You need to be authenticated with Netlify to use Netlify LM. Do you want to login now? (yes/no) "
	maxAttempts            = 3
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

	if askForLogin() {
		accessToken, err := Login(h.clientID)
		if err != nil {
			return "", "", err
		}
		return netlifyAccessTokenUser, accessToken, nil
	}

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

func askForLogin() bool {
	return askForLoginWithCancellation(1)
}

func askForLoginWithCancellation(attempts int) bool {
	if attempts >= maxAttempts {
		return false
	}
	var response string

	fmt.Print(loginMessage)
	_, err := fmt.Scanln(&response)

	if err != nil {
		return false
	}

	switch {
	case response[0] == 'y' || response[0] == 'Y':
		return true
	case response[0] == 'n' || response[0] == 'N':
		return false
	default:
		return askForLoginWithCancellation(attempts + 1)
	}
}
