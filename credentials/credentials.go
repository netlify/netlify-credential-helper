package credentials

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	netlifyEnvAccessToken  = "NETLIFY_ACCESS_TOKEN"
	netlifyEnvClientID     = "NETLIFY_CLIENT_ID"
	netlifyServerName      = "https://api.netlify.com"
	netlifyAccessTokenUser = "access-token"
	netlifyDefaultClientID = "5edad8f69d47ae8923d0cf0b4ab95ba1415e67492b5af26ad97f4709160bb31b"
	netlifyApiPath         = "/api/v1"
	netlifyLfsPath         = "/.netlify/lfs"

	gitHostKey     = "host"
	gitUsernameKey = "username"
	gitPasswordKey = "password"
	gitPathKey     = "path"
)

var (
	Version = "static-binary-version"
	SHA     = "static-binary-sha"
)

func HandleCommand() {
	var err error
	if len(os.Args) != 2 {
		err = fmt.Errorf("Usage: %s <store|get|erase|version>", os.Args[0])
	}

	if err == nil {
		err = handleCommand(os.Args[1], os.Stdin, os.Stdout)
	}

	if err != nil {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		os.Exit(1)
	}
}

// handleCommand uses a helper and a key to run a credential action.
func handleCommand(key string, in io.Reader, out io.Writer) error {
	switch key {
	case "store":
		return nil // this command is not supported, so we can ignore it
	case "get":
		return getCredentials(in, out)
	case "erase":
		return deleteAccessToken()
	case "version":
		return printVersion(out)
	}
	return fmt.Errorf("Unknown credential action `%s`", key)
}

// getCredentials retrieves the credentials for a given server url.
// The reader must contain the server URL to search.
// The writer is used to write the text serialization of the credentials.
func getCredentials(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)

	data := map[string]string{}
	buffer := new(bytes.Buffer)
	for scanner.Scan() {
		keyAndValue := bytes.SplitN(scanner.Bytes(), []byte("="), 2)
		if len(keyAndValue) > 1 {
			data[string(keyAndValue[0])] = string(keyAndValue[1])
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	host, exist := data[gitHostKey]
	if !exist {
		return fmt.Errorf("Missing host to check credentials: %s", buffer.String())
	}

	if path, exist := data[gitPathKey]; !exist || path != netlifyLfsPath {
		return fmt.Errorf("Invalid LFS path: %s", buffer.String())
	}

	accessToken, err := getAccessToken(host)
	if err != nil {
		return err
	}
	data[gitUsernameKey] = netlifyAccessTokenUser
	data[gitPasswordKey] = accessToken

	buffer.Reset()

	for key, value := range data {
		fmt.Fprintf(buffer, "%s=%s\n", key, value)
	}

	fmt.Fprint(writer, buffer.String())
	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "Version: %s\nGit SHA: %s\n", Version, SHA)
	return err
}

func getAccessToken(host string) (string, error) {
	accessToken, err := loadAccessToken(host)
	if err != nil {
		return "", err
	}

	if accessToken != "" {
		return accessToken, nil
	}

	clientID := os.Getenv(netlifyEnvClientID)
	if clientID == "" {
		clientID = netlifyDefaultClientID
	}

	accessToken, err = login(clientID, host)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
