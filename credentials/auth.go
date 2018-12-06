package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/netlify/open-api/go/plumbing/operations"
	"github.com/netlify/open-api/go/porcelain"
	apiContext "github.com/netlify/open-api/go/porcelain/context"
)

type netlifyHostAccessCheck func(host, token string) error

type netlifyAuthInfo struct {
	AccessToken string `json:"token"`
}

type netlifyUserInfo struct {
	Auth netlifyAuthInfo `json:"auth"`
}

type netlifyConfig struct {
	AccessToken string                     `json:"access_token,omitempty"`
	UserID      string                     `json:"userId,omitempty"`
	Users       map[string]netlifyUserInfo `json:"users,omitempty"`
}

var validAuthPaths = [][]string{
	{".config", "netlify"},
	{".netlify", "config"},
	{".config", "netlify.json"},
}

func saveAccessToken(token string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	args := append([]string{home}, validAuthPaths[0]...)
	f, err := os.OpenFile(filepath.Join(args...), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	config := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: token,
	}

	return json.NewEncoder(f).Encode(&config)
}

func deleteAccessToken() error {
	home, err := homedir.Dir()
	if err != nil {
		return nil
	}

	for _, p := range validAuthPaths {
		args := append([]string{home}, p...)
		os.Remove(filepath.Join(args...))
	}

	return nil
}

func loadAccessToken(host string) (string, error) {
	accessToken := os.Getenv(netlifyEnvAccessToken)
	if accessToken != "" {
		if err := tryAccessToken(host, accessToken); err != nil {
			return "", err
		}
		return accessToken, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	var f *os.File
	for _, p := range validAuthPaths {
		args := append([]string{home}, p...)
		f, err = os.Open(filepath.Join(args...))
		if err == nil {
			break
		}
	}

	if err != nil || f == nil {
		return "", err
	}
	defer f.Close()

	return loadAccessTokenFromFile(f, host, tryAccessToken)
}

func loadAccessTokenFromFile(f *os.File, host string, checkHostAccess netlifyHostAccessCheck) (string, error) {
	config := netlifyConfig{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return "", err
	}

	if config.AccessToken != "" {
		if err := checkHostAccess(host, config.AccessToken); err != nil {
			return "", err
		}
		return config.AccessToken, nil
	}

	if len(config.Users) == 0 {
		return "", nil
	}

	var lastError error
	for _, user := range config.Users {
		err := checkHostAccess(host, user.Auth.AccessToken)
		if err == nil {
			return user.Auth.AccessToken, nil
		}
		lastError = err
	}

	return "", lastError
}

func tryAccessToken(host, token string) error {
	credentials := func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "git-credential-netlify")
		r.SetHeaderParam("Authorization", "Bearer "+token)
		return nil
	}
	client, ctx := newNetlifyApiClient(credentials)
	site, err := client.GetSite(ctx, host)
	if err != nil {
		if apiErr, ok := err.(*operations.GetSiteDefault); ok && apiErr.Payload.Code == 404 {
			return fmt.Errorf("Unknown Netlify site: `%s`", host)
		}
		return err
	}

	if site == nil || len(site.Capabilities) == 0 {
		return fmt.Errorf("Unknown Netlify site: `%s`", host)
	}

	enabled, ok := site.Capabilities[netlifyLargeMediaCapability]
	if !ok {
		return fmt.Errorf("Netlify Large Media is not enabled for this site")
	}
	if e, ok := enabled.(bool); !ok || !e {
		return fmt.Errorf("Netlify Large Media is not enabled for this site")
	}

	return nil
}

func newNetlifyApiClient(credentials func(r runtime.ClientRequest, _ strfmt.Registry) error) (*porcelain.Netlify, context.Context) {
	transport := client.New(netlifyApiHost, netlifyApiPath, apiSchemes)
	client := porcelain.New(transport, strfmt.Default)

	creds := runtime.ClientAuthInfoWriterFunc(credentials)
	ctx := apiContext.WithAuthInfo(context.Background(), creds)
	return client, ctx
}
