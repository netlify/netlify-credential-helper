# Netlify Git's credential helper

Netlify Git's credential helper is a program compatible with [Git Credential Helpers](https://git-scm.com/docs/gitcredentials)
that uses Netlify's API to authenticate a user.

## Installation

1. Download the relase binary specific for your OS from our [Releases]("https://github.com/netlify/netlify-credential-helper/releases").

2. Extract the binary in your PATH.

3. Add the credential definition to you Git config:

```
[[credential "https://play.netlify.com"]]
	helper = netlify
```

### Installation with Netlify Large Media

When you run `netlify addons:large-media`, Netlify CLI's will install and setup this helper for you automatically.

## Usage with Netlify Large Media

When Git requires your authentication token to push large media to your server, it will invoke this binary directly.
If you're not logged in in Netlify, Git will give you the option to login. After this first login, this helper will
store your authentication token for future usage so you don't have to login again.

## Development

Go 1.11 or above is required to make changes in this program.

Use `make deps` to install dependencies, `make test` to run tests, and `make build` to build the binary.

## Release

Use `make release` to build all packages and create a release in GitHub Releases.

## License

[MIT](./LICENSE)
