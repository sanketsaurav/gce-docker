package driver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"

	. "gopkg.in/check.v1"
)

type VolumeSuite struct {
	key           []byte
	project, zone string

	c *http.Client
}

var _ = Suite(&VolumeSuite{})

func (s *VolumeSuite) SetUpSuite(c *C) {
	s.initEnviroment(c)
}
func (s *VolumeSuite) SetUpTest(c *C) {
	ctx := context.Background()
	jwt, err := google.JWTConfigFromJSON(s.key, compute.ComputeScope)
	c.Assert(err, IsNil)

	s.c = oauth2.NewClient(ctx, jwt.TokenSource(ctx))
}

func (s *VolumeSuite) initEnviroment(c *C) {
	s.project = os.Getenv("GCP_DEFAULT_PROJECT")
	s.zone = os.Getenv("GCP_DEFAULT_ZONE")

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *VolumeSuite) TestCreate(c *C) {
	driver, err := NewVolume(s.c, s.project, s.zone)
	c.Assert(err, IsNil)

	response := driver.Create(volume.Request{
		Options: map[string]string{"Name": s.getRandomName("create")},
	})

	c.Assert(response.Err, Equals, "")
}

func (s *VolumeSuite) getRandomName(name string) string {
	return fmt.Sprintf(
		"test-dv-gce-%s-%s",
		time.Now().Format("20060102150405"), name,
	)
}
