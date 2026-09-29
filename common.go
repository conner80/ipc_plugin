// Package: ipc_plugin
// Description: Common functions, types and constants
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
	"time"
)

// GetDefaultSocketPath get UNIX-socket default filepath for the most UNIX-like OS.
//
//	Returns:
//
// * UNIX-socket filename
//
// * Plugin's error or nil if success
func GetDefaultSocketPath() (string, *Error) {
	p_path, err := os.Executable()
	if err != nil {
		return "", ErrPluginCreateSocketError
	}
	p_fn := filepath.Base(p_path)
	s_fn := p_fn + ".sock"

	dirs := []string{filepath.Join("/run", p_fn), "/run", filepath.Join("/var/run", p_fn), "/var/run", filepath.Join("/tmp", p_fn), "/tmp"}
	for _, dir := range dirs {
		fn := filepath.Join(dir, s_fn)
		file, err := os.Create(fn)
		if err != nil {
			continue
		}
		file.Close()
		return fn, nil
	}

	return "", ErrPluginCreateSocketError
}

// GetDefaultSocketDir returns directory for UNIX-sockets
//
//	Returns:
//
// * UNIX-socket default directory
func GetDefaultSocketDir() string {
	p_path, err := os.Executable()
	if err != nil {
		return ""
	}
	p_fn := filepath.Base(p_path)
	s_fn := p_fn + ".tmp"
	dirs := []string{filepath.Join("/run", p_fn), "/run", filepath.Join("/var/run", p_fn), "/var/run", filepath.Join("/tmp", p_fn), "/tmp"}
	for _, dir := range dirs {
		fn := filepath.Join(dir, s_fn)
		file, err := os.Create(fn)
		if err != nil {
			continue
		}
		file.Close()
		return dir
	}

	return ""
}

// IsValidMethodName validate user's method name
//
//	Parameters:
//
// * name [in] - Method name
//
//	Returns:
//
// * Plugin's error or nil if success
func IsValidMethodName(name string) *Error {
	var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
	l := len(name)
	if l < 1 || l > 255 {
		return ErrMethodNameLength
	}

	if !validName.MatchString(name) {
		return ErrMethodNameFormat
	}

	return nil
}

// IsProcessExists checks existance the process by PID
//
//	Parameters:
//
// * pid [in] - PID of a process
//
//	Returns:
//
// * TRUE - process is running at the moment, otherwise FALSE
//
// * Check operation error or nil if success
func IsProcessExists(pid int) (bool, error) {
	if pid <= 0 {
		return false, os.ErrInvalid
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if runtime.GOOS == "windows" {
		return true, nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrPermission) {
		return true, nil
	}

	return false, nil
}

// KillProcess kill/stop specified process by PID
//
//	Parameters:
//
// * pid [in] - PID of a process
//
//	Returns:
//
// * Kill process operation error or nil if success
func KillProcess(pid int) error {
	if pid <= 0 {
		return os.ErrInvalid
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = proc.Kill()
	if err != nil {
		return err
	}
	waitDone := make(chan error, 1)
	go func() {
		_, err_ := proc.Wait()
		waitDone <- err_
	}()

	select {
	case err_ := <-waitDone:
		if err_ != nil {
			log.Printf("[ERROR] Signal KILL error: %s", err_.Error())
		}
	case <-time.After(1 * time.Second):
		log.Printf("[WARNING] Plugin PID %d didn’t respond to the KILL command within 1 second", pid)
	}

	return nil
}
