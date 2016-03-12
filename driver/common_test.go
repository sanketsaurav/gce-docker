package driver

import (
	"testing"

	"google.golang.org/api/compute/v1"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type CommonSuite struct{}

var _ = Suite(&CommonSuite{})

func (s *CommonSuite) TestApplyOptionsUnknown(c *C) {
	target := &compute.Disk{}
	opts := map[string]string{"foo": "qux"}

	err := applyOptions(opts, target)
	c.Assert(err.Error(), Equals, `unknown property "foo" at "compute.Disk"`)
}

func (s *CommonSuite) TestApplyOptionsTypeMissmatch(c *C) {
	target := &compute.Disk{}
	opts := map[string]string{"SizeGb": "foo"}

	err := applyOptions(opts, target)
	c.Assert(err, Not(IsNil))
}

func (s *CommonSuite) TestApplyOptionsInt(c *C) {
	target := &compute.Disk{}
	opts := map[string]string{"SizeGb": "42"}

	err := applyOptions(opts, target)
	c.Assert(err, IsNil)
	c.Assert(target.SizeGb, Equals, int64(42))
}
