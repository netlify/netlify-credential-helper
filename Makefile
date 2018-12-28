.PHONY: all build_linux build_macosx build_windows clean clean_release deps package_deb package_linux package_macosx package_rpm package_windows release test

define build
	@echo "Building git-credential-netlify for $(os)/$(arch)"
	@mkdir -p builds/$(os)-${TAG}
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X github.com/netlify/netlify-credential-helper/credentials.Version=${TAG} -X github.com/netlify/netlify-credential-helper/credentials.SHA=`git rev-parse HEAD`" -o builds/$(os)-${TAG}/git-credential-netlify cmd/netlify-credential-helper/main.go
	@echo "Built: builds/$(os)-${TAG}/git-credential-netlify"
endef

define linux_package
	@mkdir -p builds/$(os)-release
	@cp -f builds/$(os)-${TAG}/git-credential-netlify builds/$(os)-release/git-credential-netlify 
	@nfpm -f resources/nfpm.yaml pkg --target releases/${TAG}/$(binary)-$(os)-$(arch).$(1)
endef

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: deps test build ## Run tests and build the binary.

binary = git-credential-netlify
os = linux
arch = amd64
TAG = development

build_linux: override os=linux
build_linux: ## Build the binary for Linux.
	$(call build)

build_macosx: override os=darwin
build_macosx: ## Build the binary for Mac OS X.
	$(call build)

build_windows: override os=windows
build_windows: ## Build the binary for Windows.
	$(call build)

clean: ## Remove all artifacts.
	@rm -rf builds releases

clean_release: ## Remove a release artifact.
	@mkdir -p releases/${TAG}
	@rm -f releases/${TAG}/$(binary)-$(os)-$(arch)-${TAG}.*
	@rm -rf pkg-build

deps: ## Install dependencies.
	@echo "Installing dependencies"
	@GO111MODULE=on go mod verify
	@GO111MODULE=on go mod tidy

package_deb: override os=linux
package_deb: build_linux clean_release ## Build a release package for Debian and Ubuntu.
	$(call linux_package,deb)

package_linux: override os=linux
package_linux: build_linux clean_release ## Build a release package for Linux.
	@tar -czf releases/${TAG}/$(binary)-$(os)-$(arch).tar.gz -C builds/$(os)-${TAG} $(binary)

package_macosx: override os=darwin
package_macosx: build_macosx clean_release ## Build a release package for Mac OS X.
	@tar -czf releases/${TAG}/$(binary)-$(os)-$(arch).tar.gz -C builds/$(os)-${TAG} $(binary)

package_rpm: override os=linux
package_rpm: build_linux clean_release ## Build a release package for Red Hat, Fedora and CentOS.
	$(call linux_package,rpm)

package_windows: override os=windows
package_windows: build_windows clean_release ## Build a release package for Windows.
	@zip -j releases/${TAG}/$(binary)-$(os)-$(arch).zip builds/$(os)-${TAG}/$(binary)

release: release_artifacts ## Create a GitHub release and upload packages.
	@echo "Uploading release"
	@hub release create -a releases/${TAG}/$(binary)-darwin-$(arch).tar.gz -a releases/${TAG}/$(binary)-linux-$(arch).tar.gz -a releases/${TAG}/$(binary)-linux-$(arch).deb -a releases/${TAG}/$(binary)-linux-$(arch).rpm -a releases/${TAG}/$(binary)-windows-$(arch).zip -a releases/${TAG}/checksums.txt v${TAG}

release_artifacts: package_linux package_deb package_rpm package_macosx package_windows release_checksums ## Build all the release artifacts.
	@echo "Release artifacts created in releases/${TAG}"

release_checksums: ## Calculate checksums for release artifacts
	@sha256sum releases/${TAG}/$(binary)-darwin-$(arch).tar.gz >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).tar.gz  >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).deb     >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).rpm     >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-windows-$(arch).zip   >> releases/${TAG}/checksums.txt

test: deps ## Run tests.
	@GO111MODULE=on go test -v ./...
