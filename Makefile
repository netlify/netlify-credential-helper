.PHONY: all build_linux build_macosx build_windows clean clean_release deps package_deb package_linux package_macosx package_rpm package_windows release release_artifacts release_checksums release_installers release_upload test

define build
	@echo "Building git-credential-netlify for $(os)/$(arch)"
	@mkdir -p builds/$(os)-${TAG}
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build \
		-ldflags "-X github.com/netlify/netlify-credential-helper/credentials.tag=${TAG} \
		-X github.com/netlify/netlify-credential-helper/credentials.sha=`git rev-parse HEAD` \
		-X github.com/netlify/netlify-credential-helper/credentials.distro=$(os) \
		-X github.com/netlify/netlify-credential-helper/credentials.arch=$(arch)" \
		-o builds/$(os)-${TAG}/git-credential-netlify$(1) cmd/netlify-credential-helper/main.go
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
	$(call build,.exe)

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
	@zip -j releases/${TAG}/$(binary)-$(os)-$(arch).zip builds/$(os)-${TAG}/$(binary).exe

release: release_upload release_installers ## Release a new version of git-credential-netlify. Create artifacts and installers, and upload them.

release_artifacts: package_linux package_deb package_rpm package_macosx package_windows release_checksums ## Build all the release artifacts.
	@echo "Release artifacts created in releases/${TAG}"

release_checksums: ## Calculate checksums for release artifacts.
	@rm -f releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-darwin-$(arch).tar.gz >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).tar.gz  >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).deb     >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-linux-$(arch).rpm     >> releases/${TAG}/checksums.txt
	@sha256sum releases/${TAG}/$(binary)-windows-$(arch).zip   >> releases/${TAG}/checksums.txt

release_installers: ## Release Homebrew and Scoop installers.
	@git submodule update --init
	@sha256sum releases/${TAG}/git-credential-netlify-darwin-amd64.tar.gz | awk '{ print $$1 }' | xargs -I '{}' sed -e 's/{SHA256}/{}/' resources/homebrew-template.rb | sed -e 's/{TAG}/${TAG}/' > installers/homebrew-git-credential-netlify/git-credential-netlify.rb
	@sha256sum releases/${TAG}/git-credential-netlify-windows-amd64.zip | awk '{ print $$1 }' | xargs -I '{}' sed -e 's/{SHA256}/{}/' resources/scoop-template.json | sed -e 's/{TAG}/${TAG}/' > installers/scoop-git-credential-netlify/git-credential-netlify.json
	@cd installers/homebrew-git-credential-netlify/ && git add . && git commit -m "Release Version ${TAG}" && git push origin master
	@cd installers/scoop-git-credential-netlify/ && git add . && git commit -m "Release Version ${TAG}" && git push origin master

release_upload: release_artifacts ## Upload release artifacts to GitHub.
	@echo "Uploading release"
	@hub release create -a releases/${TAG}/$(binary)-darwin-$(arch).tar.gz -a releases/${TAG}/$(binary)-linux-$(arch).tar.gz -a releases/${TAG}/$(binary)-linux-$(arch).deb -a releases/${TAG}/$(binary)-linux-$(arch).rpm -a releases/${TAG}/$(binary)-windows-$(arch).zip -a releases/${TAG}/checksums.txt v${TAG}

test: deps ## Run tests.
	@GO111MODULE=on go test -v ./...
