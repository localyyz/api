FROM golang:1.8rc3

# API
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api
RUN make build
RUN make build-goose
RUN mv ./bin/api /bin/api
RUN mv ./bin/goose /bin/goose
COPY ./db/migrations/* /migrations/
COPY ./db/dbconf.yml /migrations/dbconf.yml

EXPOSE 5331

CMD ["/bin/api", "-config=/etc/api.conf", "-pem=/etc/push.pem"]
