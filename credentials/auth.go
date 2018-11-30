package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

var validAuthPaths = [][]string{
	{".config", "netlify"},
	{".netlify", "config"},
}

func SaveAccessToken(token string) error {
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

func DeleteAccessToken() error {
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

func LoadAccessToken() string {
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

	config := struct {
		AccessToken string `json:"access_token"`
	}{}

	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return ""
	}

	return config.AccessToken
}
