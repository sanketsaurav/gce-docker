package providers

import (
	"encoding/base64"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/compute/v1"
	. "gopkg.in/check.v1"
)

var integration = flag.Bool("integration", false, "Include integration tests")

func Test(t *testing.T) { TestingT(t) }

type CommonSuite struct{}

var _ = Suite(&CommonSuite{})

type BaseSuite struct {
	key                     []byte
	project, zone, instance string

	c *http.Client
}

func (s *BaseSuite) SetUpSuite(c *C) {
	s.initEnviroment(c)
}

func (s *BaseSuite) SetUpTest(c *C) {
	ctx := context.Background()
	jwt, err := google.JWTConfigFromJSON(s.key, compute.ComputeScope)
	c.Assert(err, IsNil)

	s.c = oauth2.NewClient(ctx, jwt.TokenSource(ctx))
}

func (s *BaseSuite) initEnviroment(c *C) {
	s.project = os.Getenv("GCP_DEFAULT_PROJECT")
	s.zone = os.Getenv("GCP_DEFAULT_ZONE")
	s.instance = os.Getenv("GCP_DEFAULT_INSTANCE")

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *BaseSuite) getRandomName() string {
	return time.Now().Format("20060102150405000000")
}
