# docker-volume-gce [![Build Status](https://travis-ci.org/mcuadros/docker-volume-gce.svg?branch=master)](https://travis-ci.org/mcuadros/docker-volume-gce)

Docker volume driver for Google Cloud Engine disks.

This driver is designed to run inside of a GCE instance, being able to attach, format and mount [`persistent-disks`](https://cloud.google.com/compute/docs/disks/persistent-disks) to the instance, just as Kubernetes at Google Container Engine does.

License
-------

MIT, see [LICENSE](LICENSE)
