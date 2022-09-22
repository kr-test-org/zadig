IMAGE_REPOSITORY = koderover.tencentcloudcr.com/koderover-public
VERSION = 20220922

build: MAKE_IMAGE ?= ${IMAGE_REPOSITORY}/greeting-bot:${VERSION}-amd64
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o greeting-bot main.go
	@docker build -f Dockerfile --tag ${MAKE_IMAGE} .