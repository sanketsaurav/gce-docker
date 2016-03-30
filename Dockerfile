FROM alpine:3.1
MAINTAINER Máximo Cuadros <mcuadros@gmail.com>

RUN apk add --update util-linux && rm -rf /var/cache/apk/*

ADD docker-volume-gce /bin/
CMD ["docker-volume-gce"]
