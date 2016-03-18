FROM golang:latest

RUN apt-get update && apt-get install libopus0 libopus-dev  pkg-config python3 python3-pip git libbz2-dev openssl libssl-dev -y && \
    apt-get clean && apt-get purge && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN pip3 install py-postgresql
RUN pip3 install zeroc-ice

ENV FUMBLE_PROFILER_ADDR=0.0.0.0:80
ENV GOPATH=/go
ADD . ./go/src/github.com/TF2Stadium/fumble/
RUN go get -v github.com/TF2Stadium/fumble/...
RUN go install -v github.com/TF2Stadium/fumble
ADD ./mumble-authenticator /mumble-authenticator/
ADD ./entrypoint.sh /entrypoint.sh

RUN rm -rf /tmp/* /var/tmp/* /tmp/pip-build-root/* /go/src/* && apt-get remove gcc -y && apt-get purge && apt-get autoremove -y
ENTRYPOINT /entrypoint.sh
