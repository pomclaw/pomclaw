# API Configuration Guide

## Overview

The application now supports configurable API and WebSocket ports for easier testing across different environments.

## Configuration Methods

### Method 1: Environment Variables (Build-time)

Set these environment variables during build to hardcode the ports:

```bash
# For Gateway API (default: same as frontend)
export VITE_API_GATEWAY_PORT=3000

# For WebSocket (default: same host, wss/ws based on protocol)
export VITE_API_WEBSOCKET_PORT=18792

npm run build
```

### Method 2: Runtime Configuration (Easiest for Testing)

Set `window.pomclawConfig` before the app loads. This can be done in:

**1. index.html** (before main app script):
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <link rel="icon" href="/favicon.ico" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Pomclaw</title>

  <script>
    // Set custom ports for testing
    window.pomclawConfig = {
      gatewayPort: 8080,      // API Gateway port
      websocketPort: 18792,   // WebSocket port
    }
  </script>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.tsx"></script>
</body>
</html>
```

**2. Browser DevTools Console** (after page load):
```javascript
window.pomclawConfig = {
  gatewayPort: 8080,
  websocketPort: 18792,
}
location.reload()  // Reload for changes to take effect
```

**3. Query Parameters** (add to index.html):
```html
<script>
  const params = new URLSearchParams(window.location.search)
  const gatewayPort = params.get('gatewayPort')
  const websocketPort = params.get('websocketPort')

  if (gatewayPort || websocketPort) {
    window.pomclawConfig = {}
    if (gatewayPort) window.pomclawConfig.gatewayPort = parseInt(gatewayPort)
    if (websocketPort) window.pomclawConfig.websocketPort = parseInt(websocketPort)
  }
</script>
```

Then access: `http://localhost:5173/?gatewayPort=8080&websocketPort=18792`

## Default Behavior

If no configuration is provided:
- **Gateway API**: Uses same origin as frontend (e.g., `http://localhost:5173/api/v1/...`)
- **WebSocket**: Uses same hostname/port as frontend

## Examples

### Local Development (Same Server)
No configuration needed - uses defaults.

### Separate API Server
```html
<script>
  window.pomclawConfig = {
    gatewayPort: 8080,        // API on different port
    websocketPort: 18792,     // WebSocket on different port
  }
</script>
```

### Remote Testing
```html
<script>
  window.pomclawConfig = {
    gatewayPort: 80,          // Production API port
    websocketPort: 443,       // Production WebSocket port
  }
</script>
```

## Architecture

The configuration affects the following API endpoints:

### Gateway API Calls
- Login: `POST http://localhost:8080/api/v1/auth/login`
- Register: `POST http://localhost:8080/api/v1/auth/register`
- Agents: `GET/POST http://localhost:8080/api/v1/agents`
- Sessions: `GET/POST http://localhost:8080/api/v1/sessions`

### WebSocket
- Connection: `ws://localhost:18792/ws?session_id=...&agent_id=...`

## Troubleshooting

### "Failed to fetch" errors
- Check the Gateway API port is correct and server is running
- Verify CORS headers if using different ports
- Check Network tab in DevTools

### "WebSocket connection failed"
- Check WebSocket port is correct
- Verify server is listening on that port
- Check if using secure connection (WSS vs WS)

### Configuration not taking effect
- Ensure `window.pomclawConfig` is set BEFORE app initializes
- Reload the page after changing config
- Check browser console for errors
