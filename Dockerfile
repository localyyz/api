# STAGE 1: BUILD API
FROM gcr.io/verdant-descent-153101/golang
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api
RUN mkdir -p ./bin
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -a -i -o ./bin/api ./cmd/api/main.go
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/merchant ./cmd/merchant
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/tool ./cmd/tool
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/syncer ./cmd/syncer
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/reporter ./cmd/reporter
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/scheduler ./cmd/scheduler

# STAGE 2: SCRATCH BINARY
FROM scratch
COPY ./db /db
COPY ./merchant/index.html /merchant/index.html
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/api /bin/api
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/merchant /bin/merchant
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/tool /bin/tool
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/syncer /bin/syncer
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/reporter /bin/reporter
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/scheduler /bin/scheduler
COPY --from=0 /bin/goose /bin/goose
ADD ca-certificates.crt /etc/ssl/certs/

EXPOSE 5331
EXPOSE 5333
EXPOSE 5335
EXPOSE 5337
EXPOSE 5339
EXPOSE 5341

CMD ["/bin/api", "-config=/etc/api.conf", "-pem=/etc/push.pem"]
