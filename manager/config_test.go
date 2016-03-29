package manager

import . "gopkg.in/check.v1"

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

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

	c.Assert(config.ID("42"), Equals, "f116e019f360e2e9c5ab8e6aa9777b52")
}

func (s *ConfigSuite) TestNetworkConfigName(c *C) {
	config := &NetworkConfig{GroupName: "bar"}
	c.Assert(config.Name("foo"), Equals, "gce-driver-bar-57992c1d0c0e6073e6dbaf64c39464df")
}

func (s *ConfigSuite) TestNetworkConfigTargetPool(c *C) {
	config := &NetworkConfig{
		Container:       "bar",
		SessionAffinity: SessionAffinity("qux"),
	}

	tp := config.TargetPool("bar", "baz", "foo")
	c.Assert(tp.Name, Equals, "gce-driver-bar-foo-9b044df67da785a880c43e77af0aac8b")
	c.Assert(tp.Instances, HasLen, 1)
	c.Assert(tp.Instances[0], Equals, "projects/bar/zones/baz/instances/foo")
	c.Assert(tp.SessionAffinity, Equals, "qux")
}
