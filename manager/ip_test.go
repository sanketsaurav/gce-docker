package manager

import (
	"fmt"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IPManagerSuite struct{}

var _ = Suite(&IPManagerSuite{})

func (s *IPManagerSuite) AATestStart(c *C) {
	m, err := NewIPManager()
	c.Assert(err, IsNil)

	err = m.Start()
	c.Assert(err, IsNil)

	fmt.Println(err)

}
