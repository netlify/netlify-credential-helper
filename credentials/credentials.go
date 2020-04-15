package credentials

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	netlifyEnvAccessToken       = "NETLIFY_ACCESS_TOKEN"
	netlifyEnvClientID          = "NETLIFY_CLIENT_ID"
	netlifyServerName           = "https://api.netlify.com"
	netlifyAccessTokenUser      = "access-token"
	netlifyDefaultClientID      = "5edad8f69d47ae8923d0cf0b4ab95ba1415e67492b5af26ad97f4709160bb31b"
	netlifyAPIPath              = "/api/v1"
	netlifyLargeMediaCapability = "large_media_enabled"
	netlifyHost                 = ".netlify.app"
	netlifyAltHost              = ".netlify.com"

	gitHostKey     = "host"
	gitUsernameKey = "username"
	gitPasswordKey = "password"
	gitPathKey     = "path"
)

var (
	tag    = "static-binary-tag"
	sha    = "static-binary-sha"
	distro = "static-binary-distro"
	arch   = "static-binary-arch"
)

// HandleCommand checks arguments and inits logger.
func HandleCommand() {
	initLogger()

	logrus.WithFields(logrus.Fields{
		"args": os.Args,
	}).Debug("Initializing Netlify credential helper")

	var err error
	if len(os.Args) != 2 {
		err = fmt.Errorf("Usage: %s <get|version>", os.Args[0])
	}

	if err == nil {
		err = handleCommand(os.Args[1], os.Stdin, os.Stdout)
	}

	if err != nil {
		logrus.WithError(err).Error("Aborting Netlify credential helper execution")
		os.Exit(1)
	}
}

// handleCommand uses a helper and a key to run a credential action.
func handleCommand(key string, in io.Reader, out io.Writer) error {
	switch key {
	case "get":
		return getCredentials(in, out)
	case "version":
		return printVersion(out)
	case "--version":
		return printVersion(out)
	}
	return nil
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

	fields := logrus.Fields{}
	for key, value := range data {
		fields[key] = value
	}

	logrus.WithFields(fields).Debug("Git input received")

	host, exist := data[gitHostKey]
	if !exist {
		return fmt.Errorf("Missing host to check credentials: %v", data)
	}

	if !(strings.HasSuffix(host, netlifyHost) || strings.HasSuffix(host, netlifyAltHost)) {
		// ignore hosts that are not *.netlify.app or *.netlify.com
		return nil
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

	fields = logrus.Fields{}
	for key, value := range data {
		if key == "password" {
			value = value[0:6] + "****************"
		}
		fields[key] = value
	}
	logrus.WithFields(fields).Debug("Writing output data")

	fmt.Fprint(writer, buffer.String())
	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "git-credential-netlify/%s (Netlify; %s %s; %s; git %s)\n", tag, distro, arch, runtime.Version(), sha[0:8])
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

func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	if os.Getenv("GIT_TRACE") != "" || os.Getenv("DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}
}
