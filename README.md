# ipc_plugin

Go library for developing plugins that communicate with the main application via IPC over UNIX sockets.

The package provides both sides of the interaction:

- **Plugin** — the server side, running in a separate process and handling host commands.
- **Host** — the client side, sending system and user-defined commands to the plugin.
- **PluginManager** — manages the lifecycle of multiple plugins.

> The package is intended for UNIX-like operating systems because its transport uses UNIX domain sockets.

## Features

- IPC over UNIX domain sockets.
- Binary protocol with a fixed two-byte header.
- System commands: `PING`, `PID`, `LIST`, `INFO`, and `KILL`.
- User-defined method registration via `Plugin.Use`.
- JSON parameter and result exchange.
- Connection, send, and receive timeouts.
- Validation of method names and message contents.
- Management of multiple plugins via `PluginManager`.
- Typed library errors categorized by `ErrorType`.

## Requirements

- Go `1.24.1` or a compatible version.
- A UNIX-like operating system with UNIX socket support.

Module:

```text
github.com/conner80/ipc_plugin
```

## Installation

```bash
go get github.com/conner80/ipc_plugin
```

Import:

```go
import ipc_plugin "github.com/conner80/ipc_plugin"
```

## Quick Start

### Plugin

The following plugin registers a `get_user` method, accepts JSON parameters, and returns a JSON result.

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/conner80/ipc_plugin"
    "github.com/conner80/ipc_plugin/examples/shared"
    "github.com/google/uuid"
)

// Socket filename for IPC via UNIX sockets
const socket = "/tmp/ipc_plugin.sock"

// Plugin engine
var p *ipc_plugin.Plugin

// Entry point
func main() {
    log.Print("plugin starting")

    sock := socket
    if len(os.Args) > 1 {
        sock = os.Args[1]
    }

    var err *ipc_plugin.Error
    p, err = ipc_plugin.NewPlugin(sock, "key", "Simple plugin")
    if err != nil {
        log.Fatalln(err.Error())
        return
    }

    if err := p.Use("get_user", handleFunc); err != nil {
        log.Fatalln(err.Error())
        return
    }

    log.Print("plugin started")
    gracefulShutdown()
}

// gracefulShutdown shuts down the plugin application correctly.
func gracefulShutdown() {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit
    p.Done()
}

// handleFunc handles the user's method.
func handleFunc(req *ipc_plugin.ProtocolMessage, resp *ipc_plugin.ProtocolMessage) error {
    var userID uuid.UUID
    if err := req.GetData(&userID); err != nil {
        return fmt.Errorf("%s", err.Error())
    }

    expectedID := uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df")
    if userID != expectedID {
        return fmt.Errorf("user not found")
    }

    user := shared.User{
        ID:    expectedID,
        Name:  "User",
        Mail:  "mail@mail.ru",
        Birth: time.Date(1980, time.September, 27, 15, 0, 0, 0, time.UTC),
    }

    if err := resp.SetData(user); err != nil {
        return fmt.Errorf("%s", err.Error())
    }
    return nil
}
```

`NewPlugin` creates the plugin object and starts its internal listener, which accepts commands from the host. Call `Done` to shut down the process correctly.

### Host

```go
package main

import (
    "log"
    "time"

    "github.com/conner80/ipc_plugin"
    "github.com/conner80/ipc_plugin/examples/shared"
    "github.com/google/uuid"
)

const socket = "/tmp/ipc_plugin.sock"

func main() {
    h, err := ipc_plugin.NewHost(socket, 30*time.Second, 5*time.Second, 5*time.Second, false)
    if err != nil {
        log.Fatalln(err.Error())
        return
    }

    if err = h.Ping(); err == nil {
        log.Println("ping successful")
    } else {
        log.Printf("ping failed: %s\n", err.Error())
    }

    pid, err := h.GetPid()
    if err == nil {
        log.Printf("get PID successful: %d", pid)
    } else {
        log.Printf("get PID failed: %s\n", err.Error())
    }

    info, err := h.GetDescription()
    if err == nil {
        log.Printf("get info successful: %s", info)
    } else {
        log.Printf("get info failed: %s\n", err.Error())
    }

    list, err := h.GetList()
    if err == nil {
        log.Printf("get methods list successful: %v", list)
    } else {
        log.Printf("get methods list failed: %s\n", err.Error())
    }

    var resp shared.User
    err = h.Call("get_user", uuid.Nil, &resp)
    if err == nil {
        log.Printf("call method 1 successful: %v", resp)
    } else {
        log.Printf("call method 1 failed: %s\n", err.Error())
    }

    err = h.Call("get_user", uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df"), &resp)
    if err == nil {
        log.Printf("call method 2 successful: %v", resp)
    } else {
        log.Printf("call method 2 failed: %s\n", err.Error())
    }

    pid, err = h.Kill("key")
    if err == nil {
        log.Printf("kill successful; PID: %d", pid)
    } else {
        log.Printf("kill failed: %s\n", err.Error())
    }
}
```

For graceful shutdown, the host can send `KILL` with the same access key that was provided when creating the plugin:

```go
pid, err := host.Kill("change-me-access-key")
if err != nil {
    log.Fatal(err)
}
fmt.Println("stopped plugin:", pid)
```

## API

### `Plugin`

`Plugin` is the server side of the IPC connection.

```go
plugin, err := ipc_plugin.NewPlugin(socketPath, accessKey, description)
```

Main methods:

| Method | Description |
|---|---|
| `Use(methodName, handler)` | Registers a user-defined method. |
| `Done()` | Stops the listener and releases resources. |

The handler has the following type:

```go
type PluginHanlder func(*ProtocolMessage, *ProtocolMessage) error
```

The typo in the type name is preserved for compatibility with the original API: `PluginHanlder`, not `PluginHandler`.

### `Host`

`Host` is the client used to connect to a running plugin.

```go
host, err := ipc_plugin.NewHost(
    socketPath,
    connectionTimeout,
    sendTimeout,
    receiveTimeout,
    noCheckParam,
)
```

| Method | Description |
|---|---|
| `Ping()` | Checks whether the plugin is available. |
| `GetPid()` | Returns the plugin process PID. |
| `GetList()` | Returns the list of user-defined methods. |
| `GetDescription()` | Returns the plugin description. |
| `Call(methodName, request, response)` | Calls a user-defined method. |
| `Kill(accessKey)` | Requests a graceful shutdown. |

### `PluginInfo` and `PluginManager`

`PluginManager` is used to load and control multiple plugin executables:

```go
package main

import (
    "log"
    "time"

    "github.com/conner80/ipc_plugin"
    "github.com/conner80/ipc_plugin/examples/shared"
    "github.com/google/uuid"
)

func main() {
    log.Println("manager example starting...")

    mgr := ipc_plugin.NewPluginManager("/tmp", []string{"get_user"})
    if err := mgr.LoadPlugin("112233", "key", "bin/plugin", 30*time.Second, 5*time.Second, 5*time.Second); err != nil {
        mgr.Done()
        log.Fatalf("plugin load error: %s", err.Error())
        return
    }

    plg := mgr.GetPlugin("112233")
    if plg != nil {
        var resp shared.User
        err := plg.Call("get_user", uuid.MustParse("ebdb5d69-2f5c-4732-9daa-e1d2359122df"), &resp)
        if err != nil {
            log.Printf("call method error: %s", err.Error())
        } else {
            log.Printf("call method successful: %v", resp)
        }
    }

    mgr.Done()
}
```

Available `PluginInfo` operations include `Run`, `Stop`, `Call`, `CheckRunning`, `IsAlive`, `PID`, `Description`, `List`, and `LastError`.

## ProtocolMessage

`ProtocolMessage` represents a protocol message:

```go
type ProtocolMessage struct {
    Type        MessageType
    Command     PluginCommand
    Method      string
    Error       *Error
    PID         int
    Description string
    List        CommandsList
    AccessKey   string
    Data        json.RawMessage
}
```

Use the JSON methods to work with user data:

```go
if err := message.SetData(request); err != nil {
    // handle the error
}

if err := message.GetData(&request); err != nil {
    // handle the error
}
```

`Clear()` resets the message data, while `Validate()` checks whether the message is valid.

## Exchange Protocol

All messages begin with a two-byte header:

| Offset | Size | Field |
|---:|---:|---|
| `0x0000` | 1 byte | Message type. |
| `0x0001` | 1 byte | Command. |
| `0x0002` | Variable | Message body. |

### Message Types

| Value | Constant | Description |
|---:|---|---|
| `0x00` | `RequestType` | Request from the host to the plugin. |
| `0x01` | `ResponseSuccessType` | Successful response from the plugin. |
| `0x02` | `ResponseErrorType` | Error response. |

### Commands

| Value | Constant | Description |
|---:|---|---|
| `0x00` | `PingCommand` | Checks plugin availability. |
| `0x01` | `PidCommand` | Returns the process PID. |
| `0x02` | `ListCommand` | Returns the list of methods. |
| `0x03` | `InfoCommand` | Returns the plugin description. |
| `0x04` | `KillCommand` | Performs a graceful shutdown. |
| `0x05` | `UserCommand` | Calls a user-defined method. |

Multi-byte integers are encoded in big-endian byte order. Strings use UTF-8. JSON must be valid UTF-8.

### Error Response

The body of a response with `ResponseErrorType` has the following format:

| Offset | Size | Field |
|---:|---:|---|
| `0x0002` | 1 byte | Error type. |
| `0x0003` | 1 byte | Error text length. |
| `0x0004` | 1–255 bytes | Error text. |

### User Command

`UserCommand` requests and responses use the following structure:

| Offset | Size | Field |
|---:|---:|---|
| `0x0002` | 4 bytes | Body length in big-endian order. |
| `0x0006` | 1 byte | Method name length. |
| `0x0007` | 0–255 bytes | Method name. |
| after the name | Variable | JSON parameters or result. |

The body length is `1 + method_name_length + JSON_length`.

## Constraints and Validation

- The method name must match the regular expression `^[A-Za-z_][A-Za-z0-9_-]*$`.
- The method name length must be between 1 and 255 bytes.
- The access key length must be between 1 and 255 bytes.
- Error text is limited to 255 bytes.
- The plugin description is limited to 5,000 UTF-8 characters according to the API documentation.
- All user-method data must be serializable as JSON.
- Treat the access key as a secret; do not pass it through command-line arguments or write it to logs.
- The UNIX socket should have permissions that prevent untrusted processes from connecting to it.

## Errors

The package uses its own `Error` type:

```go
type Error struct {
    Type      ErrorType
    ErrorText string
}
```

Error categories:

| Type | Description |
|---|---|
| `PluginPanicError` | Critical plugin error or panic. |
| `HandlerError` | Error returned by a user-defined handler. |
| `FormatError` | Message or parameter format error. |
| `TransportError` | IPC transport error. |

Checking an error:

```go
if err != nil {
    fmt.Printf("type=%v message=%s\n", err.Type, err.Error())
}
```

Use `errors.Is` to compare an error with predefined errors because `Error` implements the `Is` method:

```go
if errors.Is(err, ipc_plugin.ErrPluginMethodNotFound) {
    // method not found
}
```

## Transport

The transport is exposed through the `ITransport` interface:

```go
type ITransport interface {
    Send(writer io.Writer, msg *ProtocolMessage) (int, *Error)
    Recv(reader io.Reader, msg *ProtocolMessage) (int, *Error)
}
```

Create the standard implementation with:

```go
transport := ipc_plugin.NewTransport()
```

## Utility Functions

```go
socketDir := ipc_plugin.GetDefaultSocketDir()
socketPath, err := ipc_plugin.GetDefaultSocketPath()
exists, err := ipc_plugin.IsProcessExists(pid)
err := ipc_plugin.KillProcess(pid)
err := ipc_plugin.IsValidMethodName("get_user")
```

The current package version is available through the following constant:

```go
fmt.Println(ipc_plugin.Version)
```

## Build and Verification

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./...
```

Before integration, verify the following:

- The permissions of the UNIX socket directory and socket file are correct.
- The `accessKey` is the same on the plugin and host sides.
- Timeouts match the expected execution time of handlers.
- All required methods passed to `NewPluginManager` are available.
- Shutdown is performed correctly through `Done`, `Stop`, or `UnloadPlugin`.

## Documentation

- [**PROTOCOL.md**](docs/PROTOCOL.md) — detailed binary protocol specification.
- [**DOCUMENTATION.md**](docs/DOCUMENTATION.md) — generated package API documentation.

## License

[GNU license](https://www.gnu.org/licenses/gpl-3.0.html)
