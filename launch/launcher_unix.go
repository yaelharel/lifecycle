//+build linux darwin

package launch

import "syscall"

const (
	CNBDir      = `/cnb`
	exe         = ""
	profileGlob = "*"
	appProfile  = ".profile"
)

var (
	OSExecFunc   = syscall.Exec
	DefaultShell = &BashShell{Exec: OSExecFunc}
)
