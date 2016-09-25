FROM golang:1.7

# API
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api

RUN make build
RUN mv ./bin/api /bin/api

EXPOSE 5331

CMD ["/bin/api", "-config=/etc/api.conf", "-pem=/etc/push.pem"]
