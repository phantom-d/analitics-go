package daemons

import (
	"analitics/pkg/config"
	"os"
	"path/filepath"
	"syscall"
)

// Search searches daemons process by given in context pid file name.
// If success returns pointer on daemons os.Process structure,
// else returns error. Returns nil if filename is empty.
func (d *Context) Search() (daemon *os.Process, err error) {
	if len(d.PidFileName) > 0 {
		var pid int
		if pid, err = config.ReadPidFile(d.PidFileName); err != nil {
			return
		}
		daemon, err = os.FindProcess(pid)
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

	if child, err = os.StartProcess(d.abspath, d.Args, attr); err != nil {
		if d.pidFile != nil {
			_ = d.pidFile.Remove()
		}
		return
	}
	return
}

func (d *Context) openFiles() (err error) {
	if d.PidFilePerm == 0 {
		d.PidFilePerm = FILE_PERM
	}

	if len(d.PidFileName) > 0 {
		if d.PidFileName, err = filepath.Abs(d.PidFileName); err != nil {
			return err
		}
		if d.pidFile, err = config.OpenLockFile(d.PidFileName, d.PidFilePerm); err != nil {
			return
		}
		if err = d.pidFile.Lock(); err != nil {
			return
		}
		if len(d.Chroot) > 0 {
			// Calculate PID-file absolute path in child's environment
			if d.PidFileName, err = filepath.Rel(d.Chroot, d.PidFileName); err != nil {
				return err
			}
			d.PidFileName = "/" + d.PidFileName
		}
	}
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
