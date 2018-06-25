# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
# FROM gobuffalo/buffalo:<version> as builder

# RUN mkdir -p $GOPATH/src/github.com/oysterprotocol/brokernode
# WORKDIR $GOPATH/src/github.com/oysterprotocol/brokernode

# ADD . .
# RUN go get $(go list ./... | grep -v /vendor/)
# RUN buffalo build --static -o /bin/app

# FROM alpine
# RUN apk add --no-cache bash
# RUN apk add --no-cache ca-certificates

# WORKDIR /bin/

# COPY --from=builder /bin/app .

# # Comment out to run the binary in "production" mode:
# # ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
# ENV ADDR=0.0.0.0

# EXPOSE 3000

# # Comment out to run the migrations before running the binary:
# # CMD /bin/app migrate; /bin/app
# CMD exec /bin/app

FROM golang:1.10
ENV ADDR=0.0.0.0

RUN go version

# Install db client (assumes mysql)
RUN apt-get update
RUN apt-get install -y -q mysql-client
RUN apt-get install -y -q netcat

RUN mkdir -p $GOPATH/src/github.com/oysterprotocol/brokernode
WORKDIR $GOPATH/src/github.com/oysterprotocol/brokernode

RUN go get -u github.com/gobuffalo/buffalo/buffalo

COPY . .

RUN go get -t -d -v ./...
RUN go install -v ./...

RUN buffalo version
