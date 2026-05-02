# WebSocket Gateway Implementation Guide

## Overview

This implementation adds a WebSocket-based gateway for real-time bidirectional communication between the frontend and the AI agent. It enables streaming responses for a better user experience.

## Architecture

```
Frontend (WebSocket Client)
    ↓ RequestFrame (chat.send)
Gateway Server (/ws endpoint)
    ↓ MethodRouter dispatch
Chat Handler
    ↓ PublishInbound
Message Bus
    ↓ ConsumeInbound
Agent Loop (Eino)
    ↓ Streaming updates
WebSocket Streamer
    ↓ EventFrame (agent events)
Frontend (updates UI)
```

## Configuration

Add the following to your `etc/pomclaw.yaml` configuration file:

```yaml
gateway:
  host: "0.0.0.0"  # Listen address (default: 0.0.0.0)
  port: 8080       # WebSocket port (default: 8080)
```

If not specified, the gateway will use default values (host: 0.0.0.0, port: 8080).

## Protocol

### Frame Types

1. **RequestFrame** (client → server):
   ```json
   {
     "type": "req",
     "id": "unique-request-id",
     "method": "chat.send",
     "params": { ... }
   }
   ```

2. **ResponseFrame** (server → client):
   ```json
   {
     "type": "res",
     "id": "matching-request-id",
     "ok": true,
     "payload": { ... }
   }
   ```

3. **EventFrame** (server → client, push):
   ```json
   {
     "type": "event",
     "event": "chat",
     "payload": {
       "type": "chunk",
       "content": "..."
     }
   }
   ```

### Available Methods

#### 1. connect
Establishes the WebSocket connection and authenticates the client.

**Request:**
```json
{
  "type": "req",
  "id": "1",
  "method": "connect",
  "params": {
    "user_id": "test_user"
  }
}
```

**Response:**
```json
{
  "type": "res",
  "id": "1",
  "ok": true,
  "payload": {
    "protocol": 3,
    "user_id": "test_user",
    "server": {
      "name": "pomclaw",
      "version": "0.1.0"
    }
  }
}
```

#### 2. chat.send
Sends a message to the agent for processing.

**Request:**
```json
{
  "type": "req",
  "id": "2",
  "method": "chat.send",
  "params": {
    "message": "Hello, how are you?",
    "agentId": "default",
    "sessionKey": "ws:default:test_user"
  }
}
```

**Response (immediate acknowledgment):**
```json
{
  "type": "res",
  "id": "2",
  "ok": true,
  "payload": {
    "acknowledged": true,
    "sessionKey": "ws:default:test_user"
  }
}
```

**Streaming Events (pushed asynchronously):**

Incremental chunks:
```json
{
  "type": "event",
  "event": "chat",
  "payload": {
    "type": "chunk",
    "content": "Hello"
  }
}
```

Final message:
```json
{
  "type": "event",
  "event": "chat",
  "payload": {
    "type": "message",
    "content": "Hello! I'm doing well, thank you for asking!"
  }
}
```

#### 3. chat.history
Retrieves conversation history for a session.

**Request:**
```json
{
  "type": "req",
  "id": "3",
  "method": "chat.history",
  "params": {
    "sessionKey": "ws:default:test_user"
  }
}
```

**Response:**
```json
{
  "type": "res",
  "id": "3",
  "ok": true,
  "payload": {
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      },
      {
        "role": "assistant",
        "content": "Hello! I'm doing well, thank you for asking!"
      }
    ]
  }
}
```

#### 4. chat.abort
Cancels a running conversation.

**Request:**
```json
{
  "type": "req",
  "id": "4",
  "method": "chat.abort",
  "params": {
    "sessionKey": "ws:default:test_user"
  }
}
```

**Response:**
```json
{
  "type": "res",
  "id": "4",
  "ok": true,
  "payload": {
    "ok": true,
    "aborted": true
  }
}
```

#### 5. health
Returns server health status.

**Request:**
```json
{
  "type": "req",
  "id": "5",
  "method": "health"
}
```

**Response:**
```json
{
  "type": "res",
  "id": "5",
  "ok": true,
  "payload": {
    "status": "ok",
    "version": "0.1.0",
    "clients": 1,
    "currentId": "client-uuid"
  }
}
```

## Event Types

### Chat Events

- **chunk**: Incremental content update (streaming)
  ```json
  {
    "type": "chunk",
    "content": "partial text..."
  }
  ```

- **message**: Final complete message
  ```json
  {
    "type": "message",
    "content": "complete response"
  }
  ```

- **thinking**: Agent reasoning (future)
  ```json
  {
    "type": "thinking",
    "content": "reasoning process..."
  }
  ```

### Agent Events (future)

- **run.started**: Agent started processing
- **run.completed**: Agent finished processing
- **run.failed**: Agent encountered an error
- **run.cancelled**: Processing was cancelled
- **tool.call**: Agent is calling a tool
- **tool.result**: Tool execution result

## Testing

### Using the HTML Test Client

1. Start the server:
   ```bash
   go run pomclaw.go -f etc/pomclaw.yaml
   ```

2. Open the test client in your browser:
   ```bash
   # Open test_websocket.html in your browser
   ```

3. Click "Connect" to establish the WebSocket connection

4. Type a message and click "Send Message" to test the chat functionality

5. Watch the messages panel for streaming responses

### Manual Testing with wscat

Install wscat:
```bash
npm install -g wscat
```

Connect and test:
```bash
# Connect
wscat -c ws://localhost:8080/ws

# Send connect frame
{"type":"req","id":"1","method":"connect","params":{"user_id":"test"}}

# Send chat message
{"type":"req","id":"2","method":"chat.send","params":{"message":"Hello!","agentId":"default"}}

# Get history
{"type":"req","id":"3","method":"chat.history","params":{"sessionKey":"ws:default:test"}}
```

## Implementation Files

### New Files Created (in internal/handler/)

1. **websocket_interfaces.go** - Core interfaces (ClientInterface, MethodHandler)
2. **websocket_client.go** - WebSocket client connection (WSClient)
3. **websocket_router.go** - Method routing (WSMethodRouter)
4. **websocket_server.go** - WebSocket server (WSServer)
5. **websocket_chat.go** - Chat method handlers (WSChatHandler)
6. **websocket_streamer.go** - Streaming implementation (WSStreamDelegate, WSStreamer)

### Modified Files

1. **pomclaw.go** - Added WebSocket server initialization
2. **pkg/agent/loop.go** - Integrated streaming updates
3. **internal/config/config.go** - Added gateway configuration
4. **internal/svc/servicecontext.go** - Added SessionManager

## Key Features

✅ WebSocket server with gorilla/websocket
✅ Frame-based protocol (req/res/event)
✅ Method routing system
✅ Real-time streaming responses
✅ Client pool management
✅ Session-based chat history
✅ Message bus integration
✅ Interface-based design for extensibility

## Future Enhancements

- [ ] Authentication and authorization
- [ ] Rate limiting
- [ ] Multi-tenant support
- [ ] More RPC methods (config, sessions management)
- [ ] Reconnection handling
- [ ] Message queuing for offline clients
- [ ] Metrics and monitoring
- [ ] Tool execution events
- [ ] Agent lifecycle events

## Troubleshooting

### WebSocket connection fails

- Check that the server is running: `ps aux | grep pomclaw`
- Verify the port is correct in your client: default is 8080
- Check firewall settings if connecting remotely

### No streaming responses

- Verify the agent loop is running
- Check that the message bus is properly initialized
- Look for errors in server logs

### Messages not persisting

- Verify PostgreSQL/database connection
- Check session store initialization
- Ensure session keys are consistent

## Example Client Implementation

See `test_websocket.html` for a complete working example of a WebSocket client implementation.

## Support

For issues or questions, please refer to the main project documentation or create an issue in the project repository.
