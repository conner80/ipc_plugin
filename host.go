// Package: ipc_plugin
// Description: Plugin's host preprocessor
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"net"
	"os"
	"time"
)

// Host plugin's host preprocessor
//
//	Description:
//
// The client-side part of the plugin (the plugin host) that ensures the sends of commands to the plugin core.
// The entire operation mechanism is built around IPC technology based on the UNIX-sockets.
type Host struct {
	// SocketPath socket's filepath
	//
	//	example:
	//
	// /var/run/plugin.sock
	SocketPath string
	// ConnectionTimeout plugin connection timeout
	ConnectionTimeout time.Duration
	// SendTimeout send request to plugin timeout
	SendTimeout time.Duration
	// RecvTimeout receive response from plugin timeout
	RecvTimeout time.Duration
	transp      ITransport
}

// NewHost creates new plugin's host object
//
//	Parameters:
//
// * socket_path [in] - Full filename for UNIX-socket file. If empty - set default UNIX path for the file.
//
// * conn_timeout [in] - Plugin connection timeout
//
// * send_timeout [in] - Send request to plugin timeout
//
// * recv_timeout [in] - Receive response from plugin timeout
//
//	Returns:
//
// * Pointer to the created plugin's host object
//
// * Plugin's error
func NewHost(socket_path string, conn_timeout, send_timeout, recv_timeout time.Duration, no_check_param bool) (*Host, *Error) {
	fn := socket_path
	if !no_check_param {
		var err1 *Error
		if fn == "" {
			fn, err1 = GetDefaultSocketPath()
			if err1 != nil {
				return nil, err1
			}
		}
		st, err := os.Stat(socket_path)
		if err != nil {
			return nil, NewError(TransportError, err.Error())
		}
		if st.IsDir() {
			return nil, NewError(TransportError, "socket path is DIRECTORY")
		}
		conn, err := net.DialTimeout("unix", socket_path, conn_timeout)
		if err != nil {
			return nil, NewError(TransportError, err.Error())
		}
		conn.Close()
	}

	return &Host{
		SocketPath:        fn,
		transp:            NewTransport(),
		ConnectionTimeout: conn_timeout,
		SendTimeout:       send_timeout,
		RecvTimeout:       recv_timeout,
	}, nil
}

// request request to specififed plugin
func (h *Host) request(req *ProtocolMessage, resp *ProtocolMessage) *Error {
	conn, err := net.Dial("unix", h.SocketPath)
	if err != nil {
		return NewError(TransportError, err.Error())
	}
	defer conn.Close()

	err = conn.SetWriteDeadline(time.Now().Add(h.SendTimeout))
	if err != nil {
		return NewError(TransportError, err.Error())
	}

	_, err_ := h.transp.Send(conn, req)
	if err_ != nil {
		conn.SetWriteDeadline(time.Time{})
		return err_
	}
	conn.SetWriteDeadline(time.Time{})

	err = conn.SetReadDeadline(time.Now().Add(h.RecvTimeout))
	if err != nil {
		return NewError(TransportError, err.Error())
	}

	_, err_ = h.transp.Recv(conn, resp)
	if err_ != nil {
		conn.SetReadDeadline(time.Time{})
		return err_
	}
	conn.SetReadDeadline(time.Time{})

	if resp.Command != req.Command {
		return ErrIncorrectMessageCommand
	}

	if resp.Type == ResponseErrorType {
		return resp.Error
	} else if resp.Type == RequestType || resp.Command != req.Command {
		return ErrAnswerFormatError
	}

	return nil
}

// Ping check plugin exists and alive
//
//	Returns:
//
// * Plugin's error or nil if success
func (h *Host) Ping() *Error {
	req := ProtocolMessage{
		Type:    RequestType,
		Command: PingCommand,
	}
	resp := ProtocolMessage{}

	return h.request(&req, &resp)
}

// GetList returns supported methods list by specified plugin
//
//	Returns:
//
// * Supported methods names list
//
// * Plugin's error or nil if success
func (h *Host) GetList() (CommandsList, *Error) {
	req := ProtocolMessage{
		Type:    RequestType,
		Command: ListCommand,
	}
	resp := ProtocolMessage{}

	err := h.request(&req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.List, nil
}

// GetPid returns plugin's process PID
//
//	Returns:
//
// * Plugin's process PID
//
// * Plugin's error or nil if success
func (h *Host) GetPid() (int, *Error) {
	req := ProtocolMessage{
		Type:    RequestType,
		Command: PidCommand,
	}
	resp := ProtocolMessage{}

	err := h.request(&req, &resp)
	if err != nil {
		return 0, err
	}

	return resp.PID, nil
}

// GetDescription returns plugin's description text
//
//	Returns:
//
// * Plugin's description text
//
// * Plugin's error or nil if success
func (h *Host) GetDescription() (string, *Error) {
	req := ProtocolMessage{
		Type:    RequestType,
		Command: InfoCommand,
	}
	resp := ProtocolMessage{}

	err := h.request(&req, &resp)
	if err != nil {
		return "", err
	}
	return resp.Description, nil
}

// Kill graceful shutdown specified plugin
//
//	Parameters:
//
// * access_key [in] - Secret access key for perform KILL command
//
//	Returns:
//
// * Plugin's application process PID
//
// * Plugin's error or nil if success
func (h *Host) Kill(access_key string) (int, *Error) {
	req := ProtocolMessage{
		Type:      RequestType,
		Command:   KillCommand,
		AccessKey: access_key,
	}
	resp := ProtocolMessage{}

	err := h.request(&req, &resp)
	if err != nil {
		return 0, err
	}

	return resp.PID, nil
}

// Call perform specified method
//
//	Parameters:
//
// * method_name [in] - Plugin's method name
//
// * req_param [in] - Request parameter for method
//
// * resp_result [inout] - Response result
//
//	Returns:
//
// * Plugin's error or nil if success
func (h *Host) Call(method_name string, req_param interface{}, resp_result interface{}) *Error {
	req := ProtocolMessage{
		Type:    RequestType,
		Command: UserCommand,
		Method:  method_name,
	}
	if req_param != nil {
		req.SetData(req_param)
	}
	resp := ProtocolMessage{}

	err := h.request(&req, &resp)
	if err != nil {
		return err
	}

	return resp.GetData(resp_result)
}
