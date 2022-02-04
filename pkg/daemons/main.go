package daemons

import (
	"analitics/pkg/config"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sevlyar/go-daemon"
)

func New(name string) Daemon {
	if cfg, ok := config.Application.Daemons[name]; ok {
		if cfg.Enabled {
			cfg.Name = name
			d := factory.CreateInstance(name)
			dd := &DaemonData{}
			err := mapstructure.Decode(cfg, &dd)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
				return nil
			}

			if err := dd.checkBinary(); err != nil {
				config.Logger.Error().Err(err).Msg("")
				return nil
			}
			d.SetData(dd)
			return d
		} else {
			config.Logger.Debug().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		config.Logger.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

func (dd *DaemonData) Start(d Daemon, daemonize bool) {
	if daemonize {

	}
	daemon.AddCommand(daemon.StringFlag(&config.Application.Signal, "quit"), syscall.SIGQUIT, dd.termHandler)
	daemon.AddCommand(daemon.StringFlag(&config.Application.Signal, "stop"), syscall.SIGTERM, dd.termHandler)
	dd.stop = make(chan struct{})
	dd.done = make(chan struct{})

	cntxt := &daemon.Context{
		PidFileName: fmt.Sprintf("%s/%s.pid", config.Application.PidDir, dd.Name),
		PidFilePerm: 0644,
		LogFileName: "/dev/stdout",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[--daemon=" + dd.Name + "]"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			config.Logger.Fatal().Err(err).Msg("Unable send signal to the daemon")
		}
		if err = daemon.SendCommands(d); err != nil {
			config.Logger.Error().Err(err).Msg("Error send signal to the daemon")
		}
		return
	}

	dm, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if dm != nil {
		return
	}
	defer cntxt.Release()

	config.Logger.Info().Msgf("Start daemon '%s'!", dd.Name)

	go dd.worker()

	err = daemon.ServeSignals()
	if err != nil {
		config.Logger.Error().Err(err).Msg("")
	}

	config.Logger.Info().Msgf("daemon '%s' terminated", dd.Name)
	d.Run()
}

func (dd *DaemonData) checkBinary() error {
	//get path to binary and confirm its writable
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find binary path (%s)", err)
	}
	dd.binPath = binPath
	if info, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("failed to stat binary (%s)", err)
	} else if info.Size() == 0 {
		return fmt.Errorf("binary file is empty")
	} else {
		//copy permissions
		dd.binPerms = info.Mode()
	}
	f, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("cannot read binary (%s)", err)
	}
	//initial hash of file
	hash := sha1.New()
	_, _ = io.Copy(hash, f)
	dd.binHash = hash.Sum(nil)
	_ = f.Close()
	return nil
}

func (dd *DaemonData) termHandler(sig os.Signal) error {
	config.Logger.Info().Msgf("daemon '%s' terminating...", dd.Name)
	dd.stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-dd.done
	}
	return daemon.ErrStop
}

func (dd *DaemonData) worker() {
LOOP:
	for {
		// this is work to be done by worker.
		time.Sleep(time.Second)
		select {
		case <-dd.stop:
			break LOOP
		default:
		}
	}
	dd.done <- struct{}{}
}
