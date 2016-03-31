# Google Cloud Engine integration for Docker [![Build Status](https://travis-ci.org/mcuadros/gce-docker.svg?branch=master)](https://travis-ci.org/mcuadros/gce-docker)

__gce-docker__ is a service that provides integration with the GCE to Docker, the following resources are supported:

- __Persistent Disks__, the service is able to attach, format and mount [_persistent-disks_](https://cloud.google.com/compute/docs/disks/persistent-disks) allowing to use it as volumes in the container
- __Load Balancers & External IPs__: support from auto-creation of LoadBanacers and External IPs allowing direcct access to the container.


Examples
--------

#### Creating a Persistent Disk and mount is a volume to a Container

```sh
docker run -ti -v my-disk:/data --volume-driver=gce busybox df -h /data

```

#### Creating a simple Load Balancer with a static IP

```sh
docker run -d --label gce.lb.address=104.197.200.230 --label gce.lb.type=static -p 80:80tutum/hello-world
```


Installing
----------
The recommended way to install `gce-docker` is use the provided docker image.

Run the driver using the following command:
```sh
docker run -d -v /:/rootfs -v /run/docker/plugins:/run/docker/plugins -v /var/run/docker.sock:/var/run/docker.sock --privileged mcuadros/gce-docker
```

`privileged` is required since `gce-docker` needs low level access to the host mount namespace, the driver mounts, umounts and format disk.

> The instance requires `Read/Write` privileges to Google Compute Engine and IP forwarding flags should be active to.

Usage
-----

### Persistent Disks
#### Persistent disk creation

Using `docker volume create` a new disk is created.
```sh
docker volume create --driver=gce --name my-disk -o SizeGb=90
```

Options:
- __Type__ (_optional, default:pd-ssd_, options: `pd-ssd` or `pd-standard`):  Disk type to use to create the disk.
- __SizeGb__ (optional):  Size of the persistent disk, specified in GB.
- __SourceSnapshot__ (optional): The source snapshot used to create this disk.
- __SourceImaget__ (optional): The source image used to create this disk.


#### Using a disk on your container

Just add the flags `--volume-driver=gce` and the `-v <disk-name>:/data` to any docker run command:

```sh
docker run -ti -v my-disk:/data --volume-driver=gce busybox sh
```

If the disk already exists will be used, if not a new one with the default values will be created (Standard/500GB)

The disk is attached to the instance, if the disk is not formatted also is formatted with `ext4`, when the container stops, the disk is unmounted and detached.



### Load Balancer
The load balancers, are handle by a watcher, waiting for Docker events, the watched events are `start` and `die`. When a new containeris created or destroyed, the LoadBalancer and all the others dependant resources are created or deleted too.

This is a small example create a LoadBalancer for a web server:
```sh
docker run -d --label gce.lb.type=ephemeral -p 80:80 tutum/hello-world
```

Available labels:
- __gce.lb.type__ (options: `ephemeral` or `static`):  Type of IP to be used in the new load balancer
- __gce.lb.group__ (optional):  Name of group of instances to assign to the same load balancer. If not provided a combination of instance name and container id will be used.
- __gce.lb.address__ (optional, required with type `static`): Value of the reserved IP address that the forwarding rule is serving on behalf of.
- __gce.lb.source.ranges__ (optional): The IP address blocks that this load balancer applies to expressed in CIDR format. One or both of sourceRanges and sourceTags may be set.
- __gce.lb.source.tags__ (optional):A list of instance tags which this rule applies to. One or both of sourceRanges and sourceTags may be set.
- __gce.lb.session.affinity__ (optional): Sesssion affinity option, must be one of the following values:
  - `NONE`: Connections from the same client IP may go to any instance in the pool.
  - `CLIENT_IP`: Connections from the same client IP will go to the same instance in the pool while that instance remains healthy.
  - `CLIENT_IP_PROTO`: Connections from the same client IP with the same IP protocol will go to the same instance in the pool while that instance remains healthy.




License
-------

MIT, see [LICENSE](LICENSE)
