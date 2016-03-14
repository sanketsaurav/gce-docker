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
	s.instance = os.Getenv("GCP_DEFAULT_INSTANCE")

	var err error
	s.key, err = base64.StdEncoding.DecodeString(os.Getenv("GCP_JSON_KEY"))
	c.Assert(err, IsNil)
}

func (s *VolumeSuite) TestCreateRemoveAndList(c *C) {
	driver, err := NewVolume(s.c, s.project, s.zone, s.instance)
	c.Assert(err, IsNil)

	driver.Root = "/mnt/"

	name := s.getRandomName("create")

	//create disk, it not exists
	response := driver.Create(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")

	//ignore disk already exists
	response = driver.Create(volume.Request{Name: name})
	c.Assert(response.Err, Equals, "")

	//list disks
	response = driver.List(volume.Request{})
	c.Assert(response.Err, Equals, "")

	var found bool
	for _, v := range response.Volumes {
		if v.Name == name {
			found = true
			c.Assert(v.Mountpoint, Equals, "/mnt/"+name)
		}
	}

	c.Assert(found, Equals, true)

	//removes the created disk
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
	c.Assert(fs.Formatted, HasLen, 1)
	c.Assert(fs.Formatted[dev], Equals, "ext4")

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
	Mounted   map[string]string
	Formatted map[string]string
	afero.Fs
}

func NewMemFilesystem() *MemFilesystem {
	return &MemFilesystem{
		Mounted:   make(map[string]string, 0),
		Formatted: make(map[string]string, 0),

		Fs: afero.NewMemMapFs(),
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

func (fs *MemFilesystem) Format(source string) error {
	fs.Formatted[source] = "ext4"
	return nil
}

func (fs *MemFilesystem) WaitExists(path string, timeout time.Duration) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (fs *MemFilesystem) WaitNotExists(path string, timeout time.Duration) error {
	time.Sleep(10 * time.Second)
	return nil
}
