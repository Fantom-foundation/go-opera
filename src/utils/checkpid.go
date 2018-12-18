package utils

import (
	"fmt"
	"os"
	"syscall"
	"github.com/facebookgo/pidfile"
)

func CheckPid() error {
	pidfile.SetPidfilePath("/tmp/lachesis.pid")
	pid, err := pidfile.Read()
	if err == nil && pid > 0 {
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("Failed to find process: %v", err)
		} else {
			err := process.Signal(syscall.Signal(0))
			if err == nil {
				return fmt.Errorf("Perhaps another lachesis is already running with pid %d", pid)
			}
		}
	}

	if err := pidfile.Write(); err != nil {
		return fmt.Errorf("Error writing into pidfile: %v", err)
	}
	
	return nil
}
