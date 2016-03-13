package driver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/spf13/afero"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"

	. "gopkg.in/check.v1"
)

const TimeoutAfterUnmount = 15 * time.Second

type VolumeSuite struct {
	key                     []byte
	project, zone, instance string

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
	s.instance = "gke-mongo-cluster-605ad846-node-dlge"

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *VolumeSuite) TestCreateAndRemove(c *C) {
	driver, err := NewVolume(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	name := s.getRandomName("create")
	response := driver.Create(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")

	response = driver.Remove(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")
}

func (s *VolumeSuite) TestMountAndUnmount(c *C) {
	driver, err := NewVolume(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	fs := NewMemFilesystem()
	driver.fs = fs
	driver.Root = "/mnt/"

	name := s.getRandomName("create")
	dev := fmt.Sprintf(devPattern, fmt.Sprintf(DeviceNamePattern, name))
	mount := "/mnt/" + name

	response := driver.Create(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")

	response = driver.Mount(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")
	c.Assert(response.Mountpoint, Equals, mount)
	c.Assert(fs.Mounted, HasLen, 1)
	c.Assert(fs.Mounted[mount], Equals, dev)

	response = driver.Unmount(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")
	c.Assert(fs.Mounted, HasLen, 1)
	c.Assert(fs.Mounted[mount], Equals, "")

	time.Sleep(TimeoutAfterUnmount)
	response = driver.Remove(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")
}

func (s *VolumeSuite) TestMountInvalidPath(c *C) {
	driver, err := NewVolume(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	fs := NewMemFilesystem()
	driver.fs = fs
	driver.Root = "/mnt/"

	file, err := fs.Create("/mnt/foo")
	c.Assert(err, IsNil)
	_, err = file.WriteString("foo")
	c.Assert(err, IsNil)
	err = file.Close()
	c.Assert(err, IsNil)

	response := driver.Mount(volume.Request{Name: "foo"})
	c.Assert(response.Err, Equals, `error the mountpoint "/mnt/foo" already exists`)
	c.Assert(response.Mountpoint, Equals, "")
	c.Assert(fs.Mounted, HasLen, 0)
}

func (s *VolumeSuite) getRandomName(name string) string {
	return fmt.Sprintf(
		"test-dv-gce-%s-%s",
		time.Now().Format("20060102150405000000"), name,
	)
}

type MemFilesystem struct {
	Mounted map[string]string
	afero.Fs
}

func NewMemFilesystem() *MemFilesystem {
	return &MemFilesystem{
		Mounted: make(map[string]string, 0),
		Fs:      afero.NewMemMapFs(),
	}
}

func (fs *MemFilesystem) Mount(source string, target string) error {
	fs.Mounted[target] = source
	return nil
}

func (fs *MemFilesystem) Unmount(target string) error {
	fs.Mounted[target] = ""
	return nil
}
