// Package: ipc_plugin
// Description: Exchange protocol and constants
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"encoding/json"
	"unicode/utf8"
)

// PluginCommand plugin's main command
type PluginCommand uint8

const (
	// PingCommand PING command for plugin
	PingCommand PluginCommand = iota
	// PidCommand get plugin's executable process PID
	PidCommand
	// ListCommand the list of supported user's methods names
	ListCommand
	// InfoCommand get plugin's information
	InfoCommand
	// Shutdown plugin process
	KillCommand
	// UserCommand execute specified user's command
	UserCommand
)

// IsValid validate plugin's command
func (c PluginCommand) IsValid() bool {
	return c == PingCommand || c == PidCommand || c == ListCommand || c == InfoCommand || c == UserCommand || c == KillCommand
}

// String get plugin's command as string
func (c PluginCommand) String() string {
	switch c {
	case PingCommand:
		return "PING"
	case PidCommand:
		return "PID"
	case ListCommand:
		return "LIST"
	case InfoCommand:
		return "INFO"
	case UserCommand:
		return "USER"
	default:
		return ""
	}
}

// MessageType the type of message
type MessageType uint8

const (
	// RequestType request message type
	RequestType MessageType = iota
	// ResponseSuccessType response success message type
	ResponseSuccessType
	// ResponseErrorType response error message type
	ResponseErrorType
)

// IsValid validate plugin's message type
func (t MessageType) IsValid() bool {
	return t == RequestType || t == ResponseSuccessType || t == ResponseErrorType
}

// String get plugin's message type as string
func (t MessageType) String() string {
	switch t {
	case RequestType:
		return "REQUEST"
	case ResponseSuccessType:
		return "RESPONSE OK"
	case ResponseErrorType:
		return "RESPONSE ERR"
	default:
		return ""
	}
}

// CommandsList the list of supported commands of the plugin
type CommandsList []string

// ProtocolMessage a message between the Host and the Plugin containing commands and data.
//
// # Exchange protocol
//
//	Header:
//
// All commands (request and response) have protocol's header.
//
//	Offset | Length | Description
//
// ------------------------------------------------------------
//
//	0x0000  | 1      | Message type: Request or Response
//	0x0001  | 1      | Command's type: Ping, List or User
//	0x0002  | Dyn    | Body (optional)
//
//	Body:
//
// For error answers (for all commands) sends header and plugin's error message
//
//	Offset | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 1      | Plugin's error type
//	0x0003  | 1      | Error text length
//	0x0004  | 1-255  | Error text message
//
//	For other succeful commands body structrue below
//
// # 1. PING command
//
// Check plugin prensents (check IsAlive) command.
//
//		Request:
//
//	 * No body - only header
//
//		Response:
//
//	 * No body - only header
//
// # 2. PID command
//
// Request plugin's executable process PID.
//
//		Request:
//
//	* No body - only header
//
//		Response:
//
//	Offset  | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 4      | Plugin's executable process PID
//
// # 3. LIST command
//
// Request supported user's methods list.
//
//		Request:
//
//	* No body - only header
//
//		Response:
//
//	Offset  | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 4      | Method's list length in bytes
//	0x0006  | Dyn    | Message list
//
// # 4. INFO command
//
// Request plugin's information/description.
//
//		Request:
//
//	* No body - only header.
//
//		Response:
//
//	Offset  | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 4      | Information text length
//	0x0006  | 0-5000 | Information text
//
// # 5. KILL command
//
// Graceful shutdown plugin's process.
//
//		Request:
//
//	Offset  | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 1      | Access key string
//	0x0003  | 1-255  | Access key string
//
//		Response:
//
//	Offset  | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002  | 4      | Plugin's PID
//
// # 6. USER command
//
// Call specified user's method.
//
//		Request:
//
//	 Offset       | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002        | 4      | Message body length in bytes
//	0x0006        | 1      | Method name length in bytes
//	0x0007        | 1-255  | Method name
//	0x0008-0x0106 | Dyn    | User data block
//
//		Response:
//
//	 Offset       | Length | Description
//
// ------------------------------------------------------------
//
//	0x0002        | 4      | Message body length in bytes
//	0x0006        | 1      | Method name length in bytes
//	0x0007        | 1-255  | Method name
//	0x0008-0x0106 | Dyn    | User data block
type ProtocolMessage struct {
	// Type message type
	//
	// - 0 (RequestType) - Request message
	//
	// - 1 (ResponseSuccessType) - Response success message with neccessary answer data
	//
	// - 2 (ResponseErrorType) - Response error message with error type and text
	Type MessageType
	// Command plugin's command
	//
	// - 0 (PingCommand) - Ping command to the plugin's application and check alive status
	//
	// - 1 (PidCommand) - Request plugin's PID
	//
	// - 2 (ListCommand) - Request plugin's methods list
	//
	// - 3 (InfoCommand) - Request plugin description
	//
	// - 4 (KillCommand) - Graceful shutdown plugin process
	//
	// - 5 (UserCommand) - Send user command
	Command PluginCommand
	// Method user's command method name
	//
	//	Description:
	// The same principles apply to the naming of methods as to identifiers (variables) in programming languages;
	// the length of the name must be in the range from 1 to 255 characters.
	Method string
	// Error error for error response
	Error *Error
	// PID plugin's PID
	PID int
	// Description plugin's description
	Description string
	// List Supported methods list
	List CommandsList
	// AccessKey plugin's access key for performing KILL command
	AccessKey string
	// Data user's data
	Data json.RawMessage
}

// Clear clear protocol message data
func (m *ProtocolMessage) Clear() {
	clear(m.Data)
	clear(m.List)
	m.Error = nil
	m.Method = ""
	m.PID = 0
	m.Description = ""
	m.AccessKey = ""
}

// Validate checking the correctness of the message content.
//
//	Returns:
//
// * Plugin's error or nil if success
func (m *ProtocolMessage) Validate() *Error {
	if !m.Type.IsValid() {
		return ErrIncorrectMessageType
	}
	if !m.Command.IsValid() {
		return ErrIncorrectMessageCommand
	}
	if m.Type == ResponseErrorType {
		if m.Error == nil {
			return ErrIncorrectErrorResponse
		}
		if !m.Error.Type.IsValid() {
			return ErrIncorrectErrorType
		}
	}

	switch m.Type {
	case ResponseSuccessType:
		switch m.Command {
		case InfoCommand:
			l := utf8.RuneCountInString(m.Description)
			if l > 5000 {
				return ErrPluginDescriptionLengthError
			}
		case UserCommand:
			if err := IsValidMethodName(m.Method); err != nil {
				return err
			}
		}
	case RequestType:
		switch m.Command {
		case KillCommand:
			l := len(m.AccessKey)
			if l < 1 || l > 255 {
				return ErrAccessKeyFormatError
			}
		case UserCommand:
			if err := IsValidMethodName(m.Method); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetData returns user's object from message data
//
//	Parameters:
//
// * obj [out] - Pointer to the user's object variable
//
//	Returns:
//
// * Plugin's error or nil if success
func (m *ProtocolMessage) GetData(obj interface{}) *Error {
	err := json.Unmarshal(m.Data, obj)
	if err != nil {
		return NewError(FormatError, err.Error())
	}
	return nil
}

// SetData sets user's data to the message
//
//	Parameters:
//
// * obj [in] - User's object variable
//
//	Returns:
//
// * Plugin's error orr nil if success
func (m *ProtocolMessage) SetData(obj interface{}) *Error {
	clear(m.Data)
	var err error
	m.Data, err = json.Marshal(obj)
	if err != nil {
		return NewError(FormatError, err.Error())
	}

	return nil
}
