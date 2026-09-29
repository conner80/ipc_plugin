// Package: ipc_plugin
// Description: Plugin's information item
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// PluginInfo plugin's the main information
type PluginInfo struct {
	key         string
	access_key  string
	path        string
	pid         int
	description string
	list        CommandsList
	is_alive    bool
	last_err    *Error
	host        *Host
	cmd         *exec.Cmd
	mutex       sync.Mutex
}

// NewPluginInfo creates new plugin's host
//
//	Parameters:
//
// * key [in] - Plugin's unique key/name
//
// * path [in] - Plugin's executable full filename
//
// * socket_path [in] - Plugin's UNIX-socket filename
//
// * connection_ti [in] - Plugin's connection timeout
//
// * read_to [in] - Read/Receive data from plugin timeout
//
// * write_to [in] - Write/Send data to plugin timeout
//
//	Returns:
//
// * Reference to the plugin's information
//
// * Plugin's error or nil if success
func NewPluginInfo(key, access_key, path, socket_path string, connection_to, read_to, write_to time.Duration) (*PluginInfo, *Error) {
	l := len(key)
	if l < 1 || l > 255 {
		return nil, ErrPluginKeyIncorrect
	}
	l = len(access_key)
	if l < 1 || l > 255 {
		return nil, ErrAccessKeyFormatError
	}
	inf, err_ := os.Stat(path)
	if err_ != nil {
		if errors.Is(err_, os.ErrNotExist) {
			return nil, ErrPluginFileNotFound
		}
		return nil, NewError(PluginPanicError, err_.Error())
	}
	if inf.IsDir() {
		return nil, ErrPluginFileNotFound
	}
	sock := socket_path
	var err *Error
	if sock == "" {
		sock, err = GetDefaultSocketPath()
		if err != nil {
			return nil, err
		}
	}

	hst, _ := NewHost(sock, connection_to, write_to, read_to, true)
	return &PluginInfo{
		key:        key,
		access_key: access_key,
		path:       path,
		host:       hst,
	}, nil
}

// CheckRunning check alive plugin status and update it
func (p *PluginInfo) CheckRunning() {
	// Check process existance
	p.mutex.Lock()
	pid := p.pid
	p.mutex.Unlock()
	ex, _ := IsProcessExists(pid)
	if !ex {
		p.mutex.Lock()
		p.is_alive = false
		p.mutex.Unlock()
		return
	}

	// check process freeze
	if err := p.host.Ping(); err != nil {
		p.Stop()
		return
	}

	p.mutex.Lock()
	p.is_alive = true
	p.mutex.Unlock()
}

// Stop stops plugin and kill plugin's process
func (p *PluginInfo) Stop() *Error {
	p.mutex.Lock()
	pid := p.pid
	key := p.access_key
	p.mutex.Unlock()
	if p.host != nil {
		pid_, err := p.host.Kill(key)
		if err == nil {
			pid = pid_
		}
	}
	for i := 0; i <= 5; i++ {
		ex, _ := IsProcessExists(pid)
		if !ex {
			p.mutex.Lock()
			p.cmd = nil
			p.pid = 0
			p.is_alive = false
			p.description = ""
			clear(p.list)
			p.mutex.Unlock()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if p.cmd != nil {
		p.cmd.Process.Signal(syscall.SIGKILL)
		waitDone := make(chan error, 1)
		func() {
			_, err_ := p.cmd.Process.Wait()
			waitDone <- err_
		}()

		select {
		case err_ := <-waitDone:
			if err_ != nil {
				log.Printf("[ERROR] Signal KILL error: %s", err_.Error())
			}
		case <-time.After(1 * time.Second):
			log.Printf("[WARNING] Plugin PID %d didn’t respond to the KILL command within 1 second", p.cmd.Process.Pid)
		}
	} else {
		err_ := KillProcess(pid)
		if err_ != nil {
			return NewError(PluginPanicError, err_.Error())
		}
	}
	p.mutex.Lock()
	p.cmd = nil
	p.pid = 0
	p.is_alive = false
	p.description = ""
	clear(p.list)
	fn := p.host.SocketPath
	p.mutex.Unlock()

	inf, err_ := os.Stat(fn)
	if err_ == nil {
		if !inf.IsDir() {
			os.Remove(fn)
		}
	}
	return nil
}

// Run start the plugin
//
//	Returns:
//
// * Plugin's error or nil if success
func (p *PluginInfo) Run() *Error {
	p.mutex.Lock()
	fn := p.path
	sock := p.host.SocketPath
	p.last_err = nil
	p.cmd = exec.Command(fn, sock)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr
	p.mutex.Unlock()

	err_ := p.cmd.Start()
	if err_ != nil {
		p.mutex.Lock()
		p.cmd = nil
		p.mutex.Unlock()
		return NewError(PluginPanicError, err_.Error())
	}

	for i := 0; i <= 10; i++ {
		_, err_ = os.Stat(sock)
		if err_ == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(200 * time.Microsecond)

	var err *Error
	var pid int
	var desc string
	var lst CommandsList
	pid, err = p.host.GetPid()
	if err != nil {
		if err.Type == PluginPanicError {
			p.mutex.Lock()
			p.last_err = err
			p.mutex.Unlock()
		}
		p.Stop()
		return err
	}
	desc, err = p.host.GetDescription()
	if err != nil {
		if err.Type == PluginPanicError {
			p.mutex.Lock()
			p.last_err = err
			p.mutex.Unlock()
		}
		p.Stop()
		return err
	}
	lst, err = p.host.GetList()
	if err != nil {
		if err.Type == PluginPanicError {
			p.mutex.Lock()
			p.last_err = err
			p.mutex.Unlock()
		}
		p.Stop()
		return err
	}

	p.mutex.Lock()
	p.pid = pid
	p.description = desc
	p.list = lst
	p.is_alive = true
	p.mutex.Unlock()
	return nil
}

// IsAlive plugin live status
//
//	Returns:
//
// * TRUE - plugin is alive, otherwise FALSE
func (p *PluginInfo) IsAlive() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.is_alive
}

// Call call user-defined specified method
//
//	Parameters:
//
// * method_name [in] - User-defined method name
//
// * req [in] - Method request parameter
//
// * resp [inout] - Method response
//
//	Returns:
//
// * Plugin's error or nil if success
func (p *PluginInfo) Call(method_name string, req interface{}, resp interface{}) *Error {
	if !p.IsAlive() {
		return ErrPluginStopped
	}
	return p.host.Call(method_name, req, resp)
}

// Description returns plugin's description text
//
//	Returns:
//
// * Plugin's description text
func (p *PluginInfo) Description() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.description
}

// List returns plugin's supported methos list
//
//	Returns:
//
// * Supported methods list
func (p *PluginInfo) List() CommandsList {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.list
}

// PID returns plugin's application PID
//
//	Returns:
//
// * Plugin's application PID
func (p *PluginInfo) PID() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.pid
}

// LastError returns plugin's critical last error
//
//	Returns:
//
// * Plugin's critical last error
func (p *PluginInfo) LastError() *Error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.last_err
}
