FROM golang:1.6.2

RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates xmlsec1

# API
ADD . /go/src/bitbucket.org/moodie-app/moodie-api
WORKDIR /go/src/bitbucket.org/moodie-app/moodie-api

RUN echo $PWD
RUN ls -al

RUN make build
COPY bin/api /bin/api

EXPOSE 5331

CMD ["/bin/api", "-config=/etc/api.conf"]
