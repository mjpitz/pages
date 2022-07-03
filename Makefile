CWD = $(shell pwd)
SKAFFOLD_DEFAULT_REPO ?= img.pitz.tech/mya
VERSION ?= latest

define HELP_TEXT
Welcome to pages!

Targets:
  help             provides help text
  test             run tests
  docker           rebuild the pages docker container
  docker/release   releases pages
  legal            prepends legal header to source code
  dist             recompiles pages binaries

endef
export HELP_TEXT

help:
	@echo "$$HELP_TEXT"

docker: .docker
.docker:
	docker build . \
		--tag $(SKAFFOLD_DEFAULT_REPO)/pages:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/pages:$(VERSION) \
		--file ./cmd/pages/Dockerfile

docker/release:
	docker buildx build . \
		--platform linux/amd64,linux/arm64 \
		--label "org.opencontainers.image.source=https://code.pitz.tech/mya/pages" \
		--label "org.opencontainers.image.version=$(VERSION)" \
		--label "org.opencontainers.image.licenses=agpl3" \
		--label "org.opencontainers.image.title=pages" \
		--label "org.opencontainers.image.description=" \
		--tag $(SKAFFOLD_DEFAULT_REPO)/pages:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/pages:$(VERSION) \
		--file ./cmd/pages/Dockerfile \
		--push

# actual targets

test:
	go test -v -race -coverprofile=.coverprofile -covermode=atomic ./...

legal: .legal
.legal:
	addlicense -f ./legal/header.txt -skip yaml -skip yml .

dist: .dist
.dist:
	sh ./scripts/dist-go.sh

# useful shortcuts for release

tag/release:
	npm version "$(shell date +%y.%m.0)"
	git push --follow-tags

tag/patch:
	npm version patch
	git push --follow-tags
