FROM jpetazzo/nsenter
MAINTAINER Máximo Cuadros <mcuadros@gmail.com>

ADD docker-volume-gce /bin/
EXPOSE 5678

CMD ["docker-volume-gce"]
