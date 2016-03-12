package main

import (
	"log"

	"github.com/mcuadros/docker-volume-gce/driver"

	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	d := &driver.Volume{}
	h := volume.NewHandler(d)
	log.Fatal(h.ServeTCP("test", ":57895"))
}
