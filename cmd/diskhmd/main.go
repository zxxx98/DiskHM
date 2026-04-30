package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/example/diskhm/internal/app"
	"github.com/example/diskhm/internal/config"
)

const (
	defaultConfigPath  = "/etc/diskhm/config.yaml"
	serviceUnitName    = "diskhm.service"
	installedBinaryPath  = "/usr/local/bin/diskhm"
	installedServicePath = "/etc/systemd/system/diskhm.service"
	installedConfigDir   = "/etc/diskhm"
	installedDataDir     = "/var/lib/diskhm"
)

const (
	commandDaemon    = "daemon"
	commandStart     = "start"
	commandStop      = "stop"
	commandEnable    = "enable"
	commandDisable   = "disable"
	commandUninstall = "uninstall"
)

type cliCommand struct {
	Name       string
	ConfigPath string
}

type commandDeps struct {
	euid       func() int
	systemctl  func(...string) error
	removePath func(string) error
}

func main() {
	cmd, err := parseCommandArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("parse command: %v", err)
	}

	switch cmd.Name {
	case commandDaemon:
		if err := runDaemon(cmd.ConfigPath); err != nil {
			log.Fatal(err)
		}
	case commandStart, commandStop, commandEnable, commandDisable, commandUninstall:
		if err := runServiceCommand(cmd, defaultCommandDeps()); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command: %s", cmd.Name)
	}
}

func parseCommandArgs(args []string) (cliCommand, error) {
	if len(args) == 0 {
		return cliCommand{}, fmt.Errorf("expected one of: %s, %s, %s, %s, %s, %s", commandDaemon, commandStart, commandStop, commandEnable, commandDisable, commandUninstall)
	}

	switch args[0] {
	case commandDaemon:
		configPath, err := configPathFromArgs(flag.NewFlagSet("diskhm daemon", flag.ContinueOnError), defaultConfigPath, args[1:])
		if err != nil {
			return cliCommand{}, err
		}
		return cliCommand{Name: commandDaemon, ConfigPath: configPath}, nil
	case commandStart, commandStop, commandEnable, commandDisable, commandUninstall:
		if len(args) != 1 {
			return cliCommand{}, fmt.Errorf("%s does not accept extra arguments", args[0])
		}
		return cliCommand{Name: args[0]}, nil
	default:
		return cliCommand{}, fmt.Errorf("unknown command %q", args[0])
	}
}

func runDaemon(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	a := app.New(cfg)

	if err := http.ListenAndServe(cfg.Server.ListenAddr, a.Handler); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func configPathFromArgs(fs *flag.FlagSet, defaultPath string, args []string) (string, error) {
	configPath := defaultPath
	fs.StringVar(&configPath, "config", defaultPath, "path to config file")

	if err := fs.Parse(args); err != nil {
		return "", err
	}

	return configPath, nil
}

func runServiceCommand(cmd cliCommand, deps commandDeps) error {
	if deps.euid == nil {
		return fmt.Errorf("missing effective uid dependency")
	}
	if deps.euid() != 0 {
		return fmt.Errorf("command %q requires root; run it with sudo", cmd.Name)
	}
	if deps.systemctl == nil {
		return fmt.Errorf("missing systemctl dependency")
	}
	if deps.removePath == nil {
		return fmt.Errorf("missing removePath dependency")
	}

	switch cmd.Name {
	case commandStart:
		return deps.systemctl("start", serviceUnitName)
	case commandStop:
		return deps.systemctl("stop", serviceUnitName)
	case commandEnable:
		return deps.systemctl("enable", serviceUnitName)
	case commandDisable:
		return deps.systemctl("disable", serviceUnitName)
	case commandUninstall:
		if err := deps.systemctl("disable", "--now", serviceUnitName); err != nil {
			return err
		}
		for _, path := range []string{
			installedBinaryPath,
			installedServicePath,
			installedConfigDir,
			installedDataDir,
		} {
			if err := deps.removePath(path); err != nil {
				return err
			}
		}
		return deps.systemctl("daemon-reload")
	default:
		return fmt.Errorf("unsupported service command %q", cmd.Name)
	}
}

func defaultCommandDeps() commandDeps {
	return commandDeps{
		euid:       currentEUID,
		systemctl:  runSystemctl,
		removePath: os.RemoveAll,
	}
}

func runSystemctl(args ...string) error {
	return exec.Command("systemctl", args...).Run()
}

func currentEUID() int {
	if runtime.GOOS == "windows" {
		return -1
	}

	output, err := exec.Command("id", "-u").Output()
	if err != nil {
		return -1
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return -1
	}

	return value
}
