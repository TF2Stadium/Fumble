FROM alpine

ADD fumble /bin/fumble

ENTRYPOINT /bin/fumble