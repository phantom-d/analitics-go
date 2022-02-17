package daemons

import (
	"analitics/pkg/config"
	"context"
	"os"
	"syscall"
	"time"
)

// Default file permissions for log and pid files.
const FILE_PERM = os.FileMode(0640)

type Daemon interface {
	Run() error
	Data() *DaemonData
	SetData(*DaemonData)
	MakeDaemon()
	Terminate(os.Signal)
}

type DaemonData struct {
	Name        string                 `mapstructure:"Name"`
	MemoryLimit uint64                 `mapstructure:"MemoryLimit"`
	Workers     []config.Worker        `mapstructure:"Workers"`
	Params      map[string]interface{} `mapstructure:"Params"`
	Sleep       time.Duration          `mapstructure:"Sleep"`
	Context     *Context
	ctx         context.Context
	signalChan  chan os.Signal
	done        chan struct{}
}

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
	// If Chroot is non-empty, the child changes root directory
	Chroot string

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
	pidFile *config.LockFile

	rpipe *os.File
}

type Factory map[string]func() Daemon

var factory = make(Factory)

func init() {
	factory.Register("watcher", func() Daemon { return &Watcher{} })
	factory.Register("import", func() Daemon { return &Import{} })
	factory.Register("status", func() Daemon { return &Status{} })
}

func (factory *Factory) Register(name string, factoryFunc func() Daemon) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) Daemon {
	return (*factory)[name]()
}
