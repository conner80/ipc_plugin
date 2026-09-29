// Package: ipc_plugin
// Description: Plugin's errors
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

var (
	// Incorrect method name format error
	ErrMethodNameFormat = NewError(FormatError, "method name format error")
	// Incorrect method name length
	ErrMethodNameLength = NewError(FormatError, "incorrect method's name length")
	// Incorrect message type error
	ErrIncorrectMessageType = NewError(FormatError, "incorrect message type")
	// Incorrect message command
	ErrIncorrectMessageCommand = NewError(FormatError, "incorrect message command")
	// Incorrect error type
	ErrIncorrectErrorType = NewError(FormatError, "incorrect error type")
	// Plugin's access key format error
	ErrAccessKeyFormatError = NewError(FormatError, "plugin's access key format error")
	// Incorrect access key - access denied
	ErrIncorrectAccessKey = NewError(HandlerError, "incorrect access key")
	// Plugin's description length error
	ErrPluginDescriptionLengthError = NewError(FormatError, "plugin's description length error")
	// Message is empty (NIL)
	ErrMessageEmpty = NewError(PluginPanicError, "message is NIL")
	// Incorrect error response - no error object in the message
	ErrIncorrectErrorResponse = NewError(FormatError, "incorrect error response")
	// Plugin's description format error
	ErrPluginDescriptionError = NewError(FormatError, "plugin's description format error")
	// Create new plugin's socket error
	ErrPluginCreateSocketError = NewError(PluginPanicError, "can't create socket")
	// User's method not found error
	ErrPluginMethodNotFound = NewError(FormatError, "user's method not found")
	// Plugin's command format error
	ErrCommandFormatError = NewError(FormatError, "plugin's command format error")
	// User's method handler is not specified
	ErrNoHandler = NewError(PluginPanicError, "user's method handler is NIL")
	// Specified user's method name already registered
	ErrMethodAllreadyRegistered = NewError(PluginPanicError, "user's method name already registered")
	// Receive answer from specified plugin error
	ErrReceiveError = NewError(TransportError, "receive answer error")
	// Response format error
	ErrAnswerFormatError = NewError(FormatError, "answer/response format error")
	// Plugin's name/key is incorrect
	ErrPluginKeyIncorrect = NewError(FormatError, "plugin's name is incorrect")
	// Plugin with this key is already registered
	ErrPluginKeyAllreadyExists = NewError(FormatError, "plugin with this key is already registered")
	// Plugin executable file doesn't exists
	ErrPluginFileNotFound = NewError(PluginPanicError, "plugin's executable file is not found")
	// Plugin is stopped error
	ErrPluginStopped = NewError(PluginPanicError, "specified plugin is stopped")
	// Plugin with specified key already registered
	ErrPluginAllreadyRegistered = NewError(PluginPanicError, "plugin with specified key already registered")
	// Plugin doesn't support mandatory methods
	ErrPluginNotSupportedMandatoryMethods = NewError(PluginPanicError, "plugin doesn't support mandatory methods")
)
