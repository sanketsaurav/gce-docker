package main

import (
	"os"

	"github.com/mcuadros/docker-volume-gce/driver"

	"github.com/docker/go-plugins-helpers/volume"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/cloud/compute/metadata"
	"gopkg.in/inconshreveable/log15.v2"
)

const DriverName = "gce"

func main() {
	if !metadata.OnGCE() {
		log15.Error("docker-volume-gce driver only runs on Google Compute Engine")
		os.Exit(126)
	}

	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log15.Error("error building compute client", "error", err)
		os.Exit(1)
	}

	project, zone, instance := getMetadataInfo()
	log15.Info("starting volume driver", "project", project, "zone", zone, "instance", instance)

	d, err := driver.NewVolume(c, project, zone, instance)
	if err != nil {
		log15.Error("error creating volume driver", "error", err)
		os.Exit(1)
	}

	h := volume.NewHandler(d)
	if err := h.ServeTCP(DriverName, ":0"); err != nil {
		log15.Error("error starting volume driver server", "error", err)
		os.Exit(1)
	}

	log15.Info("http server started")
}

func getMetadataInfo() (project string, zone string, instance string) {
	var err error
	instance, err = metadata.InstanceName()
	if err != nil {
		log15.Error("error retrieving instance name", "error", err)
		os.Exit(126)
	}

	zone, err = metadata.Zone()
	if err != nil {
		log15.Error("error retrieving zone", "error", err)
		os.Exit(126)
	}

	project, err = metadata.ProjectID()
	if err != nil {
		log15.Error("error retrieving project", "error", err)
		os.Exit(126)
	}

	return
}
