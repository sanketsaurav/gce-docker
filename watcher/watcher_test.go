package watcher

import (
	"fmt"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IPManagerSuite struct{}

var _ = Suite(&IPManagerSuite{})

func (s *IPManagerSuite) AATestStart(c *C) {
	m, err := NewWatcher()
	c.Assert(err, IsNil)

	err = m.Watch()
	c.Assert(err, IsNil)

	fmt.Println(err)

}
