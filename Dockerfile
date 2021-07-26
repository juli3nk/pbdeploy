FROM golang:1.16-alpine3.14 AS builder

RUN apk --update add \
		ca-certificates \
		gcc \
		git \
		musl-dev

WORKDIR /go/src/github.com/juli3nk/pbdeploy

COPY . .

ENV GO111MODULE=off
RUN go get \
	&& go build -ldflags "-linkmode external -extldflags -static -s -w" -trimpath -o /tmp/pbdeploy


FROM alpine:3.14

RUN apk --update --no-cache add \
		ca-certificates \
		npm

COPY --from=builder /tmp/pbdeploy /usr/local/bin/pbdeploy

ENTRYPOINT ["/usr/local/bin/pbdeploy"]
