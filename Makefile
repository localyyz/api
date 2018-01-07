.PHONY: help run docs build tools db-migrate

TEST_FLAGS ?=
API_CONFIG := $$PWD/config/api.conf
MERCHANT_CONFIG := $$PWD/config/merchant.conf
PEM := ./config/push.pem

all:
	@echo "****************************"
	@echo "** Localyyz build tool    **"
	@echo "****************************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  run                   - run API in dev mode"
	@echo "  run-merchant          - run merchant app in dev mode"
	@echo "  build                 - build api into bin/ directory"
	@echo "  docs                  - generate api documentation"
	@echo "  tools                 - go get's a bunch of tools for dev"
	@echo ""
	@echo "  db-create             - create dev db"
	@echo "  db-drop               - drop dev db"
	@echo "  db-reset              - reset dev db (drop, create, migrate)"
	@echo "  db-up                 - migrate dev DB to latest version"
	@echo "  db-down               - roll back dev DB to a previous version"
	@echo "  db-migrate            - create new db migration"
	@echo "  db-status             - status of current dev DB version"
	@echo ""

print-%: ; @echo $*=$($*)


##
## Tools
##
tools:
	go get -u github.com/pressly/sup/cmd/sup
	go get -u github.com/pressly/fresh
	go get -u bitbucket.org/liamstask/goose/cmd/goose

docs:
	go run ./docs/main.go

##
## Database
##
db-status:
	goose status

db-up:
	goose up

db-down:
	goose down

db-migrate:
	goose create ${NAME} sql

db-create:
	@./db/db.sh create localyyz

db-drop:
	@./db/db.sh drop localyyz

db-reset:
	@./db/db.sh reset localyyz
	goose up

db-loadprod:
	@./db/db.sh loadprod localyyz

run:
	@(export CONFIG=${API_CONFIG}; export PEM=${PEM}; fresh -c runner.conf -p ./cmd/api)

run-merchant:
	@(export CONFIG=${MERCHANT_CONFIG}; fresh -c runner.conf -p ./cmd/merchant)

build-merchant:
	@mkdir -p ./bin
	GOGO=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/merchant ./cmd/merchant

build:
	@mkdir -p ./bin
	GOGC=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/api ./cmd/api/main.go
