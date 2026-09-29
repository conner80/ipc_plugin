// Package: ipc_plugin
// Description: Plugin's error structure
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

// ErrorType plugin's error type
type ErrorType uint8

const (
	// PluginPanicError critical plugin error
	PluginPanicError ErrorType = iota
	// HandlerError plugin's method handler internal error (not PANIC)
	HandlerError
	// FormatError request or response format error
	FormatError
	// TransportError exhange between host and plugin error
	TransportError
)

// IsValid validate error type
func (e ErrorType) IsValid() bool {
	return e == PluginPanicError || e == HandlerError || e == FormatError || e == TransportError
}

// Error plugin's error structure
type Error struct {
	// Type plugin error's type
	Type ErrorType
	// ErrorText error text message
	ErrorText string
}

// NewError create new error
//
//	Description:
//
// Create new plugin's error and returns pointer to the created error object.
//
//	Parameters:
//
// * err_type [in] - Plugin's error type
//
// * err_text [in] - Plugin's error text message
//
//	Returns:
//
// * Pointer to the created plugin's error
func NewError(err_type ErrorType, err_text string) *Error {
	return &Error{
		Type:      err_type,
		ErrorText: err_text,
	}
}

// Error get error text string
func (e *Error) Error() string {
	return e.ErrorText
}

// Is compare error
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Type == t.Type && e.ErrorText == t.ErrorText
}
