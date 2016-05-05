FROM debian:jessie
MAINTAINER Máximo Cuadros <mcuadros@gmail.com>

ENV DOCKER_HOST unix:///var/run/docker.sock

ADD gce-docker /bin/
CMD ["gce-docker"]
