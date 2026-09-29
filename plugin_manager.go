// Package: ipc_plugin
// Description: Plugins manager
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"fmt"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// PluginManager Plugins manager
type PluginManager struct {
	methods  CommandsList
	sock_dir string
	list     map[string]*PluginInfo
	mutex    sync.Mutex
	done     chan struct{}
}

// NewPluginManager creates new plugin manager
//
//	Parameters:
//
// * socket_directory [in] - plugin's UNIX-sockets default directory
//
//	Returns:
//
// * Reference to the created plugin manager
func NewPluginManager(socket_directory string, methods CommandsList) *PluginManager {
	dir := socket_directory
	if socket_directory == "" {
		dir = GetDefaultSocketDir()
	}
	m := &PluginManager{
		sock_dir: dir,
		methods:  methods,
		list:     make(map[string]*PluginInfo),
		done:     make(chan struct{}),
	}

	go m.run()

	return m
}

// checkPlugins checks plugin's activity
func (p *PluginManager) checkPlugins() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, plg := range p.list {
		if plg != nil {
			plg.CheckRunning()
		}
	}
}

// run starts plugin's checker loop
func (p *PluginManager) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkPlugins()
		case <-p.done:
			return
		}
	}
}

// Done finalization plugins manager and stop all goroutines
func (p *PluginManager) Done() {
	close(p.done)
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, plg := range p.list {
		plg.Stop()
	}

	clear(p.list)
}

// LoadPlugin loads specified plugin file and checks supported methods by plugin
func (p *PluginManager) LoadPlugin(key, access_key, filename string, connect_to, read_to, write_to time.Duration) *Error {
	p_fn := filepath.Base(filename)
	fn := filepath.Join(p.sock_dir, fmt.Sprintf("spsp_%s.sock", p_fn))
	plg, err := NewPluginInfo(key, access_key, filename, fn, connect_to, read_to, write_to)
	if err != nil {
		return err
	}

	err = plg.Run()
	if err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, ok := p.list[key]
	if ok {
		return ErrPluginAllreadyRegistered
	}

	lst := plg.List()
	if len(p.methods) > len(lst) {
		return ErrPluginNotSupportedMandatoryMethods
	}

	var cnt int = 0
	for _, s := range p.methods {
		if slices.Contains(lst, s) {
			cnt++
		}
	}
	if cnt != len(p.methods) {
		return ErrPluginNotSupportedMandatoryMethods
	}

	p.list[key] = plg
	return nil
}

// UnloadPlugin unloads specified plugin by plugin's key
//
//	Parameters:
//
// * key [in] - Plugin's key
func (p *PluginManager) UnloadPlugin(key string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	plg, ok := p.list[key]
	if !ok {
		return
	}

	plg.Stop()
	delete(p.list, key)
}

// GetPlugin Returns reference to specified plugin by plugin's key
//
//	Parameters:
//
// * key [in] - Plugin's specified unique key
//
//	Returns:
//
// * Reference to the plugin controller or nil if not found
func (p *PluginManager) GetPlugin(key string) *PluginInfo {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	plg, ok := p.list[key]
	if !ok {
		return nil
	}

	return plg
}
