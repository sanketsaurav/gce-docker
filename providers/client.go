package providers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

const MaxWaitDuration = time.Minute

type Client struct {
	s        *compute.Service
	zone     string
	region   string
	project  string
	instance string
}

func NewClient(c *http.Client, project, zone, instance string) (*Client, error) {
	s, err := compute.New(c)
	if err != nil {
		return nil, err
	}

	client := &Client{
		s:        s,
		project:  project,
		zone:     zone,
		instance: instance,
	}

	return client, client.loadRegion()
}

func (c *Client) loadRegion() error {
	z, err := c.s.Zones.Get(c.project, c.zone).Do()
	if err != nil {
		return fmt.Errorf("error retrieving region from zone: %s", err)
	}

	region := strings.Split(z.Region, "/")
	c.region = region[len(region)-1]
	return nil
}

func (c *Client) WaitDone(op *compute.Operation) error {
	var doer func(...googleapi.CallOption) (*compute.Operation, error)
	switch {
	case op.Region != "":
		doer = c.s.RegionOperations.Get(c.project, c.region, op.Name).Do
	case op.Zone != "":
		doer = c.s.ZoneOperations.Get(c.project, c.zone, op.Name).Do
	default:
		doer = c.s.GlobalOperations.Get(c.project, op.Name).Do
	}

	start := time.Now()
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		rop, err := doer()
		if err != nil {
			log15.Error("error waiting for operation %q: %s", "name", op.Name, err)
			continue
		}

		if rop.Status == "DONE" {
			return nil
		}

		if time.Since(start) > MaxWaitDuration {
			return fmt.Errorf("max. time reached waiting for operation %q", op.Name)
		}
	}

	return nil
}
