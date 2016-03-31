package plugin

import (
	"fmt"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/mcuadros/docker-volume-gce/providers"
	"github.com/spf13/afero"
	"google.golang.org/api/compute/v1"
	. "gopkg.in/check.v1"
)

const TimeoutAfterUnmount = 15 * time.Second

type VolumeSuite struct {
	v  *Volume
	fs *MemFilesystem
	p  *DiskProviderFixture
}

var _ = Suite(&VolumeSuite{})

func (s *VolumeSuite) SetUpTest(c *C) {
	s.fs = NewMemFilesystem()
	s.p = NewDiskProviderFixture()
	s.v = &Volume{p: s.p, fs: s.fs, Root: "/mnt/"}
}

func (s *VolumeSuite) TestCreateDiskConfig(c *C) {
	config, err := s.v.createDiskConfig(volume.Request{Name: "foo"})
	c.Assert(err, IsNil)
	c.Assert(config.Name, Equals, "foo")

	config, err = s.v.createDiskConfig(volume.Request{
		Name:    "foo",
		Options: map[string]string{"SizeGb": "42"},
	})
	c.Assert(err, IsNil)
	c.Assert(config.SizeGb, Equals, int64(42))

	config, err = s.v.createDiskConfig(volume.Request{
		Name:    "foo",
		Options: map[string]string{"Type": "foo"},
	})
	c.Assert(err, IsNil)
	c.Assert(config.Type, Equals, "foo")

	config, err = s.v.createDiskConfig(volume.Request{
		Name:    "foo",
		Options: map[string]string{"SourceSnapshot": "foo"},
	})
	c.Assert(err, IsNil)
	c.Assert(config.SourceSnapshot, Equals, "foo")

	config, err = s.v.createDiskConfig(volume.Request{
		Name:    "foo",
		Options: map[string]string{"SourceImage": "foo"},
	})
	c.Assert(err, IsNil)
	c.Assert(config.SourceImage, Equals, "foo")
}

func (s *VolumeSuite) TestCreate(c *C) {
	r := s.v.Create(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	c.Assert(s.p.disks, HasLen, 1)
	c.Assert(s.p.disks["foo"], Equals, true)
}

func (s *VolumeSuite) TestList(c *C) {
	r := s.v.Create(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	r = s.v.List(volume.Request{})
	c.Assert(r.Err, HasLen, 0)
	c.Assert(r.Volumes, HasLen, 1)
	c.Assert(r.Volumes[0].Name, Equals, "foo")
}

func (s *VolumeSuite) TestRemove(c *C) {
	r := s.v.Create(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	r = s.v.Remove(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	c.Assert(s.p.disks, HasLen, 0)
}

func (s *VolumeSuite) TestPath(c *C) {
	r := s.v.Path(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)
	c.Assert(r.Mountpoint, Equals, "/mnt/foo")

	fs, err := s.fs.Stat(r.Mountpoint)
	c.Assert(err, IsNil)
	c.Assert(fs.IsDir(), Equals, true)
}

func (s *VolumeSuite) TestMount(c *C) {
	r := s.v.Create(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	r = s.v.Mount(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)
	c.Assert(r.Mountpoint, Equals, "/mnt/foo")

	fs, err := s.fs.Stat(r.Mountpoint)
	c.Assert(err, IsNil)
	c.Assert(fs.IsDir(), Equals, true)

	c.Assert(s.p.attached, HasLen, 1)
	c.Assert(s.p.attached["foo"], Equals, true)
	c.Assert(s.fs.Mounted["/mnt/foo"], Equals, "/dev/disk/by-id/google-docker-volume-foo")
}

func (s *VolumeSuite) TestUnmount(c *C) {
	r := s.v.Create(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	r = s.v.Mount(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)
	c.Assert(r.Mountpoint, Equals, "/mnt/foo")

	r = s.v.Unmount(volume.Request{Name: "foo"})
	c.Assert(r.Err, HasLen, 0)

	c.Assert(s.p.attached, HasLen, 0)
	c.Assert(s.fs.Mounted["/mnt/foo"], Equals, "")
}

type DiskProviderFixture struct {
	disks    map[string]bool
	attached map[string]bool
}

func NewDiskProviderFixture() *DiskProviderFixture {
	return &DiskProviderFixture{
		disks:    make(map[string]bool, 0),
		attached: make(map[string]bool, 0),
	}
}

func (d *DiskProviderFixture) Create(c *providers.DiskConfig) error {
	d.disks[c.Name] = true
	return nil
}

func (d *DiskProviderFixture) Attach(c *providers.DiskConfig) error {
	if _, ok := d.disks[c.Name]; !ok {
		return fmt.Errorf("unable to find disk %s", c.Name)
	}

	d.attached[c.Name] = true
	return nil
}

func (d *DiskProviderFixture) Detach(c *providers.DiskConfig) error {
	delete(d.attached, c.Name)
	return nil
}

func (d *DiskProviderFixture) Delete(c *providers.DiskConfig) error {
	delete(d.disks, c.Name)
	return nil
}

func (d *DiskProviderFixture) List() ([]*compute.Disk, error) {
	var l []*compute.Disk
	for name, _ := range d.disks {
		l = append(l, &compute.Disk{Name: name, Status: "READY"})
	}

	l = append(l, &compute.Disk{Name: "no-ready", Status: "PENDING"})
	return l, nil
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
