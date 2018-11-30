.PHONY: all build deps image release test

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: deps test build ## Run tests and build the binary.

binary = git-credential-netlify
os = linux
arch = amd64
TAG = development

build:
	@echo "Building git-credential-netlify for $(os)/$(arch)"
	@mkdir -p builds/$(os)-${TAG}
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X github.com/netlify/netlify-credential-helper/credentials.Version=${TAG} -X github.com/netlify/netlify-credential-helper/credentials.SHA=`git rev-parse HEAD`" -o builds/$(os)-${TAG}/git-credential-netlify cmd/netlify-credential-helper/main.go
	@echo "Built: builds/$(os)-${TAG}/git-credential-netlify"

build_linux: override os=linux ## Build the binary for Linux hosts.
build_linux: build

build_windows: override os=windows ## Build the binary for Windows hosts.
build_windows: build

clean: ## Remove all artifacts.
	@rm -rf builds releases

clean_release: ## Remove a release artifact.
	@mkdir -p releases/${TAG}
	@rm -f releases/${TAG}/$(binary)-$(os)-$(arch)-${TAG}.tar.gz

deps: ## Install dependencies.
	@echo "Installing dependencies"
	@GO111MODULE=on go mod verify

package: build clean_release ## Build a release package with the default flags.
	@tar -czf releases/${TAG}/$(binary)-$(os)-$(arch)-${TAG}.tar.gz -C builds/$(os)-${TAG} $(binary)

package_linux: override os=linux ## Build a release package for Linux.
package_linux: package

package_macosx: override os=darwin ## Build a release package for Mac OS X.
package_macosx: package

package_windows: override os=darwin ## Build a release package for Windows.
package_windows: build clean_release
	@zip -j releases/${TAG}/$(binary)-$(os)-$(arch)-${TAG}.zip builds/$(os)-${TAG}/$(binary)

release: package_linux package_macosx package_windows ## Create a GitHub release and upload packages.
	@echo "Creating release"
	@hub release create -a releases/${TAG}/$(binary)-darwin-$(arch)-${TAG}.tar.gz -a releases/${TAG}/$(binary)-linux-$(arch)-${TAG}.tar.gz -a releases/${TAG}/$(binary)-windows-$(arch)-${TAG}.zip v${TAG}

test: deps ## Run tests.
	@go test -v ./...
