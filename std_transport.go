// Package: ipc_plugin
// Description: Plugin's default transport structure
// Author: John Doe
// Version: 1.0.0
package ipc_plugin

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

// StdTransport plugin's default transport structure
type StdTransport struct {
}

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
func (t *StdTransport) Send(writer io.Writer, msg *ProtocolMessage) (n int, err *Error) {
	if msg == nil {
		return 0, ErrMessageEmpty
	}
	if !msg.Type.IsValid() {
		return 0, ErrIncorrectMessageType
	}
	if !msg.Command.IsValid() {
		return 0, ErrIncorrectMessageCommand
	}
	var cnt int = 0
	var header []byte

	// response error message
	if msg.Type == ResponseErrorType {
		if msg.Error == nil {
			return 0, ErrCommandFormatError
		}
		s := msg.Error.ErrorText
		l := len(s)
		if l > 255 {
			s = s[:255]
			l = 255
		}
		header = make([]byte, l+4)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		header[2] = uint8(msg.Error.Type)
		header[3] = uint8(l)
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i
		if l != 0 {
			i, err_ = writer.Write([]byte(s))
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			cnt += i
		}
		return cnt, nil
	}

	// other responses
	switch msg.Command {
	case PingCommand:
		header = make([]byte, 2)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i

	case PidCommand:
		if msg.Type == RequestType {
			header = make([]byte, 2)
			header[0] = uint8(msg.Type)
			header[1] = uint8(msg.Command)
			i, err_ := writer.Write(header)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			cnt += i
			return cnt, nil
		}
		header = make([]byte, 6)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		binary.BigEndian.PutUint32(header[2:6], uint32(msg.PID))
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i

	case ListCommand:
		if msg.Type == RequestType {
			header = make([]byte, 2)
			header[0] = uint8(msg.Type)
			header[1] = uint8(msg.Command)
			i, err_ := writer.Write(header)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			cnt += i
			return cnt, nil
		}

		buf, err_ := json.Marshal(msg.List)
		if err_ != nil {
			return cnt, NewError(FormatError, err_.Error())
		}
		l := len(buf)
		header = make([]byte, l+6)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		binary.BigEndian.PutUint32(header[2:6], uint32(l))
		if l != 0 {
			copy(header[6:], buf)
		}
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i

	case InfoCommand:
		if msg.Type == RequestType {
			header = make([]byte, 2)
			header[0] = uint8(msg.Type)
			header[1] = uint8(msg.Command)
			i, err_ := writer.Write(header)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			cnt += i
			return cnt, nil
		}

		l := len([]byte(msg.Description))
		header = make([]byte, l+6)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		binary.BigEndian.PutUint32(header[2:6], uint32(l))
		if l != 0 {
			copy(header[6:], []byte(msg.Description))
		}
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i

	case KillCommand:
		if msg.Type == RequestType {
			s := msg.AccessKey
			l := len([]byte(s))
			if l > 255 {
				s = s[:255]
				l = 255
			}
			header = make([]byte, l+3)
			header[0] = uint8(msg.Type)
			header[1] = uint8(msg.Command)
			header[2] = uint8(l)
			if l != 0 {
				copy(header[3:], []byte(s))
			}
			i, err_ := writer.Write(header)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			cnt += i
			return cnt, nil
		}

		header = make([]byte, 6)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		binary.BigEndian.PutUint32(header[2:6], uint32(msg.PID))
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i

	case UserCommand:
		m := msg.Method
		m_l := len(m)
		if m_l > 255 {
			m = m[:255]
			m_l = 255
		}
		d_l := len(msg.Data)
		l := m_l + d_l + 1
		header = make([]byte, l+7)
		header[0] = uint8(msg.Type)
		header[1] = uint8(msg.Command)
		binary.BigEndian.PutUint32(header[2:6], uint32(l))
		header[6] = uint8(m_l)
		ofs := 7
		if m_l != 0 {
			copy(header[ofs:], []byte(m))
			ofs += m_l
		}
		if d_l != 0 {
			copy(header[ofs:], msg.Data)
		}
		i, err_ := writer.Write(header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		cnt += i
	}

	return cnt, nil
}

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
func (t *StdTransport) Recv(reader io.Reader, msg *ProtocolMessage) (n int, err *Error) {
	if msg == nil {
		msg = &ProtocolMessage{}
	} else {
		msg.Clear()
	}

	var cnt int = 0
	header := make([]byte, 2)
	i, err_ := io.ReadFull(reader, header)
	if err_ != nil {
		return cnt, NewError(TransportError, err_.Error())
	}
	if i != 2 {
		return cnt, ErrReceiveError
	}
	cnt += i

	msg.Type = MessageType(header[0])
	msg.Command = PluginCommand(header[1])
	if !msg.Type.IsValid() {
		return cnt, ErrIncorrectMessageType
	}
	if !msg.Command.IsValid() {
		return cnt, ErrIncorrectMessageCommand
	}

	if msg.Command == PingCommand || (msg.Type == RequestType && (msg.Command == PidCommand || msg.Command == InfoCommand || msg.Command == ListCommand)) {
		return cnt, nil
	}

	if msg.Type == ResponseErrorType {
		i, err_ = io.ReadFull(reader, header)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		if i != 2 {
			return cnt, ErrReceiveError
		}
		cnt += i

		et := ErrorType(header[0])
		if !et.IsValid() {
			return cnt, ErrIncorrectErrorType
		}
		l := header[1]
		if l != 0 {
			buf := make([]byte, l)
			i, err_ = io.ReadFull(reader, buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if i != int(l) {
				return cnt, ErrReceiveError
			}
			cnt += i
			msg.Error = NewError(et, string(buf))
		}
		return cnt, nil
	}

	switch msg.Command {
	case PidCommand:
		buf := make([]byte, 4)
		i, err_ = io.ReadFull(reader, buf)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		if i != 4 {
			return cnt, ErrReceiveError
		}
		cnt += i
		msg.PID = int(binary.BigEndian.Uint32(buf))

	case ListCommand:
		l_buf := make([]byte, 4)
		i, err_ = io.ReadFull(reader, l_buf)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		if i != 4 {
			return cnt, ErrReceiveError
		}
		cnt += i
		l := binary.BigEndian.Uint32(l_buf)
		if l != 0 {
			buf := make([]byte, l)
			i, err_ = io.ReadFull(reader, buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if uint32(i) != l {
				return cnt, ErrReceiveError
			}
			cnt += i
			err_ = json.Unmarshal(buf, &msg.List)
			if err_ != nil {
				return cnt, NewError(FormatError, err_.Error())
			}
		}

	case InfoCommand:
		l_buf := make([]byte, 4)
		i, err_ = io.ReadFull(reader, l_buf)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		if i != 4 {
			return cnt, ErrReceiveError
		}
		cnt += i
		l := binary.BigEndian.Uint32(l_buf)
		if l != 0 {
			buf := make([]byte, l)
			i, err_ = io.ReadFull(reader, buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if uint32(i) != l {
				return cnt, ErrReceiveError
			}
			cnt += i
			msg.Description = string(buf)
		}

	case KillCommand:
		if msg.Type == RequestType {
			l_buf := make([]byte, 1)
			i, err_ = io.ReadFull(reader, l_buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if i != 1 {
				return cnt, ErrReceiveError
			}
			cnt += i
			l := l_buf[0]
			if l != 0 {
				buf := make([]byte, l)
				i, err_ = io.ReadFull(reader, buf)
				if err_ != nil {
					return cnt, NewError(TransportError, err_.Error())
				}
				if i != int(l) {
					return cnt, ErrReceiveError
				}
				cnt += i
				msg.AccessKey = string(buf)
			}
		} else {
			buf := make([]byte, 4)
			i, err_ = io.ReadFull(reader, buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if i != 4 {
				return cnt, ErrReceiveError
			}
			cnt += i
			msg.PID = int(binary.BigEndian.Uint32(buf))
		}

	case UserCommand:
		l_buf := make([]byte, 4)
		i, err_ = io.ReadFull(reader, l_buf)
		if err_ != nil {
			return cnt, NewError(TransportError, err_.Error())
		}
		if i != 4 {
			return cnt, ErrReceiveError
		}
		cnt += i
		l := binary.BigEndian.Uint32(l_buf)
		if l != 0 {
			buf := make([]byte, l)
			i, err_ = io.ReadFull(reader, buf)
			if err_ != nil {
				return cnt, NewError(TransportError, err_.Error())
			}
			if uint32(i) != l {
				return cnt, ErrReceiveError
			}
			cnt += i
			m_l := int(buf[0])
			if m_l >= 0 && uint32(m_l) <= l {
				if m_l > 0 {
					msg.Method = string(buf[1 : m_l+1])
				} else {
					msg.Method = ""
				}
			}
			st := m_l + 1
			dl := int(l) - st
			if dl > 0 {
				msg.Data = make(json.RawMessage, dl)
				copy(msg.Data, buf[st:])
			}
		}
	}

	return cnt, nil
}
