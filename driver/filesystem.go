package driver

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/afero"
)

var (
	DefaultFStype       = "ext4"
	DefaultMountOptions = []string{"discard", "defaults"}
)

type Filesystem interface {
	afero.Fs
	Mount(source string, target string) error
	Unmount(target string) error
}

type OSFilesystem struct {
	afero.Fs
}

func NewFilesystem() Filesystem {
	return &OSFilesystem{Fs: afero.NewOsFs()}
}

func (fs *OSFilesystem) Mount(source string, target string) error {
	args := fs.getMountArgs(source, target, DefaultFStype, DefaultMountOptions)

	command := exec.Command("mount", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"mount failed, arguments: %q\noutput: %s\n",
			args, string(output),
		)
	}

	return err
}

//$ sudo mount -o discard,defaults DISK_LOCATION MOUNT_POINT
func (fs *OSFilesystem) getMountArgs(source, target, fstype string, options []string) []string {
	var args []string
	if len(fstype) > 0 {
		args = append(args, "-t", fstype)
	}

	if len(options) > 0 {
		args = append(args, "-o", strings.Join(options, ","))
	}

	args = append(args, source)
	args = append(args, target)

	return args
}

func (fs *OSFilesystem) Unmount(target string) error {
	command := exec.Command("umount", target)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"unmount failed, arguments: %q\noutput: %s\n",
			target, string(output),
		)
	}

	return nil
}
