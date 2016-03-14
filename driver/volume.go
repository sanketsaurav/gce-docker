package driver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/docker/go-plugins-helpers/volume"
	"google.golang.org/api/compute/v1"
)

const (
	CreatingStatus  Status = "CREATING"
	FailedStatus    Status = "FAILED"
	ReadyStatus     Status = "READY"
	RestoringStatus Status = "RESTORING"

	devPattern       = "/dev/disk/by-id/google-%s"
	sourceURLPattern = "projects/%s/zones/%s/disks/%s"
)

var (
	DeviceNamePattern = "docker-volume-%s"
	WaitStatusTimeout = 30 * time.Second
)

type Volume struct {
	Root string

	s        *compute.Service
	fs       Filesystem
	zone     string
	project  string
	instance string
}

func NewVolume(c *http.Client, project, zone, instance string) (*Volume, error) {
	s, err := compute.New(c)
	if err != nil {
		return nil, err
	}

	return &Volume{
		s:        s,
		fs:       NewFilesystem(),
		project:  project,
		zone:     zone,
		instance: instance,
	}, nil
}

func (v *Volume) Create(r volume.Request) volume.Response {
	log15.Info("create request received", "name", r.Name)

	exists, err := v.diskExists(r.Name)
	if err != nil {
		return buildReponseError(err)
	}

	if exists {
		return volume.Response{}
	}

	d := &compute.Disk{Name: r.Name}
	if err := applyOptions(r.Options, d); err != nil {
		return buildReponseError(err)
	}

	if _, err := v.s.Disks.Insert(v.project, v.zone, d).Do(); err != nil {
		return buildReponseError(err)
	}

	if err := v.waitStatus(d.Name, ReadyStatus); err != nil {
		return buildReponseError(err)
	}

	return volume.Response{}
}

func (v *Volume) List(volume.Request) volume.Response {
	log15.Info("list request received")

	disks, err := v.listAvailableDisks()
	if err != nil {
		return buildReponseError(err)
	}

	r := volume.Response{}
	for _, d := range disks {
		if !ReadyStatus.Equals(d.Status) {
			continue
		}

		r.Volumes = append(r.Volumes, &volume.Volume{
			Name:       d.Name,
			Mountpoint: v.mountPoint(d.Name),
		})
	}

	return r
}

func (v *Volume) Get(volume.Request) volume.Response {
	log15.Info("get request received")

	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Remove(r volume.Request) volume.Response {
	log15.Info("remove request received", "name", r.Name)

	_, err := v.s.Disks.Delete(v.project, v.zone, r.Name).Do()
	if err != nil {
		return buildReponseError(err)
	}

	return volume.Response{}
}

func (v *Volume) Path(r volume.Request) volume.Response {
	mnt := v.mountPoint(r.Name)
	log15.Info("path request received", "name", r.Name, "mnt", mnt)

	if err := v.createMountPoint(r.Name); err != nil {
		return buildReponseError(err)
	}

	return volume.Response{Mountpoint: mnt}
}

func (v *Volume) Mount(r volume.Request) volume.Response {
	log15.Info("mount request received", "name", r.Name)

	if err := v.createMountPoint(r.Name); err != nil {
		return buildReponseError(err)
	}

	if err := v.attachDisk(r.Name); err != nil {
		return buildReponseError(err)
	}

	if err := v.formatAttachedDisk(r.Name); err != nil {
		return buildReponseError(err)
	}

	if err := v.mountAttachedDisk(r.Name); err != nil {
		return buildReponseError(err)
	}

	return volume.Response{
		Mountpoint: v.mountPoint(r.Name),
	}
}

func (v *Volume) diskExists(name string) (bool, error) {
	disks, err := v.listAvailableDisks()
	if err != nil {
		return false, err
	}

	for _, d := range disks {
		if d.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (v *Volume) listAvailableDisks() ([]*compute.Disk, error) {
	op, err := v.s.Disks.List(v.project, v.zone).Do()
	if err != nil {
		return nil, err
	}

	return op.Items, err
}

func (v *Volume) createMountPoint(name string) error {
	target := v.mountPoint(name)
	fi, err := v.fs.Stat(target)
	if os.IsNotExist(err) {
		return v.fs.MkdirAll(target, 0755)
	}

	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("error the mountpoint %q already exists", target)
	}

	return nil
}

func (v *Volume) attachDisk(name string) error {
	d := &compute.AttachedDisk{
		Source:     v.sourceURL(name),
		DeviceName: v.deviceName(name),
	}

	_, err := v.s.Instances.AttachDisk(v.project, v.zone, v.instance, d).Do()
	if err != nil {
		return err
	}

	dev := fmt.Sprintf(devPattern, d.DeviceName)
	if err := v.fs.WaitExists(dev, WaitStatusTimeout); err != nil {
		return err
	}

	//wait to become available, still need a seconds
	time.Sleep(300 * time.Millisecond)

	return nil
}

func (v *Volume) formatAttachedDisk(name string) error {
	disk, err := v.getAttachedDisk(name)
	if err != nil {
		return err
	}

	source := fmt.Sprintf(devPattern, disk.DeviceName)
	return v.fs.Format(source)
}

func (v *Volume) mountAttachedDisk(name string) error {
	disk, err := v.getAttachedDisk(name)
	if err != nil {
		return err
	}

	source := fmt.Sprintf(devPattern, disk.DeviceName)
	target := v.mountPoint(name)

	return v.fs.Mount(source, target)
}

func (v *Volume) Unmount(r volume.Request) volume.Response {
	log15.Info("unmount request received", "name", r.Name)

	if err := v.unmountAttachedDisk(r.Name); err != nil {
		return buildReponseError(err)
	}

	if err := v.detachDisk(r.Name); err != nil {
		return buildReponseError(err)
	}

	return volume.Response{}
}

func (v *Volume) unmountAttachedDisk(name string) error {
	target := v.mountPoint(name)
	return v.fs.Unmount(target)
}

func (v *Volume) detachDisk(name string) error {
	disk, err := v.getAttachedDisk(name)
	if err != nil {
		return err
	}

	if disk == nil {
		return fmt.Errorf("device %q is not attached to %q", name, v.instance)
	}

	_, err = v.s.Instances.DetachDisk(v.project, v.zone, v.instance, disk.DeviceName).Do()
	if err != nil {
		return err
	}

	dev := fmt.Sprintf(devPattern, disk.DeviceName)
	if err := v.fs.WaitNotExists(dev, WaitStatusTimeout); err != nil {
		return err
	}

	return nil
}

func (v *Volume) isAttachedDisk(name string) (bool, error) {
	disk, err := v.getAttachedDisk(name)
	if err != nil {
		return false, err
	}

	if disk == nil {
		return false, nil
	}

	return true, nil
}

func (v *Volume) getAttachedDisk(name string) (*compute.AttachedDisk, error) {
	i, err := v.s.Instances.Get(v.project, v.zone, v.instance).Do()
	if err != nil {
		return nil, err
	}

	for _, d := range i.Disks {
		if strings.HasSuffix(d.Source, v.sourceURL(name)) {
			return d, nil
		}
	}

	return nil, nil
}

func (v *Volume) waitStatus(disk string, status Status) error {
	c := time.Tick(500 * time.Millisecond)
	start := time.Now()

	for range c {
		op, err := v.s.Disks.Get(v.project, v.zone, disk).Do()
		if err != nil {
			return err
		}

		if status.Equals(op.Status) {
			return nil
		}

		if time.Since(start) > WaitStatusTimeout {
			return fmt.Errorf("Timeout exceeded waiting status %q", status)
		}
	}

	return nil
}

func (v *Volume) mountPoint(name string) string {
	return filepath.Join(v.Root, name)
}

func (v *Volume) deviceName(name string) string {
	return fmt.Sprintf(DeviceNamePattern, name)
}

func (v *Volume) sourceURL(name string) string {
	return fmt.Sprintf(sourceURLPattern, v.project, v.zone, name)
}

func buildReponseError(err error) volume.Response {
	log15.Error("request failed", "error", err.Error())
	return volume.Response{Err: err.Error()}
}
