// Package: ipc_plugin
// Description: Plugin's preprocessor structure
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
	"unicode/utf8"

	"log"
)

// PluginHanlder plugin's user handler preprocessor
//
//	Parameters:
//
// * req [in] - Request message to plugin
//
// * resp [inout] - Response message from plugin
//
//	Returns:
//
// * Internal user's handler error
type PluginHanlder func(*ProtocolMessage, *ProtocolMessage) error

// Plugin plugin's preprocessor structure
//
//	Description:
//
// The server-side part of the plugin (the plugin core) that ensures the execution of commands from the host.
// The entire operation mechanism is built around IPC technology based on the UNIX-sockets.
type Plugin struct {
	// SocketPath socket's filepath
	//
	//	example:
	//
	// /var/run/plugin.sock
	SocketPath string
	// Description plugin's information/description
	//
	//	Warning:
	//
	// Maximum length 5000 characters
	Description string
	// AccessKey the secret key for performing KILL command
	AccessKey string
	mutex     sync.Mutex
	methods   map[string]PluginHanlder
	listener  net.Listener
	done      chan struct{}
	transp    ITransport
}

// NewPlugin creates new plugin's object and start internal listener for accept commands from host
//
//	Parameters:
//
// * socket_path [in] - Full filename for UNIX-socket file. If empty - set default UNIX path for the file.
//
// * description [in] - Plugin's description
//
//	Returns:
//
// * Pointer to the created plugin's command preprocessor
//
// * Plugin's error
func NewPlugin(socket_path string, access_key string, description string) (*Plugin, *Error) {
	l := utf8.RuneCountInString(description)
	if l > 5000 {
		return nil, ErrPluginDescriptionError
	}
	l = len(access_key)
	if l < 1 || l > 255 {
		return nil, ErrAccessKeyFormatError
	}
	fn := socket_path
	var err *Error
	if fn == "" {
		fn, err = GetDefaultSocketPath()
		if err != nil {
			return nil, err
		}
	}

	_ = os.Remove(fn)
	listener, err_ := net.Listen("unix", fn)
	if err_ != nil {
		return nil, NewError(PluginPanicError, err_.Error())
	}

	plg := Plugin{
		SocketPath:  fn,
		Description: description,
		AccessKey:   access_key,
		listener:    listener,
		methods:     make(map[string]PluginHanlder),
		done:        make(chan struct{}),
		transp:      NewTransport(),
	}

	go plg.run()
	return &plg, nil
}

// run start plugin's listener goroutine
func (p *Plugin) run() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case _, ok := <-p.done:
				if !ok {
					return
				}
			default:
			}

			continue
		}

		go p.handleConnection(conn)
	}
}

// sendError send error answer to the host
func (p *Plugin) sendError(conn net.Conn, req *ProtocolMessage, err *Error) {
	resp := ProtocolMessage{
		Type:    ResponseErrorType,
		Command: req.Command,
		Error:   err,
	}
	p.transp.Send(conn, &resp)
}

// call safe user's handler call
func (p *Plugin) call(handler PluginHanlder, req *ProtocolMessage, resp *ProtocolMessage) (err_result *Error) {
	defer func() {
		if r := recover(); r != nil {
			err_result = NewError(PluginPanicError, fmt.Sprintf("%v", r))
		}
	}()

	if handler == nil {
		return ErrNoHandler
	}
	if req == nil {
		return ErrMessageEmpty
	}
	if resp == nil {
		resp = &ProtocolMessage{
			Type:    ResponseSuccessType,
			Command: req.Command,
			Method:  req.Method,
		}
	}
	err := handler(req, resp)
	if err != nil {
		return NewError(HandlerError, err.Error())
	}

	return nil
}

// setList sets methods list to the protocol message
func (p *Plugin) setList(msg *ProtocolMessage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	msg.List = make(CommandsList, len(p.methods))
	var i int = 0
	for key := range p.methods {
		msg.List[i] = key
		i++
	}
}

// handleConnection socket's connection main handler
func (p *Plugin) handleConnection(conn net.Conn) {
	defer conn.Close()

	req := ProtocolMessage{}
	_, err := p.transp.Recv(conn, &req)
	if err != nil {
		log.Printf("plugin receive data error: %s", err.ErrorText)
		return
	}

	err = req.Validate()
	if err != nil {
		log.Printf("plugin validate error: %s", err.ErrorText)
		return
	}

	if req.Type != RequestType {
		log.Print("request type error")
		p.sendError(conn, &req, ErrIncorrectMessageType)
		return
	}

	resp := ProtocolMessage{
		Type:    ResponseSuccessType,
		Command: req.Command,
	}
	switch req.Command {
	case PidCommand:
		resp.PID = os.Getpid()

	case ListCommand:
		p.setList(&resp)

	case InfoCommand:
		resp.Description = p.Description

	case KillCommand:
		if req.AccessKey != p.AccessKey {
			p.sendError(conn, &req, ErrIncorrectAccessKey)
			return
		}
		resp.PID = os.Getpid()
		p.transp.Send(conn, &resp)
		p.Done()
		os.Exit(0)
		return

	case UserCommand:
		if req.Method == "" {
			p.sendError(conn, &req, ErrMethodNameLength)
			return
		}

		p.mutex.Lock()
		handler, ok := p.methods[req.Method]
		if !ok {
			p.mutex.Unlock()
			p.sendError(conn, &req, ErrPluginMethodNotFound)
			return
		}
		p.mutex.Unlock()
		err = p.call(handler, &req, &resp)
		if err != nil {
			p.sendError(conn, &req, err)
			return
		}
	}

	p.transp.Send(conn, &resp)
}

// Use register new user's method and handler
//
//	Description:
//
// Register new user's method and handler. Rules for method name like and identifier name.
//
//	Parameter:
//
// * method_name [in] - The name of the user's plugin method
//
// * handler [in] - Pointer to the method's handler function
//
//	Returns:
//
// * Plugin error or nil if success
func (p *Plugin) Use(method_name string, handler PluginHanlder) *Error {
	var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
	l := len(method_name)
	if l < 1 || l > 255 {
		return ErrMethodNameLength
	}

	if !validName.MatchString(method_name) {
		return ErrMethodNameFormat
	}

	if handler == nil {
		return ErrNoHandler
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, ok := p.methods[method_name]
	if ok {
		return ErrMethodAllreadyRegistered
	}

	p.methods[method_name] = handler
	return nil
}

// Done finalization plugin object and reset all methods to zero
func (p *Plugin) Done() {
	close(p.done)
	p.listener.Close()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	clear(p.methods)
	os.Remove(p.SocketPath)
}
