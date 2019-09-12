package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	layerDir := filepath.Join(os.Args[1], "fake-layer")
	if err := os.Mkdir(layerDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	layerConfigFile, err := os.Create(filepath.Join(os.Args[1], "fake-layer.toml"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	defer layerConfigFile.Close()
	_, err = layerConfigFile.Write([]byte(`
launch = true
build = true
	`))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
	contentFile, err := os.Create(filepath.Join(layerDir, "file.txt"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(4)
	}
	defer contentFile.Close()
	_, err = contentFile.Write([]byte("contents"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(5)
	}

	if err := os.Mkdir(filepath.Join(layerDir, "env.launch"), 0755); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(6)
	}
	envFile, err := os.Create(filepath.Join(layerDir, "env.launch", "MY_KEY"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(7)
	}
	defer envFile.Close()
	_, err = envFile.Write([]byte("MY_VAL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(8)
	}

	launchConfigFile, err := os.Create(filepath.Join(os.Args[1], "launch.toml"))
	if err != nil {
		os.Exit(9)
	}
	defer launchConfigFile.Close()
	_ , err = launchConfigFile.Write([]byte(`
[[processes]]
type = "web"
command = "set"
direct = false
`))
	if err != nil {
		os.Exit(6)
	}

}
