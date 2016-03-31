FROM alpine:3.1
MAINTAINER Máximo Cuadros <mcuadros@gmail.com>

RUN apk add --update util-linux ca-certificates && rm -rf /var/cache/apk/*
ENV DOCKER_HOST unix:///var/run/docker.sock

ADD gce-docker /bin/
CMD ["gce-docker"]
