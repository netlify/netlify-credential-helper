.PHONY: all clean deps release release_artifacts release_installers release_upload test

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: deps test build ## Run tests and build the binary.

binary = git-credential-netlify
TAG = development

clean: ## Remove all artifacts.
	@rm -rf builds releases

deps: ## Install dependencies.
	@echo "Installing dependencies"
	@GO111MODULE=on go mod verify
	@GO111MODULE=on go mod tidy

test: deps ## Run tests.
	@GO111MODULE=on go test -v ./...

build:
	./build.sh build ${TAG}

release_artifacts: build ## Build release artifacts with checksums
	./build.sh package ${TAG}	
	@echo "Release artifacts created in releases/${TAG}"

release: release_upload release_installers ## Release a new version of git-credential-netlify. Create artifacts and installers, and upload them.

release_installers: ## Release Homebrew and Scoop installers.
	@git submodule update --init
	@sha256sum releases/${TAG}/git-credential-netlify-darwin-amd64.tar.gz | awk '{ print $$1 }' | xargs -I '{}' sed -e 's/{SHA256}/{}/' resources/homebrew-template.rb | sed -e 's/{TAG}/${TAG}/' > installers/homebrew-git-credential-netlify/git-credential-netlify.rb
	@sha256sum releases/${TAG}/git-credential-netlify-windows-amd64.zip | awk '{ print $$1 }' | xargs -I '{}' sed -e 's/{SHA256}/{}/' resources/scoop-template.json | sed -e 's/{TAG}/${TAG}/' > installers/scoop-git-credential-netlify/git-credential-netlify.json
	@cd installers/homebrew-git-credential-netlify/ && git add . && git commit -m "Release Version ${TAG}" && git push origin master
	@cd installers/scoop-git-credential-netlify/ && git add . && git commit -m "Release Version ${TAG}" && git push origin master
	@git checkout -b release_${TAG} && git add . && git commit -m "Update installer submodules for release ${TAG}" && git push -u origin release_${TAG}
	@hub pull-request -m "Update installer submodules for release ${TAG}"

release_upload: release_artifacts ## Upload release artifacts to GitHub.
	@echo "Uploading release"
	./build.sh build ${TAG}