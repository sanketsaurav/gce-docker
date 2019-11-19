FROM golang:1.13.4-buster

RUN apt-get update \
	&& apt-get install -y ca-certificates \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV DOCKER_HOST unix:///var/run/docker.sock

RUN mkdir -p /go/src/github.com/bloomapi/gce-docker
ADD . /go/src/github.com/bloomapi/gce-docker
WORKDIR /go/src/github.com/bloomapi/gce-docker
RUN go get -d ./...

# This is super hacky, but I couldn't get this to work with dep as I'm not sure how to resolve version miss-matches
RUN cd /go/src/github.com/docker/go-plugins-helpers; git checkout dd9c6831a796ea025dfae8448d6a34d081b99898

RUN cd /go/src/github.com/docker/go-connections; git checkout 4ccf312bf1d35e5dbda654e57a9be4c3f3cd0366

RUN cd /go/src/github.com/bloomapi/gce-docker

RUN go get -d ./...

RUN go install .

CMD ["/go/bin/gce-docker"]
