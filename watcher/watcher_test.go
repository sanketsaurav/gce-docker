package watcher

import (
	"encoding/base64"
	"net/http"
	"os"

	"github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"

	. "gopkg.in/check.v1"
)

type WatcherSuite struct {
	key                     []byte
	project, zone, instance string

	c *http.Client
}

var _ = Suite(&WatcherSuite{})

func (s *WatcherSuite) SetUpSuite(c *C) {
	s.initEnviroment(c)
}

func (s *WatcherSuite) SetUpTest(c *C) {
	ctx := context.Background()
	jwt, err := google.JWTConfigFromJSON(s.key, compute.ComputeScope)
	c.Assert(err, IsNil)

	s.c = oauth2.NewClient(ctx, jwt.TokenSource(ctx))
}

func (s *WatcherSuite) initEnviroment(c *C) {
	s.project = os.Getenv("GCP_DEFAULT_PROJECT")
	s.zone = os.Getenv("GCP_DEFAULT_ZONE")
	s.instance = os.Getenv("GCP_DEFAULT_INSTANCE")

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *WatcherSuite) TestStart(c *C) {
	c.Skip("playground")
	client, err := docker.NewClientFromEnv()
	c.Assert(err, IsNil)

	w, err := NewWatcher(client, s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	container := &docker.Container{
		ID: "abcdefghijklm",
		NetworkSettings: &docker.NetworkSettings{
			Ports: map[docker.Port][]docker.PortBinding{
				docker.Port("80/tcp"): []docker.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "80"},
				},
				docker.Port("443/tcp"): []docker.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "443"},
				},
				docker.Port("53/udp"): []docker.PortBinding{
					{HostIP: "127.0.0.1", HostPort: "53"},
				},
			},
		},
	}

	w.createNetworkConfig(container, map[string]string{
		"gce.network.group": "foo",
	})

	err = w.Watch()
	c.Assert(err, IsNil)
}
