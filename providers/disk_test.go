package providers

import . "gopkg.in/check.v1"

type DiskSuite struct {
	BaseSuite
}

var _ = Suite(&DiskSuite{})

func (s *DiskSuite) TestCreate(c *C) {
	if !*integration {
		c.Skip("-integration not provided")
	}

	n, err := NewDisk(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	config := &DiskConfig{
		Name: "test",
	}

	err = n.Create(config)
	c.Assert(err, IsNil)

	err = n.Attach(config)
	c.Assert(err, IsNil)

	err = n.Detach(config)
	c.Assert(err, IsNil)

	err = n.Delete(config)
	c.Assert(err, IsNil)
}
