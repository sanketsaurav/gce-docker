package manager

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var NetworkBaseName = "gce-network-%s-%s"

type Network struct {
	s        *compute.Service
	zone     string
	region   string
	project  string
	instance string
}

func NewNetwork(c *http.Client, project, zone, instance string) (*Network, error) {
	s, err := compute.New(c)
	if err != nil {
		return nil, err
	}

	n := &Network{
		s:        s,
		project:  project,
		zone:     zone,
		instance: instance,
	}

	return n, n.loadRegion()
}

func (n *Network) loadRegion() error {
	z, err := n.s.Zones.Get(n.project, n.zone).Do()
	if err != nil {
		return fmt.Errorf("error retrieving region from zone: %s", err)
	}

	region := strings.Split(z.Region, "/")
	n.region = region[len(region)-1]
	return nil
}

func (n *Network) Create(c *NetworkConfig) error {
	if err := c.Validate(); err != nil {
		return err
	}

	if err := n.updateInstanceTags(c); err != nil {
	}

	if err := n.createOrUpdateTargetPool(c); err != nil {
		return fmt.Errorf("error creating/updating target pool: %s", err)
	}

	if err := n.createForwardingRule(c); err != nil {
		return fmt.Errorf("error creating forwarding rule: %s", err)
	}

	if err := n.createOrUpdateFirewall(c); err != nil {
		return fmt.Errorf("error creating firewall rule: %s", err)
	}

	return nil
}

func (n *Network) updateInstanceTags(c *NetworkConfig) error {
	i, err := n.s.Instances.Get(n.project, n.zone, n.instance).Do()
	if err != nil {
		return err
	}

	tag := c.Name(n.instance)
	if contains(i.Tags.Items, tag) {
		return nil
	}

	op, err := n.s.Instances.SetTags(n.project, n.zone, n.instance, &compute.Tags{
		Items:       append(i.Tags.Items, tag),
		Fingerprint: i.Tags.Fingerprint,
	}).Do()

	fmt.Println(op, err)
	return err

}

func (n *Network) createOrUpdateTargetPool(c *NetworkConfig) error {
	new := c.TargetPool(n.project, n.zone, n.instance)
	old, err := n.s.TargetPools.Get(n.project, n.region, new.Name).Do()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
			return err
		}

		return n.createTargetPool(new)
	}

	return n.updateTargetPool(old, new)
}

func (n *Network) createTargetPool(pool *compute.TargetPool) error {
	op, err := n.s.TargetPools.Insert(n.project, n.region, pool).Do()
	fmt.Println("create", op)

	return err
}

func (n *Network) updateTargetPool(old, new *compute.TargetPool) error {
	op, err := n.s.TargetPools.AddInstance(n.project, n.region, new.Name, &compute.TargetPoolsAddInstanceRequest{
		Instances: []*compute.InstanceReference{{
			Instance: InstanceURL(n.project, n.zone, n.instance),
		}},
	}).Do()
	fmt.Println("update", op)

	return err
}

func (n *Network) createForwardingRule(c *NetworkConfig) error {
	targetPoolURL := TargetPoolURL(n.project, n.region, c.Name(n.instance))

	rule := c.ForwardingRule(n.instance, targetPoolURL)
	_, err := n.s.ForwardingRules.Get(n.project, n.region, rule.Name).Do()
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
		return err
	}

	op, err := n.s.ForwardingRules.Insert(n.project, n.region, rule).Do()
	fmt.Println("create, rule", op)

	return err
}

func (n *Network) createOrUpdateFirewall(c *NetworkConfig) error {
	new := c.Firewall(n.instance)
	old, err := n.s.Firewalls.Get(n.project, new.Name).Do()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
			return err
		}
	}

	op, err := n.s.Firewalls.Insert(n.project, new).Do()
	if err != nil {
		return err
	}

	fmt.Println("create, firewall", op, old, err)

	return nil
}
