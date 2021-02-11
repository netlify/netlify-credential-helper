# Netlify Git's credential helper

Netlify Git's credential helper is a program compatible with [Git Credential Helpers](https://git-scm.com/docs/gitcredentials)
that uses Netlify's API to authenticate a user.

## Install

Our preferred way to install this software is by using Netlify's CLI plugins system:

1. Install Netlify CLI if you have not yet: `npm install -g netlify-cli`
2. Install [Netlify's Large Media](https://github.com/netlify/netlify-lm-plugin) plugin: `netlify plugins:install netlify-lm-plugin`
3. Run the LM Setup: `netlify lm:setup`.

Netlify's Large Media plugin will download the latest version of this sofware 
for your OS, and configure Git to use it when it's necessary. You don't need to
do anything else.

Alternatively, you can also install this credentials helper manually following one of the guides below:

- [Install on Debian/Ubuntu](#install-on-debianubuntu)
- [Install on Fedora/RedHat](#install-on-fedoraredhat)
- [Install on MacOS X](#install-on-macos-x-with-homebrew)
- [Install on Windows with Powershell](#install-on-windows-with-powershell)
- [Install on Windows with Scoop](#install-on-windows-with-scoop)
- [Manual install](#manual-install)

After manually installing the helper, you'll need to add the credential definition to you Git config:

```
[credential]
	helper = netlify
```

### Install on Debian/Ubuntu

1. Download the deb file from our [Releases]("https://github.com/netlify/netlify-credential-helper/releases").

2. Install with dpkg:

```
sudo dpkg -i git-credential-netlify-linux-amd64.deb
```

### Install on Fedora/RedHat

1. Download the rpm file from our [Releases]("https://github.com/netlify/netlify-credential-helper/releases").

2. Install with dpkg:

```
sudo dnf install git-credential-netlify-linux-amd64.rpm
```

### Install on MacOS X with Homebrew

1. Open a terminal and copy these two commands:

```
brew tap netlify/git-credential-netlify
brew install git-credential-netlify
```

### Install on Windows with Powershell

1. Start a Powershell session and copy these two commands:

```
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
iex (iwr -UseBasicParsing -Uri https://github.com/netlify/netlify-credential-helper/raw/master/resources/install.ps1)
```

### Install on Windows with Scoop

1. Start a Powershell session and copy these two commands:

```
scoop bucket add netlifyctl https://github.com/netlify/scoop-git-credential-netlify
scoop install git-credential-netlify
```

### Manual install

1. Download the relase binary specific for your OS from our [Releases]("https://github.com/netlify/netlify-credential-helper/releases").

2. Extract the binary in your PATH.

## Usage with Netlify Large Media

When Git requires your authentication token to push large media to your server, it will invoke this binary directly.
If you're not logged in in Netlify, Git will give you the option to login. After this first login, this helper will
store your authentication token for future usage so you don't have to login again.

## Development

Go 1.11 or above is required to make changes in this program.

Use `make deps` to install dependencies, `make test` to run tests, and `make build` to build the binary.

## Release

1. Install `nfpm` [version v1.3.1](https://github.com/goreleaser/nfpm/releases/tag/v1.3.1)
2. Install `hub` https://hub.github.com/
3. Create a [GitHub personal access token](https://github.com/settings/tokens/new) and add it to your shell (e.g. `export GITHUB_TOKEN=<token>`)
4. Use `make release TAG=0.1.X` to build all packages and create a release in GitHub Releases.

## License

[MIT](./LICENSE)
