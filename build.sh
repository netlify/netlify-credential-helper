#!/bin/bash

set -e

binary="git-credential-netlify"

# cannot loop over a list of os and arches as 
# dawrin has no arm and 386 architecture
compile_targets=(
  windows/386 
  windows/amd64 
  windows/arm 
  windows/arm64
  darwin/amd64
  darwin/arm64
  linux/386
  linux/amd64
  linux/arm
  linux/arm64
)

# linux_package_formats=(
#   rpm
#   deb
# )


CMD=$1
TAG=$2
TAG="${TAG:=development}" # second parameter of script is tags
shasum_file="releases/${TAG}/checksums.txt"

if [ -z "${CMD}" ]; then
  echo "Need to provide a CMD as first parameter to the script like 'build', 'package' or 'publish'"
  exit 1
fi


if [ "${CMD}" == "package" ];then
  # cleanup on package
  rm -rf "releases/${TAG}"
  rm -rf pkg-build
  mkdir -p "releases/${TAG}"
  touch "${shasum_file}"

# Run the actual publishing command
elif [ "$CMD" == "publish" ]; then
  args=""

  for artifact in "releases/${TAG}"/*; do
    echo " -a ${artifact}"
    args="${args} -a ${artifact}"
  done
  
  hub release create "${args}" "v${TAG}"
  exit 0
fi

# $1 OS
# $2 ARCH
function os_task {
  os=$1
  arch=$2
  ext=""

  if [ "${os}" == "windows" ]; then
    ext=".exe"
  fi

  folder="builds/${os}-${arch}-${TAG}"
  file="${folder}/${binary}${ext}"

  
  # run the build command
  if [ "$CMD" == "build" ];then
    echo "Building ${binary} for ${os}/${arch}"
    mkdir -p "${folder}"

    GOOS=${os} GOARCH=${arch} go build \
    -ldflags "-X github.com/netlify/netlify-credential-helper/credentials.tag=${TAG} \
    -X github.com/netlify/netlify-credential-helper/credentials.sha=$(git rev-parse HEAD) \
    -X github.com/netlify/netlify-credential-helper/credentials.distro=${os} \
    -X github.com/netlify/netlify-credential-helper/credentials.arch=${arch}" \
    -o "${file}" cmd/netlify-credential-helper/main.go

    echo "Built: ${file}"

  # Publishing goes here
  elif [ "$CMD" == "package" ]; then
    if [ "${os}" == "windows" ]; then
      dist="releases/${TAG}/${binary}-${os}-${arch}.zip"
    	zip -j "${dist}" "${file}"
      # append shasum to file
      sha256sum "${dist}" >> "releases/${TAG}/checksums.txt"
    elif [ "${os}" == "linux" ]; then
      dist="releases/${TAG}/${binary}-${os}-${arch}.tar.gz"
      tar -czf "${dist}" -C "builds/${os}-${arch}-${TAG}" "${binary}"
      # append shasum to file
      sha256sum "${dist}" >> "releases/${TAG}/checksums.txt"

      # Maybe we can skip the .deb and .rpm packages

      # build additional linux package formats
      # dist_folder="builds/${os}-${arch}-release"
	    # mkdir -p "builds/${os}-${arch}-release"

      # for format in "${linux_package_formats[@]}"; do
      #   dist="${dist_folder}/${binary}.${format}"

      #   cp -f "./${file}" "${dist}"
      #   nfpm -f resources/nfpm.yaml pkg --target "releases/${TAG}/${binary}-${os}-${arch}.${format}"
        
      #   sha256sum "${dist}" >> "releases/${TAG}/checksums.txt"
      # done
    else 	
      # for everything else use a tar file
      dist="releases/${TAG}/${binary}-${os}-${arch}.tar.gz"
      tar -czf "${dist}" -C "builds/${os}-${arch}-${TAG}" "${binary}"
      # append shasum to file
      sha256sum "${dist}" >> "releases/${TAG}/checksums.txt"
    fi
  fi
}


for target in "${compile_targets[@]}"; do
    os=${target%%"/"*}
    arch=${target#*"/"}

    os_task "${os}" "${arch}"
done
