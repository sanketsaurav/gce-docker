package providers

import (
	"fmt"
	"net"
	"net/http"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type NetworkProvider interface {
	Create(c *DiskConfig) error
	Delete(c *DiskConfig) error
}

type Network struct {
	Client
}

func NewNetwork(c *http.Client, project, zone, instance string) (*Network, error) {
	client, err := NewClient(c, project, zone, instance)
	if err != nil {
		return nil, err
	}

	return &Network{Client: *client}, nil
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

	if err := n.createForwardingRules(c); err != nil {
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

	if err != nil {
		return err
	}

	return n.WaitDone(op)

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
	if err != nil {
		return err
	}

	return n.WaitDone(op)
}

func (n *Network) updateTargetPool(old, new *compute.TargetPool) error {
	op, err := n.s.TargetPools.AddInstance(n.project, n.region, new.Name, &compute.TargetPoolsAddInstanceRequest{
		Instances: []*compute.InstanceReference{{
			Instance: InstanceURL(n.project, n.zone, n.instance),
		}},
	}).Do()

	if err != nil {
		return err
	}

	return n.WaitDone(op)
}

func (n *Network) createForwardingRules(c *NetworkConfig) error {
	targetPoolURL := TargetPoolURL(n.project, n.region, c.Name(n.instance))
	for _, rule := range c.ForwardingRule(n.instance, targetPoolURL) {
		if err := n.createForwardingRule(rule); err != nil {
			return err
		}
	}

	return nil
}

func (n *Network) createForwardingRule(rule *compute.ForwardingRule) error {
	if err := n.resolveForwardingRule(rule); err != nil {
		return err
	}

	_, err := n.s.ForwardingRules.Get(n.project, n.region, rule.Name).Do()
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
		return err
	}

	op, err := n.s.ForwardingRules.Insert(n.project, n.region, rule).Do()
	if err != nil {
		return err
	}

	return n.WaitDone(op)
}

func (n *Network) resolveForwardingRule(rule *compute.ForwardingRule) error {
	test := net.ParseIP(rule.IPAddress)
	if test.To4() != nil {
		return nil
	}

	addr, err := n.s.Addresses.Get(n.project, n.region, rule.IPAddress).Do()
	if err != nil {
		return err
	}

	rule.IPAddress = addr.Address
	return nil
}

func (n *Network) createOrUpdateFirewall(c *NetworkConfig) error {
	rule := c.Firewall(n.instance)
	if _, err := n.s.Firewalls.Get(n.project, rule.Name).Do(); err != nil {
		if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
			return err
		}

		op, err := n.s.Firewalls.Insert(n.project, rule).Do()
		if err != nil {
			return err
		}

		return n.WaitDone(op)
	}

	return nil
}

func (n *Network) Delete(c *NetworkConfig) error {
	if err := n.deleteFirewall(c); err != nil {
		return err
	}

	if err := n.deleteForwardingRules(c); err != nil {
		return err
	}

	if err := n.deleteTargetPool(c); err != nil {
		return err
	}

	return nil
}

func (n *Network) deleteFirewall(c *NetworkConfig) error {
	rule := c.Firewall(n.instance)
	op, err := n.s.Firewalls.Delete(n.project, rule.Name).Do()
	if err != nil {
		return err
	}

	return n.WaitDone(op)
}

func (n *Network) deleteForwardingRules(c *NetworkConfig) error {
	targetPoolURL := TargetPoolURL(n.project, n.region, c.Name(n.instance))
	for _, rule := range c.ForwardingRule(n.instance, targetPoolURL) {
		if err := n.deleteForwardingRule(rule); err != nil {
			return err
		}
	}

	return nil
}

func (n *Network) deleteForwardingRule(rule *compute.ForwardingRule) error {
	op, err := n.s.ForwardingRules.Delete(n.project, n.region, rule.Name).Do()
	if err != nil {
		return err
	}

	return n.WaitDone(op)
}

func (n *Network) deleteTargetPool(c *NetworkConfig) error {
	pool := c.TargetPool(n.project, n.zone, n.instance)
	op, err := n.s.TargetPools.Delete(n.project, n.region, pool.Name).Do()
	if err != nil {
		return err
	}

	return n.WaitDone(op)
}
