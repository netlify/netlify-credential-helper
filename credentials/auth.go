package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/adrg/xdg"
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

func getValidAuthPaths() ([][]string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	var configPaths []string
	if goruntime.GOOS == "windows" {
		// on windows the go xdg lib behaves a bit different than the Node.js one
		// https://github.com/sindresorhus/env-paths/blob/de22adb240117c7aacf0365187472907a9e06872/index.js#L28
		// https://github.com/adrg/xdg/blob/af0f1bbdcb2e9415a67e57c4125afc9daeb3ca17/paths_windows.go#L40
		appDataDir := os.Getenv("APPDATA")
		if appDataDir == "" {
			appDataDir = filepath.Join(home, "AppData", "Roaming")
		}
		configPaths = []string{appDataDir, "netlify", "Config", "config.json"}
	} else {
		configPaths = []string{xdg.ConfigHome, "netlify", "config.json"}
	}

	validAuthPaths := [][]string{
		configPaths,
		{home, ".netlify", "config.json"},
		{home, ".config", "netlify"},
		{home, ".netlify", "config"},
		{home, ".config", "netlify.json"},
	}

	return validAuthPaths, nil
}

func saveAccessToken(token string) error {
	validAuthPaths, err := getValidAuthPaths()
	if err != nil {
		return err
	}

	configPath := filepath.Join(validAuthPaths[0]...)

	// make sure the directory structure exists
	err = os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, 0600)
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

func loadAccessToken(host string) (string, error) {
	accessToken := os.Getenv(netlifyEnvAccessToken)
	if accessToken != "" {
		if err := tryAccessToken(host, accessToken); err != nil {
			return "", err
		}
		return accessToken, nil
	}

	return loadAccessTokenFromAuthPaths(host, tryAccessToken)
}

func loadAccessTokenFromAuthPaths(host string, checkHostAccess netlifyHostAccessCheck) (string, error) {
	validAuthPaths, err := getValidAuthPaths()
	if err != nil {
		return "", err
	}
	var f *os.File
	for _, p := range validAuthPaths {
		f, err = os.Open(filepath.Join(p...))
		if err == nil {
			break
		}
	}

	if err != nil || f == nil {
		return "", err
	}
	defer f.Close()

	return loadAccessTokenFromFile(f, host, checkHostAccess)
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
	client, ctx := newNetlifyAPIClient(credentials)
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

func newNetlifyAPIClient(credentials func(r runtime.ClientRequest, _ strfmt.Registry) error) (*porcelain.Netlify, context.Context) {
	transport := client.New(netlifyAPIHost, netlifyAPIPath, apiSchemes)
	client := porcelain.New(transport, strfmt.Default)

	creds := runtime.ClientAuthInfoWriterFunc(credentials)
	ctx := apiContext.WithAuthInfo(context.Background(), creds)
	return client, ctx
}
