package driver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
	"google.golang.org/api/compute/v1"
)

const (
	WaitStatusTimeout             = 30 * time.Second
	CreatingDiskStatus DiskStatus = "CREATING"
	FailedStatus       DiskStatus = "FAILED"
	ReadyStatus        DiskStatus = "READY"
	RestoringStatus    DiskStatus = "RESTORING"
)

type Volume struct {
	s       *compute.Service
	zone    string
	project string
}

func NewVolume(c *http.Client, project, zone string) (*Volume, error) {
	s, err := compute.New(c)
	if err != nil {
		return nil, err
	}

	return &Volume{
		s:       s,
		project: project,
		zone:    zone,
	}, nil
}

//https://godoc.org/google.golang.org/api/compute/v1#Disk
func (v *Volume) Create(r volume.Request) volume.Response {
	d := &compute.Disk{}
	if err := applyOptions(r.Options, d); err != nil {
		return buildReponseError(err)
	}

	_, err := v.s.Disks.Insert(v.project, v.zone, d).Do()
	if err != nil {
		return buildReponseError(err)
	}

	if err := v.waitStatus(d.Name, ReadyStatus); err != nil {
		return buildReponseError(err)
	}

	return volume.Response{}
}

func (v *Volume) waitStatus(disk string, status DiskStatus) error {
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

func (v *Volume) List(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Get(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Remove(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Path(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Mount(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func (v *Volume) Unmount(volume.Request) volume.Response {
	return volume.Response{Err: "not implemented"}
}

func buildReponseError(err error) volume.Response {
	return volume.Response{Err: err.Error()}
}
