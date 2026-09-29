# IPC Plugin Exchange Protocol

This document describes the binary exchange protocol for communication between the plugin host (client) and plugin core (server) via UNIX sockets.

***

## Message Header

All messages (requests and responses) begin with a fixed 2-byte header:

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | **Message Type**: See message type below |
| 0x0001 | 1      | **Command Type**: See command codes below |

### Message Types

| Value | Name              | Description                          |
|-------|-------------------|--------------------------------------|
| 0x00  | `RequestType`     | Request from host to plugin          |
| 0x01  | `ResponseSuccessType` | Successful response from plugin   |
| 0x02  | `ResponseErrorType`   | Error response from plugin        |

### Command Types

| Value | Name            | Description                           |
|-------|-----------------|---------------------------------------|
| 0x00  | `PingCommand`   | Check if plugin is alive              |
| 0x01  | `PidCommand`    | Get plugin process PID                |
| 0x02  | `ListCommand`   | Get supported methods list            |
| 0x03  | `InfoCommand`   | Get plugin description                |
| 0x04  | `KillCommand`   | Graceful shutdown plugin              |
| 0x05  | `UserCommand`   | Call user-defined method              |

***

## Error Response Format

For error responses (Message Type = `0x02`), the body contains error details:

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0002 | 1      | **Error Type** (see table below)     |
| 0x0003 | 1      | **Error Text Length** (1–255 bytes)  |
| 0x0004 | 1–255  | **Error Text Message** (ASCII string)       |

### Error Types

| Value | Name                  | Description                          |
|-------|-----------------------|--------------------------------------|
| 0x00  | `PluginPanicError`    | Plugin panic/recover                 |
| 0x01  | `HandlerError`        | User handler execution error         |
| 0x02  | `FormatError`         | Message format/parsing error         |
| 0x03  | `TransportError`      | Network/IPC transport error          |

***

## Command Specifications

Commands responses have success response type (Message Type = `0x01`) by default.

### 1. PING Command

Check if the plugin is alive and responding.

#### Request

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | Message Type = `0x00` (Request) |
| 0x0001 | 1      | Command = `0x00` (PingCommand) |

**No body.**

#### Response

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | Message Type = `0x01` (Response Success) |
| 0x0001 | 1      | Command = `0x00` (PingCommand) |

**No body.**

***

### 2. PID Command

Request the plugin's process ID.

#### Request

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | Message Type = `0x00` (Request) |
| 0x0001 | 1      | Command = `0x01` (PidCommand) |

**No body.**

#### Response

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0000 | 1      | Message Type = `0x01` (Response Success)     |
| 0x0001 | 1      | Command = `0x01` (PidCommand)        |
| 0x0002 | 4      | **PID** (uint32, big-endian)         |

***

### 3. LIST Command

Request the list of supported user methods.

#### Request

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | Message Type = `0x00` (Request) |
| 0x0001 | 1      | Command = `0x02` (ListCommand) |

**No body.**

#### Response

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0000 | 1      | Message Type = `0x01` (Response Success)     |
| 0x0001 | 1      | Command = `0x02` (ListCommand)       |
| 0x0002 | 4      | **List Length** (uint32, big-endian) — length of JSON array in bytes |
| 0x0006 | Dyn    | **Method List** — JSON array of strings (e.g., `["get_user","set_data"]`) |

***

### 4. INFO Command

Request the plugin's description text.

#### Request

| Offset | Length | Description |
|--------|--------|-------------|
| 0x0000 | 1      | Message Type = `0x00` (Request) |
| 0x0001 | 1      | Command = `0x03` (InfoCommand) |

**No body.**

#### Response

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0000 | 1      | Message Type = `0x01` (Response Success)     |
| 0x0001 | 1      | Command = `0x03` (InfoCommand)       |
| 0x0002 | 4      | **Description Length** (uint32, big-endian) — length in bytes (max 5000 characters in UTF8) |
| 0x0006 | 0–5000 | **Description Text** (UTF-8)         |

***

### 5. KILL Command

Graceful shutdown of the plugin process.

#### Request

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0000 | 1      | Message Type = `0x00` (Request)      |
| 0x0001 | 1      | Command = `0x04` (KillCommand)       |
| 0x0002 | 1      | **Access Key Length** (1–255)        |
| 0x0003 | 1–255  | **Access Key** (ASCII string)        |

#### Response

| Offset | Length | Description                          |
|--------|--------|--------------------------------------|
| 0x0000 | 1      | Message Type = `0x01` (Response Success)     |
| 0x0001 | 1      | Command = `0x04` (KillCommand)       |
| 0x0002 | 4      | **PID** (uint32, big-endian) — plugin's PID before shutdown |

***

### 6. USER Command

Call a user-defined method with optional parameters.

#### Request

| Offset       | Length | Description                          |
|--------------|--------|--------------------------------------|
| 0x0000       | 1      | Message Type = `0x00` (Request)      |
| 0x0001       | 1      | Command = `0x05` (UserCommand)       |
| 0x0002       | 4      | **Body Length** (uint32, big-endian) — total length of: `1 + method_name_len + data_len` |
| 0x0006       | 1      | **Method Name Length** (0–255)       |
| 0x0007       | 0–255  | **Method Name** (ASCII identifier: `^[A-Za-z_][A-Za-z0-9_-]*$`) |
| 0x0008–0x0106| Dyn    | **User Data** — JSON-encoded parameters (optional) |

#### Response

| Offset       | Length | Description                          |
|--------------|--------|--------------------------------------|
| 0x0000       | 1      | Message Type = `0x01` (Response Success)     |
| 0x0001       | 1      | Command = `0x05` (UserCommand)       |
| 0x0002       | 4      | **Body Length** (uint32, big-endian) — total length of: `1 + method_name_len + data_len` |
| 0x0006       | 1      | **Method Name Length** (0–255) |
| 0x0007       | 0–255  | **Method Name** (ASCII identifier: `^[A-Za-z_][A-Za-z0-9_-]*$`) |
| 0x0008–0x0106| Dyn    | **Result Data** — JSON-encoded result or error details |

***

## Example Flows

### Ping Request/Response

**Request (2 bytes):**
```
[0x00, 0x00]  // Type=Request, Command=Ping
```

**Response (2 bytes):**
```
[0x01, 0x00]  // Type=Response, Command=Ping
```

### User Method Call (get_user)

**Request (54 bytes):**
```
[0x00, 0x05, 0x00, 0x00, 0x00, 0x2F, 0x08, 
 'g','e','t','_','u','s','e','r',
 '"','e','b','d','b','5','d','6','9','-','2','f','5','c','-','4','7','3','2','-','9','d','a','a','-','e','1','d','2','3','5','9','1','2','2','d','f','"']
```

Breakdown:
- `0x00` = RequestType
- `0x05` = UserCommand
- `0x0000002F` = Body Length (47 bytes: 1 + 8 + 38)
- `0x08` = Method Name Length (8)
- `get_user` = Method Name
- `"ebdb5d69-2f5c-4732-9daa-e1d2359122df"` = JSON data (38 bytes)

**Response (125 bytes):**
```
[0x01, 0x05, 0x00, 0x00, 0x00, 0x76, 0x00,
 '{','"','i','d','"':','"','e','b','d','b',...,'}']
```

Breakdown:
- `0x01` = ResponseSuccessType
- `0x05` = UserCommand
- `0x00000076` = Body Length (118 bytes: 1 + 0 + 117)
- `0x00` = Method Name Length (0 — not included in response)
- `{"id":"ebdb5d69-...","name":"User",...}` = JSON result (117 bytes)

***

## Notes

- All multi-byte integers are encoded in **big-endian** (network byte order).
- Strings are encoded as **UTF-8**.
- JSON data must be valid UTF-8 and properly escaped.
- Maximum lengths:
  - Error text: 255 bytes (ASCII)
  - Access key: 255 bytes (ASCII)
  - Method name: 255 bytes (ASCII)
  - Description: 5000 characters (UTF8)
- Method names must match regex: `^[A-Za-z_][A-Za-z0-9_-]*$`