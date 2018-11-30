package main

import (
	"github.com/docker/docker-credential-helpers/credentials"
	netlify "github.com/netlify/netlify-credential-helper/credentials"
)

func main() {
	credentials.Serve(netlify.NewNetlifyCredentials())
}
