// Package: ipc_plugin
// Description: Plugin's transport interface
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import "io"

// ITransport plugin's transport interface
type ITransport interface {
	// Send sends message to specified data channel
	//
	//	Parameters:
	//
	// * writer [inout] - Streaming writer with specified writing data channel
	//
	// * msg [in] - Message for transfer
	//
	//	Returns:
	//
	// * Transfered bytes count
	//
	// * Plugin's error or nil if success
	Send(writer io.Writer, msg *ProtocolMessage) (n int, err *Error)
	// Recv receive message from specified data channel
	//
	//	Parameters:
	//
	// * reader [in] - Streaming reader with specified reading data channel
	//
	// * msg [inout] - Message for transfer
	//
	//	Returns:
	//
	// * Transfered bytes count
	//
	// * Plugin's error or nil if success
	Recv(reader io.Reader, msg *ProtocolMessage) (n int, err *Error)
}

// NewTransport creates new transport object
//
//	Returns:
//
// * Created transport interface
func NewTransport() ITransport {
	return &StdTransport{}
}
