FROM debian:latest

RUN apt-get update && apt-get install libopus0 golang-go libopus-dev gcc pkg-config python3 python3-pip git -y && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

pip install py-postgresql
pip install zeroc-ice

ENV GOPATH=/go
ADD . ./go/src/github.com/TF2Stadium/fumble/
RUN go get -v github.com/TF2Stadium/fumble/...
RUN go install -v github.com/TF2Stadium/fumble
ADD ./mumble-authenticator /mumble-authenticator/
ADD ./entrypoint.sh /entrypoint.sh

RUN ls /go/bin/
ENTRYPOINT /entrypoint.sh
