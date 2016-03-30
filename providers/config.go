package providers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/fsouza/go-dockerclient"

	"google.golang.org/api/compute/v1"
)

var (
	NetworkBaseName        = "docker-network-%s-%s"
	DiskDeviceNameBaseName = "docker-volume-%s"
	DiskDevBasePath        = "/dev/disk/by-id/google-%s"
)

type DiskConfig struct {
	Name           string
	Type           string
	SizeGb         int64
	SourceSnapshot string
	SourceImage    string
}

func (c *DiskConfig) Disk() *compute.Disk {
	return &compute.Disk{
		Name:           c.Name,
		Type:           c.Type,
		SizeGb:         c.SizeGb,
		SourceSnapshot: c.SourceSnapshot,
		SourceImage:    c.SourceImage,
	}
}

func (c *DiskConfig) DeviceName() string {
	return fmt.Sprintf(DiskDeviceNameBaseName, c.Name)
}

func (c *DiskConfig) Dev() string {
	return fmt.Sprintf(DiskDevBasePath, c.DeviceName())
}

func (c *DiskConfig) MountPoint(root string) string {
	return filepath.Join(root, c.Name)
}

func (c *DiskConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("invalid disk config, name field cannot be empty")
	}

	if c.SourceSnapshot != "" && c.SourceImage != "" {
		return fmt.Errorf("invalid dick config, source snapshot and source image can't be presents at the same time.")
	}

	return nil
}

type SessionAffinity string
type NetworkConfig struct {
	GroupName string
	Container string
	Network   string
	Address   string
	Ports     []docker.Port
	Source    struct {
		Ranges []string
		Tags   []string
	}
	SessionAffinity SessionAffinity
}

func (c *NetworkConfig) TargetPool(project, zone, instance string) *compute.TargetPool {
	return &compute.TargetPool{
		Name:            c.Name(instance),
		Instances:       []string{InstanceURL(project, zone, instance)},
		SessionAffinity: string(c.SessionAffinity),
	}
}

func (c *NetworkConfig) ForwardingRule(instance, targetPoolURL string) []*compute.ForwardingRule {
	var rules []*compute.ForwardingRule
	for _, p := range c.Ports {
		rules = append(rules, &compute.ForwardingRule{
			Name:       fmt.Sprintf("%s-%s-%s", c.Name(instance), p.Port(), p.Proto()),
			IPAddress:  c.Address,
			IPProtocol: p.Proto(),
			PortRange:  p.Port(),
			Target:     targetPoolURL,
		})
	}

	return rules
}

func (c *NetworkConfig) Firewall(instance string) *compute.Firewall {
	sourceRanges := c.Source.Ranges
	if len(c.Source.Ranges) == 0 && len(c.Source.Tags) == 0 {
		sourceRanges = []string{"0.0.0.0/0"}
	}

	network := c.Network
	if len(network) == 0 {
		network = "global/networks/default"
	}

	name := c.Name(instance)
	var allowed []*compute.FirewallAllowed
	for _, p := range c.Ports {
		allowed = append(allowed, &compute.FirewallAllowed{
			IPProtocol: p.Proto(),
			Ports:      []string{p.Port()},
		})
	}

	return &compute.Firewall{
		Name:         name,
		SourceRanges: sourceRanges,
		SourceTags:   c.Source.Tags,
		TargetTags:   []string{name},
		Network:      network,
		Allowed:      allowed,
	}
}

func (c *NetworkConfig) Name(instance string) string {
	return fmt.Sprintf(NetworkBaseName, c.Group(instance), c.ID(instance))
}

func (c *NetworkConfig) Group(instance string) string {
	if c.GroupName != "" {
		return c.GroupName
	}

	return fmt.Sprintf("%s-%s", c.Container, instance)
}

func (c *NetworkConfig) ID(instance string) string {
	var unique string
	unique += c.Group(instance)
	unique += c.Address
	for _, p := range c.Ports {
		unique += string(p)
	}

	hash := md5.Sum([]byte(unique))
	return hex.EncodeToString(hash[:])[:8]
}

func (c *NetworkConfig) Validate() error {
	if c.Container == "" {
		return fmt.Errorf("invalid network config, container field cannot be empty")
	}

	if len(c.Ports) == 0 {
		return fmt.Errorf("invalid network config, ports field cannot be empty")
	}

	return nil
}
