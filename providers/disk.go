package providers

import (
	"net/http"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type DiskProvider interface {
	Create(c *DiskConfig) error
	Attach(c *DiskConfig) error
	Detach(c *DiskConfig) error
	Delete(c *DiskConfig) error
	List() ([]*compute.Disk, error)
}

type Disk struct {
	Client
}

func NewDisk(c *http.Client, project, zone, instance string) (*Disk, error) {
	client, err := NewClient(c, project, zone, instance)
	if err != nil {
		return nil, err
	}

	return &Disk{Client: *client}, nil
}

func (d *Disk) Create(c *DiskConfig) error {
	disk := c.Disk()
	if _, err := d.s.Disks.Get(d.project, d.zone, disk.Name).Do(); err != nil {
		if apiErr, ok := err.(*googleapi.Error); !ok || apiErr.Code != 404 {
			return err
		}

		op, err := d.s.Disks.Insert(d.project, d.zone, disk).Do()
		if err != nil {
			return err
		}

		return d.WaitDone(op)
	}

	return nil
}

func (d *Disk) Attach(c *DiskConfig) error {
	ad := &compute.AttachedDisk{
		Source:     DiskURL(d.project, d.zone, c.Name),
		DeviceName: c.DeviceName(),
	}

	op, err := d.s.Instances.AttachDisk(d.project, d.zone, d.instance, ad).Do()
	if err != nil {
		return err
	}

	return d.WaitDone(op)
}

func (d *Disk) Detach(c *DiskConfig) error {
	op, err := d.s.Instances.DetachDisk(d.project, d.zone, d.instance, c.DeviceName()).Do()
	if err != nil {
		return err
	}

	return d.WaitDone(op)
}

func (d *Disk) Delete(c *DiskConfig) error {
	op, err := d.s.Disks.Delete(d.project, d.zone, c.Name).Do()
	if err != nil {
		return err
	}

	return d.WaitDone(op)
}

func (d *Disk) List() ([]*compute.Disk, error) {
	op, err := d.s.Disks.List(d.project, d.zone).Do()
	if err != nil {
		return nil, err
	}

	return op.Items, err
}
