.PHONY: help run test coverage docs build dist clean tools dist-tools vendor-list vendor-update

TEST_FLAGS ?=

all:
	@echo "****************************"
	@echo "** Pressly API build tool **"
	@echo "****************************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  run                   - run API in dev mode"
	@echo "  build                 - build api into bin/ directory"
	@echo "  build-all             - build all binaries into bin/ directory"
	@echo ""

print-%: ; @echo $*=$($*)

run:
	@(export CONFIG=$$PWD/config/api.conf && go run main.go)

bindir:
	@mkdir -p ./bin

build: bindir
	GOGC=off go build -i -o ./bin/api ./
