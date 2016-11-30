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
	@echo "  tools                 - go get's a bunch of tools for dev"
	@echo ""
	@echo "  db-create             - create dev db"
	@echo "  db-drop               - drop dev db"
	@echo "  db-reset              - reset dev db (drop, create, migrate)"
	@echo "  db-up                 - migrate dev DB to latest version"
	@echo "  db-down               - roll back dev DB to a previous version"
	@echo "  db-status             - status of current dev DB version"
	@echo ""

print-%: ; @echo $*=$($*)


##
## Tools
##
tools:
	go get -u github.com/pressly/sup/cmd/sup
	go get -u bitbucket.org/liamstask/goose/cmd/goose

##
## Database
##
db-status:
	goose status

db-up:
	goose up

db-down:
	goose down

db-create:
	@./db/db.sh create localyyz

db-drop:
	@./db/db.sh drop localyyz

db-reset:
	@./db/db.sh reset localyyz
	goose up

run:
	@(export CONFIG=$$PWD/config/api.conf; export PEM=$$PWD/config/push.pem && go run cmd/api/main.go)

build-goose:
	@mkdir -p ./bin
	GOGO=off go build -i -o ./bin/goose ./vendor/bitbucket.org/liamstask/goose/cmd/goose

build:
	@mkdir -p ./bin
	GOGC=off go build -i -o ./bin/api ./cmd/api/main.go

