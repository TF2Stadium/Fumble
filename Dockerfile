FROM alpine
RUN apk add --update gcc pkgconfig python3 libbz2 openssl openssl-dev 

RUN pip3 install py-postgresql
RUN pip3 install zeroc-ice

ADD fumble /bin/fumble
ADD ./mumble-authenticator /mumble-authenticator/
ADD ./entrypoint.sh /entrypoint.sh

RUN rm -rf /tmp/* /var/tmp/* /tmp/pip-build-root/*
ENTRYPOINT /entrypoint.sh
