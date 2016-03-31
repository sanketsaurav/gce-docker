# docker-volume-gce [![Build Status](https://travis-ci.org/mcuadros/docker-volume-gce.svg?branch=master)](https://travis-ci.org/mcuadros/docker-volume-gce)

Docker volume driver for Google Cloud Engine disks.

This driver is designed to run inside of a GCE instance, being able to attach, format and mount [`persistent-disks`](https://cloud.google.com/compute/docs/disks/persistent-disks) to the instance, just as Kubernetes at Google Container Engine does.


Installing
----------
The recommended way to install `docker-volume-gce` is use the provided docker image.

Run the driver using the following command:
```sh
docker run -d -v /:/rootfs -v /run/docker/plugins:/run/docker/plugins --privileged mcuadros/docker-volume-gce
```

`privileged` is required since the driver needs low level access to the host mount namespace, the driver mounts, umounts and format disk.

> The instance requires `Read/Write` privileges to Google Compute Engine.

Usage
-----
### Persistent disk creation

Using `docker volume create` a new disk is created.
```sh
docker volume create --driver=gce --name my-disk -o SizeGb=90
```

### Using a disk on your container

Just add the flags `--volume-driver=gce` and the `-v <disk-name>:/data` to any docker run command:

```sh
docker run -ti -v my-disk:/data --volume-driver=gce busybox sh
```

If the disk already exists will be used, if not a new one with the default values will be created (Standard/500GB)

The disk is attached to the instance, if the disk is not formatted also is formatted with `ext4`, when the container stops, the disk is unmounted and detached.


License
-------

MIT, see [LICENSE](LICENSE)
