package credentials

import (
	"fmt"
	"os"
	"testing"
)

func TestLoadAccessToken(t *testing.T) {
	t.Run("with the node config format", testLoadAccessTokenFromNewCli)
	t.Run("with the old config format", testLoadAccessTokenFromOldCli)
	t.Run("with the node multi user config format", testLoadAccessTokenFromMultiUser)
	t.Run("with the node invalid config", testLoadAccessTokenFromInvalidUser)
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
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	token, err := loadAccessTokenFromFile(f, "foobar.com", checkFakeHostAccess)
	if err != nil {
		t.Fatal(err)
	}
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
