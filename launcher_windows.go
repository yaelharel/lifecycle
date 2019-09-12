package lifecycle

import (
	"fmt"
	"os"
	"os/exec"
)

func ExecW(argv0 string, argv []string, envv []string) error {
	cmd := exec.Command(argv0, argv[1:]...)
	fmt.Println("ARGS:", cmd.Args)
	cmd.Env = envv
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
	return nil
}
