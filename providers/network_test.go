package providers

import . "gopkg.in/check.v1"

type NetworkSuite struct {
	BaseSuite
}

var _ = Suite(&NetworkSuite{})

func (s *NetworkSuite) TestCreate(c *C) {
	if !*integration {
		c.Skip("-integration not provided")
	}

	n, err := NewNetwork(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	config := &NetworkConfig{
		Container: "test",
		Protocol:  "tcp",
		Port:      "8000",
	}

	err = n.Create(config)
	c.Assert(err, IsNil)

	err = n.Delete(config)
	c.Assert(err, IsNil)
}
