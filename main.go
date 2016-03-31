package main

import (
	"net/http"
	"os"

	"github.com/mcuadros/gce-docker/plugin"
	"github.com/mcuadros/gce-docker/watcher"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/fsouza/go-dockerclient"
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
	go executeWatcher(c, project, zone, instance)
	go executeVolumePlugin(c, project, zone, instance)
	select {}
}

func executeWatcher(c *http.Client, project, zone, instance string) {
	log15.Info("starting watcher", "project", project, "zone", zone, "instance", instance)
	d, err := docker.NewClientFromEnv()
	if err != nil {
		log15.Error("error creating docker client", "error", err)
		os.Exit(1)
	}

	w, err := watcher.NewWatcher(d, c, project, zone, instance)
	if err != nil {
		log15.Error("error creating watcher", "error", err)
		os.Exit(1)
	}

	if err := w.Watch(); err != nil {
		log15.Error("error starting watcher", "error", err)
		os.Exit(1)
	}
}

func executeVolumePlugin(c *http.Client, project, zone, instance string) {
	log15.Info("starting volume driver", "project", project, "zone", zone, "instance", instance)
	d, err := plugin.NewVolume(c, project, zone, instance)
	if err != nil {
		log15.Error("error creating volume plugin", "error", err)
		os.Exit(1)
	}

	h := volume.NewHandler(d)
	if err := h.ServeUnix("docker", "gce"); err != nil {
		log15.Error("error starting volume driver server", "error", err)
		os.Exit(1)
	}

	log15.Info("volume plugin: http server started")
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
