BUILD_VERSION   := $(shell cat version)
BUILD_TIME      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD)

all:
	gox -osarch="darwin/amd64 linux/386 linux/amd64" \
        -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
        -tags="containers_image_openpgp" \
        -ldflags   "-X 'github.com/mritd/imgsync/cmd.version=${BUILD_VERSION}' \
					-X 'github.com/mritd/imgsync/cmd.buildTime=${BUILD_TIME}' \
					-X 'github.com/mritd/imgsync/cmd.commit=${COMMIT_SHA1}'"

pre-release: all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -prerelease --debug ${BUILD_VERSION} dist

release: all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate --debug ${BUILD_VERSION} dist

clean:
	rm -rf dist

install:
	go install -tags="containers_image_openpgp" \
			   -ldflags "-X 'github.com/mritd/imgsync/cmd.version=${BUILD_VERSION}' \
						 -X 'github.com/mritd/imgsync/cmd.buildTime=${BUILD_TIME}' \
						 -X 'github.com/mritd/imgsync/cmd.commit=${COMMIT_SHA1}'"

bin:
	go build -tags="containers_image_openpgp" \
		     -ldflags "-X 'github.com/mritd/imgsync/cmd.version=${BUILD_VERSION}' \
					   -X 'github.com/mritd/imgsync/cmd.buildTime=${BUILD_TIME}' \
					   -X 'github.com/mritd/imgsync/cmd.commit=${COMMIT_SHA1}'"

.PHONY: all release clean install bin

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.cn
