package providers

import (
	"github.com/fsouza/go-dockerclient"
	. "gopkg.in/check.v1"
)

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
		Ports: []docker.Port{
			docker.Port("53/udp"),
			docker.Port("80/tcp"),
			docker.Port("443/tcp"),
		},
	}

	err = n.Create(config)
	c.Assert(err, IsNil)

	err = n.Delete(config)
	c.Assert(err, IsNil)
}
