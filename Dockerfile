FROM golang:alpine

RUN apk add --update opus git pkgconfig opus-dev gcc

ADD . ./go/src/github.com/TF2Stadium/fumble/
RUN go get -v github.com/TF2Stadium/fumble/...
RUN go install -v github.com/TF2Stadium/fumble

ENTRYPOINT go/bin/fumble