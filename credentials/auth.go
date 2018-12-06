package credentials

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/netlify/open-api/go/porcelain"
	apiContext "github.com/netlify/open-api/go/porcelain/context"
)

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

func loadAccessToken(host string) string {
	home, err := homedir.Dir()
	if err != nil {
		return ""
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
		return ""
	}
	defer f.Close()

	return loadAccessTokenFromFile(f, host)
}

func loadAccessTokenFromFile(f *os.File, host string) string {
	config := netlifyConfig{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return ""
	}

	if config.AccessToken != "" {
		return config.AccessToken
	}

	if len(config.Users) == 0 {
		return ""
	}

	if len(config.Users) == 1 {
		// return the first token but range over the map
		// because Go doesn't have a way to give you the
		// first element
		for _, user := range config.Users {
			return user.Auth.AccessToken
		}
	}

	for _, user := range config.Users {
		if err := tryAccessToken(host, user.Auth.AccessToken); err == nil {
			return user.Auth.AccessToken
		}
	}

	return ""
}

func tryAccessToken(host, token string) error {
	transport := client.New(netlifyApiHost, "/api/v1", apiSchemes)
	client := porcelain.New(transport, strfmt.Default)

	credentials := func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "git-credential-netlify")
		r.SetHeaderParam("Authorization", "Bearer "+token)
		return nil
	}

	creds := runtime.ClientAuthInfoWriterFunc(credentials)
	ctx := apiContext.WithAuthInfo(context.Background(), creds)

	_, err := client.GetSite(ctx, host)
	return err
}
