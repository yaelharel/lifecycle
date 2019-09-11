package lifecycle

import (
	"os"
	"syscall"
)

func ExecW(argv0 string, argv []string, envv []string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	pid, handle, err := syscall.StartProcess(argv0, argv, &syscall.ProcAttr{
		Dir: dir,
		Env: envv,
		Files: []uintptr{
			uintptr(syscall.Stdin),
			uintptr(syscall.Stdout),
			uintptr(syscall.Stderr),
		},
	})
	syscall.WaitForSingleObject(handle, )
}
