.PHONY: help run docs build tools db-migrate list-tags connect-prod

TEST_FLAGS ?=
API_CONFIG := $$PWD/config/api.conf
MERCHANT_CONFIG := $$PWD/config/merchant.conf
REPORTER_CONFIG := $$PWD/config/reporter.conf
SCHEDULER_CONFIG := $$PWD/config/scheduler.conf
TOOL_CONFIG := $$PWD/config/tool.conf
SYNCER_CONFIG := $$PWD/config/syncer.conf
TEST_CONFIG := $$PWD/config/test.conf
PEM := ./config/push.pem
TAG_CONTAINER := gcr.io/verdant-descent-153101/api

all:
	@echo "****************************"
	@echo "** Localyyz build tool    **"
	@echo "****************************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  run                   - run API in dev mode"
	@echo "  run-merchant          - run merchant app in dev mode"
	@echo "  run-tool              - run tool app in dev mode"
	@echo "  run-syncer            - run syncer app in dev mode"
	@echo "  run-reporter          - run reporter app in dev mode"
	@echo "  run-scheduler         - run scheduler app in dev mode"
	@echo "  connect-prod          - DANGER: open ssh tunnel to production db"
	@echo ""
	@echo "  eetest                - run end to end tests"
	@echo "  tests                 - run all tests under project"
	@echo "  build                 - build api into bin/ directory"
	@echo "  docs                  - generate api documentation"
	@echo "  tools                 - go get's a bunch of tools for dev"
	@echo ""
	@echo "  db-create             - create dev db"
	@echo "  db-drop               - drop dev db"
	@echo "  db-reset              - reset dev db (drop, create, migrate)"
	@echo "  db-up                 - migrate dev DB to latest version"
	@echo "  db-down               - roll back dev DB to a previous version"
	@echo "  db-migrate            - create new db migration (NAME specifies migration name)"
	@echo "  db-status             - status of current dev DB version"
	@echo "  db-loadstaging        - downloads and loads production database locally"
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

db-loadstaging:
	@./db/db.sh loadstaging localyyz

## TESTS
eetest:
	@(export CONFIG=${TEST_CONFIG}; export DBSCRIPTS=$$PWD/db/db.sh; export MIGRATIONDIR=$$PWD/db; go test -v ./tests/endtoend/)

synctest:
	@(export CONFIG=${TEST_CONFIG}; export DBSCRIPTS=$$PWD/db/db.sh; export MIGRATIONDIR=$$PWD/db; go test -v ./tests/synctest/)
##
# Deploy / GCP
##

list-tags:
	@(export IMAGE=${TAG_CONTAINER}; ./scripts/tags.sh list);

clean-tags:
	@(export IMAGE=${TAG_CONTAINER}; ./scripts/tags.sh clean);

# LOCAL

run:
	@(export CONFIG=${API_CONFIG}; export PEM=${PEM}; fresh -c runner.conf -p ./cmd/api)

run-merchant:
	@(export CONFIG=${MERCHANT_CONFIG}; fresh -c runner.conf -p ./cmd/merchant)

run-tool:
	@(export CONFIG=${TOOL_CONFIG}; go run -v ./cmd/tool/*.go)

run-syncer:
	@(export CONFIG=${SYNCER_CONFIG}; fresh -c runner.conf -p ./cmd/syncer)

run-reporter:
	@(export CONFIG=${REPORTER_CONFIG}; fresh -c runner.conf -p ./cmd/reporter)

run-scheduler:
	@(export CONFIG=${SCHEDULER_CONFIG}; fresh -c runner.conf -p ./cmd/scheduler)

run-nats:
	@nats-streaming-server -DV -cid development

build-merchant:
	@mkdir -p ./bin
	GOGO=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/merchant ./cmd/merchant

build:
	@mkdir -p ./bin
	GOGC=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/api ./cmd/api/main.go

connect-prod:
	@ssh -v -nNT -L 1234:localhost:5432 root@localyyz.db
