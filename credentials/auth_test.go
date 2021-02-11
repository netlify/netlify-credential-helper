package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
)

func TestLoadAccessToken(t *testing.T) {
	t.Run("with the node config format", testLoadAccessTokenFromNewCli)
	t.Run("with the old config format", testLoadAccessTokenFromOldCli)
	t.Run("with the node multi user config format", testLoadAccessTokenFromMultiUser)
	t.Run("with the node invalid config", testLoadAccessTokenFromInvalidUser)
	t.Run("with shared config dir", testLoadAccessTokenFromConfigDir)
}

func testLoadAccessTokenFromNewCli(t *testing.T) {
	testLoadAccessTokenFromCli(t, "_testdata/netlify-node-cli.json")
}

func testLoadAccessTokenFromOldCli(t *testing.T) {
	testLoadAccessTokenFromCli(t, "_testdata/netlify-old-cli.json")
}

func testLoadAccessTokenFromMultiUser(t *testing.T) {
	testLoadAccessTokenFromCli(t, "_testdata/netlify-node-cli-multi-creds.json")
}

func testLoadAccessTokenFromInvalidUser(t *testing.T) {
	f, err := os.Open("_testdata/netlify-invalid-creds.json")
	check(t, err)
	defer f.Close()

	token, err := loadAccessTokenFromFile(f, "foobar.com", checkFakeHostAccess)
	if err == nil {
		t.Fatal("Expected unauthorized error")
	}

	if err.Error() != "unauthorized" {
		t.Fatalf("Expected unauthorized error, got `%v`", err)
	}

	if token != "" {
		t.Fatalf("expected ``, got `%s`", token)
	}
}

func testLoadAccessTokenFromCli(t *testing.T, path string) {
	f, err := os.Open(path)
	check(t, err)
	defer f.Close()

	token, err := loadAccessTokenFromFile(f, "foobar.com", checkFakeHostAccess)
	check(t, err)
	if token != "verysecret" {
		t.Errorf("expected `verysecret`, got `%s`", token)
	}
}

func testLoadAccessTokenFromConfigDir(t *testing.T) {
	home, err := homedir.Dir()
	check(t, err)

	var configPath string
	switch runtime.GOOS {
	case "windows":
		configPath = filepath.Join(home, "AppData", "Roaming", "netlify", "Config")
	case "darwin":
		configPath = filepath.Join(home, "Library", "Preferences", "netlify")
	default:
		configPath = filepath.Join(home, ".config", "netlify")
	}

	// copy test token to shared config path
	input, err := ioutil.ReadFile(filepath.Join("_testdata", "netlify-node-cli.json"))
	check(t, err)

	err = os.MkdirAll(configPath, 0700)
	check(t, err)

	err = ioutil.WriteFile(filepath.Join(configPath, "config.json"), input, 0600)
	check(t, err)

	token, err := loadAccessTokenFromAuthPaths("foobar.com", checkFakeHostAccess)
	check(t, err)
	if token != "verysecret" {
		t.Errorf("expected `verysecret`, got `%s`", token)
	}
}

func checkFakeHostAccess(host, token string) error {
	if host == "foobar.com" && token == "verysecret" {
		return nil
	}
	return fmt.Errorf("unauthorized")
}

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
