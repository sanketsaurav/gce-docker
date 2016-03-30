package plugin

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/spf13/afero"
)

var (
	DefaultFStype       = "ext4"
	DefaultMountOptions = []string{"discard", "defaults"}
	HostFilesystem      = "/rootfs/"
	MountNamespace      = "/rootfs/proc/1/ns/mnt"
	CGroupFilename      = "/proc/1/cgroup"
)

type Filesystem interface {
	afero.Fs
	Mount(source string, target string) error
	Unmount(target string) error
	Format(source string) error
}

type OSFilesystem struct {
	inContainer bool
	afero.Fs
}

func NewFilesystem() Filesystem {
	fs := afero.NewOsFs()

	inContainer := inContainer()

	if inContainer {
		log15.Info("running inside of container")
		fs = afero.NewBasePathFs(fs, HostFilesystem)
	}

	return &OSFilesystem{inContainer: inContainer, Fs: fs}
}

var nsenterArgs = []string{
	"nsenter",
	fmt.Sprintf("--mount=%s", MountNamespace),
	"--",
}

func (fs *OSFilesystem) Mount(source string, target string) error {
	args := fs.getMountArgs(source, target, DefaultFStype, DefaultMountOptions)

	command := exec.Command(args[0], args[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"mount failed, arguments: %q\noutput: %s\n",
			args, string(output),
		)
	}

	return err
}

func (fs *OSFilesystem) getMountArgs(source, target, fstype string, options []string) []string {
	var args []string
	args = append(args, "mount")

	if len(fstype) > 0 {
		args = append(args, "-t", fstype)
	}

	if len(options) > 0 {
		args = append(args, "-o", strings.Join(options, ","))
	}

	args = append(args, source)
	args = append(args, target)

	if fs.inContainer {
		return append(nsenterArgs, args...)
	}

	return args
}

func (fs *OSFilesystem) Unmount(target string) error {
	args := fs.getUnmountArgs(target)

	command := exec.Command(args[0], args[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"unmount failed, arguments: %q\noutput: %s\n",
			args, string(output),
		)
	}

	return nil
}

func (fs *OSFilesystem) getUnmountArgs(target string) []string {
	var args []string
	args = append(args, "umount", target)

	if fs.inContainer {
		return append(nsenterArgs, args...)
	}

	return args
}

func (fs *OSFilesystem) Format(source string) error {
	if fs.isFormatted(source) {
		return nil
	}

	args := fs.getMkfsExt4Args(source)
	command := exec.Command(args[0], args[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"mkfs.ext4 failed, arguments: %q\noutput: %s\n",
			args, string(output),
		)
	}

	return nil
}

func (fs *OSFilesystem) getMkfsExt4Args(source string) []string {
	var args []string
	args = append(args, "mkfs.ext4", source)

	if fs.inContainer {
		return append(nsenterArgs, args...)
	}

	return args
}

func (fs *OSFilesystem) isFormatted(source string) bool {
	args := fs.getBlkidArgs(source)

	command := exec.Command(args[0], args[1:]...)
	_, err := command.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func (fs *OSFilesystem) getBlkidArgs(source string) []string {
	var args []string
	args = append(args, "blkid", source)

	if fs.inContainer {
		return append(nsenterArgs, args...)
	}

	return args
}

func inContainer() bool {
	content, err := ioutil.ReadFile(CGroupFilename)
	if err != nil {
		return false
	}

	for _, l := range strings.Split(string(content), "\n") {
		p := strings.Split(l, ":")
		if len(p) != 3 {
			continue
		}

		if strings.TrimSpace(p[2]) != "/" {
			return true
		}
	}

	return false
}
