package main

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/BurntSushi/toml"

	"github.com/buildpack/lifecycle"
	"github.com/buildpack/lifecycle/cmd"
	"github.com/buildpack/lifecycle/metadata"
)

var (
	layersDir string
	appDir    string
)

func main() {
	cmd.Exit(launch())
}

func launch() error {
	defaultProcessType := cmd.DefaultProcessType
	if v := os.Getenv(cmd.EnvProcessType); v != "" {
		defaultProcessType = v
	} else if v := os.Getenv(cmd.EnvProcessTypeLegacy); v != "" {
		defaultProcessType = v
	}
	_ = os.Unsetenv(cmd.EnvProcessType)
	_ = os.Unsetenv(cmd.EnvProcessTypeLegacy)

	layersDir := cmd.DefaultLayersDir
	if v := os.Getenv(cmd.EnvLayersDir); v != "" {
		layersDir = v
	}
	_ = os.Unsetenv(cmd.EnvLayersDir)

	appDir := cmd.DefaultAppDir
	if v := os.Getenv(cmd.EnvAppDir); v != "" {
		appDir = v
	}
	_ = os.Unsetenv(cmd.EnvAppDir)

	var md lifecycle.BuildMetadata
	metadataPath := metadata.MetadataFilePath(layersDir)
	if _, err := toml.DecodeFile(metadataPath, &md); err != nil {
		return cmd.FailErr(err, "read metadata")
	}

	env := &lifecycle.Env{
		LookupEnv: os.LookupEnv,
		Getenv:    os.Getenv,
		Setenv:    os.Setenv,
		Unsetenv:  os.Unsetenv,
		Environ:   os.Environ,
		Map:       lifecycle.POSIXLaunchEnv,
	}

	execf := syscall.Exec
	if runtime.GOOS == "windows" {
		execf := func(argv0 string, argv []string, envv []string) (err error) {
			p, _ := syscall.GetCurrentProcess()
			fd := make([]syscall.Handle, 3)
			for i, file := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
				err := syscall.DuplicateHandle(p, syscall.Handle(file.Fd()), p, &fd[i], 0, true, syscall.DUPLICATE_SAME_ACCESS)
				if err != nil {
					return err
				}
				defer syscall.CloseHandle(syscall.Handle(fd[i]))
			}
			pid, handle, err := syscall.StartProcess(argv0, argv, &syscall.ProcAttr{
				Dir: appDir,
				Env: envv,
			})
			if err != nil {
				return err
			}
		}
	}
	launcher := &lifecycle.Launcher{
		DefaultProcessType: defaultProcessType,
		LayersDir:          layersDir,
		AppDir:             appDir,
		Processes:          md.Processes,
		Buildpacks:         md.Buildpacks,
		Env:                env,
		Exec:               execf,
	}

	if err := launcher.Launch(os.Args[0], os.Args[1:]); err != nil {
		return cmd.FailErrCode(err, cmd.CodeFailedLaunch, "launch")
	}
	return nil
}
