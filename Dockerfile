# STAGE 1: BUILD API
FROM gcr.io/verdant-descent-153101/golang
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api
RUN mkdir -p ./bin
RUN GOGC=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/api ./cmd/api/main.go
RUN GOGO=off go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -i -o ./bin/merchant ./cmd/merchant

# STAGE 2: BINARY
FROM scratch
COPY ./db /db
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/api /bin/api
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/merchant /bin/merchant

EXPOSE 5331
EXPOSE 5333

CMD ["/bin/api", "-config=/etc/api.conf", "-pem=/etc/push.pem"]
