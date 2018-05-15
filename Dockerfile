# STAGE 1: BUILD API
FROM gcr.io/verdant-descent-153101/golang
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api
RUN mkdir -p ./bin
RUN GOGC=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -a -installsuffix cgo -i -o ./bin/api ./cmd/api/main.go
RUN GOGO=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -a -installsuffix cgo -i -o ./bin/merchant ./cmd/merchant
RUN GOGO=off CGO_ENABLED=0 GOOS=linux go build -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -a -installsuffix cgo -i -o ./bin/tool ./cmd/tool

# STAGE 2: SCRATCH BINARY
FROM scratch
COPY ./db /db
COPY ./merchant/index.html /merchant/index.html
COPY ./merchant/approval.html /merchant/approval.html
COPY ./merchant/approvallist.html /merchant/approvallist.html
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/api /bin/api
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/merchant /bin/merchant
COPY --from=0 /go/src/bitbucket.org/moodie-app/moodie-api/bin/tool /bin/tool
COPY --from=0 /bin/goose /bin/goose
ADD ca-certificates.crt /etc/ssl/certs/

EXPOSE 5331
EXPOSE 5333
EXPOSE 5335

CMD ["/bin/api", "-config=/etc/api.conf", "-pem=/etc/push.pem"]
