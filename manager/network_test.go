package manager

import (
	"encoding/base64"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"

	. "gopkg.in/check.v1"
)

type NetworkSuite struct {
	key                     []byte
	project, zone, instance string

	c *http.Client
}

var _ = Suite(&NetworkSuite{})

func (s *NetworkSuite) SetUpSuite(c *C) {
	s.initEnviroment(c)
}
func (s *NetworkSuite) SetUpTest(c *C) {
	ctx := context.Background()
	jwt, err := google.JWTConfigFromJSON(s.key, compute.ComputeScope)
	c.Assert(err, IsNil)

	s.c = oauth2.NewClient(ctx, jwt.TokenSource(ctx))
}

func (s *NetworkSuite) initEnviroment(c *C) {
	s.project = os.Getenv("GCP_DEFAULT_PROJECT")
	s.zone = os.Getenv("GCP_DEFAULT_ZONE")
	s.instance = os.Getenv("GCP_DEFAULT_INSTANCE")

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *NetworkSuite) TestCreate(c *C) {
	n, err := NewNetwork(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	config := &NetworkConfig{
		GroupName: "qux",
		Container: "foo",
		Protocol:  "upd",
		Port:      "53",
	}

	err = n.Create(config)
	c.Assert(err, IsNil)
}
