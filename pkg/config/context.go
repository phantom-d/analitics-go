package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// Default file permissions for log and pid files.
const FilePerm = os.FileMode(0640)

// A Context describes daemon context.
type Context struct {
	// If PidFileName is non-empty, parent process will try to create and lock
	// pid file with given name. Child process writes process id to file.
	PidFileName string
	// Permissions for new pid file.
	PidFilePerm os.FileMode

	// If WorkDir is non-empty, the child changes into the directory before
	// creating the process.
	WorkDir string

	// If Env is non-nil, it gives the environment variables for the
	// daemon-process in the form returned by os.Environ.
	// If it is nil, the result of os.Environ will be used.
	Env []string
	// If Args is non-nil, it gives the command-line args for the
	// daemon-process. If it is nil, the result of os.Args will be used.
	Args []string

	// Credential holds user and group identities to be assumed by a daemon-process.
	Credential *syscall.Credential
	// If Umask is non-zero, the daemon-process call Umask() func with given value.
	Umask int

	// Struct contains only serializable public fields (!!!)
	abspath string
	pidFile *LockFile

	rpipe *os.File
}

// Search searches daemons process by given in context pid file name.
// If success returns pointer on daemons os.Process structure,
// else returns error. Returns nil if filename is empty.
func (d *Context) Search() (daemon *os.Process, err error) {
	if len(d.PidFileName) > 0 {
		var pid int
		if _, err = os.Stat(d.PidFileName); err == nil {
			if pid, err = ReadPidFile(d.PidFileName); err != nil {
				return
			}
			daemon, err = os.FindProcess(pid)
		} else if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
	}
	return
}

// Release provides correct pid-file release in daemon.
func (d *Context) Release() (err error) {
	if d.pidFile != nil {
		err = d.pidFile.Remove()
	}
	return
}

func (d *Context) Run() (child *os.Process, err error) {
	if err = d.prepareEnv(); err != nil {
		return
	}

	defer d.closeFiles()
	if err = d.openFiles(); err != nil {
		return
	}

	attr := &os.ProcAttr{
		Dir:   d.WorkDir,
		Env:   d.Env,
		Files: d.files(),
		Sys: &syscall.SysProcAttr{
			//Chroot:     d.Chroot,
			Credential: d.Credential,
			Setsid:     true,
		},
	}

	d.Args = append([]string{d.abspath}, d.Args...)
	if child, err = os.StartProcess(d.abspath, d.Args, attr); err != nil {
		if d.pidFile != nil {
			_ = d.pidFile.Remove()
		}
		return
	}
	return
}

func (d *Context) CreatePidFile() (err error) {
	if len(d.PidFileName) > 0 {
		if d.PidFilePerm == 0 {
			d.PidFilePerm = FilePerm
		}
		if d.PidFileName, err = filepath.Abs(d.PidFileName); err != nil {
			return
		}
		if d.pidFile, err = CreatePidFile(d.PidFileName, d.PidFilePerm); err != nil {
			return
		}
	}
	return
}

func (d *Context) openFiles() (err error) {
	d.rpipe, _, err = os.Pipe()
	return
}

func (d *Context) closeFiles() (err error) {
	if d.pidFile != nil {
		_ = d.pidFile.Close()
		d.pidFile = nil
	}
	return
}

func (d *Context) prepareEnv() (err error) {
	if d.abspath, err = os.Executable(); err != nil {
		return
	}

	if len(d.Args) == 0 {
		d.Args = os.Args
	}

	if len(d.Env) == 0 {
		d.Env = os.Environ()
	}
	return
}

func (d *Context) files() (f []*os.File) {
	f = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	if d.pidFile != nil {
		f = append(f, d.pidFile.File)
	}
	return
}
