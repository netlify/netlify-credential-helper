package credentials

import (
	"os"
	"testing"
)

func TestLoadAccessToken(t *testing.T) {
	t.Run("with the node config format", testLoadAccessTokenFormNewCli)
	t.Run("with the old config format", testLoadAccessTokenFormOldCli)
}

func testLoadAccessTokenFormNewCli(t *testing.T) {
	testLoadAccessTokenFromCli(t, "_testdata/netlify-node-cli.json")
}

func testLoadAccessTokenFormOldCli(t *testing.T) {
	testLoadAccessTokenFromCli(t, "_testdata/netlify-old-cli.json")
}

func testLoadAccessTokenFromCli(t *testing.T, path string) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	token := loadAccessTokenFromFile(f, "foobar.com")
	if token != "verysecret" {
		t.Errorf("expected `verysecret`, got `%s`", token)
	}
}
