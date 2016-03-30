package providers

import . "gopkg.in/check.v1"

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) TestNetworkConfigDisk(c *C) {
	config := &DiskConfig{
		Name:           "foo",
		Type:           "qux",
		SizeGb:         42,
		SourceSnapshot: "bar",
		SourceImage:    "baz",
	}

	d := config.Disk()
	c.Assert(d.Name, Equals, "foo")
	c.Assert(d.Type, Equals, "qux")
	c.Assert(d.SizeGb, Equals, int64(42))
	c.Assert(d.SourceSnapshot, Equals, "bar")
	c.Assert(d.SourceImage, Equals, "baz")
}

func (s *ConfigSuite) TestNetworkConfigValidate(c *C) {
	config := &DiskConfig{}
	err := config.Validate()
	c.Assert(err, NotNil)

	config = &DiskConfig{Name: "foo"}
	err = config.Validate()
	c.Assert(err, IsNil)

	config = &DiskConfig{Name: "foo", SourceSnapshot: "foo", SourceImage: "foo"}
	err = config.Validate()
	c.Assert(err, NotNil)
}

func (s *ConfigSuite) TestNetworkConfigDeviceName(c *C) {
	config := &DiskConfig{Name: "foo"}
	c.Assert(config.DeviceName(), Equals, "docker-volume-foo")
}

func (s *ConfigSuite) TestNetworkConfigDev(c *C) {
	config := &DiskConfig{Name: "docker-volume-foo"}
	c.Assert(config.Dev(), Equals, "/dev/disk/by-id/google-docker-volume-foo")
}

func (s *ConfigSuite) TestNetworkConfigMountPoint(c *C) {
	config := &DiskConfig{Name: "foo"}
	c.Assert(config.MountPoint("/mnt/"), Equals, "/mnt/foo")
}

func (s *ConfigSuite) TestNetworkConfigGroup(c *C) {
	config := &NetworkConfig{Container: "bar"}
	c.Assert(config.Group("foo"), Equals, "bar-foo")

	config = &NetworkConfig{GroupName: "qux"}
	c.Assert(config.Group("foo"), Equals, "qux")
}

func (s *ConfigSuite) TestNetworkConfigID(c *C) {
	config := &NetworkConfig{
		Container: "foo",
		Address:   "qux",
		Protocol:  "bar",
		Port:      "baz",
	}

	c.Assert(config.ID("42"), Equals, "f116e019")
}

func (s *ConfigSuite) TestNetworkConfigName(c *C) {
	config := &NetworkConfig{GroupName: "bar"}
	c.Assert(config.Name("foo"), Equals, "docker-network-bar-57992c1d")
}

func (s *ConfigSuite) TestNetworkConfigTargetPool(c *C) {
	config := &NetworkConfig{
		Container:       "bar",
		SessionAffinity: SessionAffinity("qux"),
	}

	tp := config.TargetPool("bar", "baz", "foo")
	c.Assert(tp.Name, Equals, "docker-network-bar-foo-9b044df6")
	c.Assert(tp.Instances, HasLen, 1)
	c.Assert(tp.Instances[0], Equals, "https://www.googleapis.com/compute/v1/projects/bar/zones/baz/instances/foo")
	c.Assert(tp.SessionAffinity, Equals, "qux")
}
