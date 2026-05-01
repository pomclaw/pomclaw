# Logging Configuration

The pomclaw application now supports centralized logging configuration through the configuration file (JSON or YAML).

## Configuration Structure

```yaml
logging:
  level: INFO  # DEBUG, INFO, WARN, ERROR, FATAL
  file_path: ""  # Optional: specify a file path for file-based logging
```

## Log Levels

| Level | Description | Usage |
|-------|-------------|-------|
| DEBUG | Detailed debug information | Development and troubleshooting |
| INFO | General informational messages (default) | Standard operation |
| WARN | Warning messages for potentially problematic situations | Warning about issues |
| ERROR | Error messages for failures | Application errors |
| FATAL | Fatal errors that terminate the application | Critical failures |

## Examples

### Console Logging Only (Default)

```yaml
logging:
  level: INFO
  file_path: ""
```

Output will be printed to standard output with timestamps and levels.

### Debug Mode

```yaml
logging:
  level: DEBUG
  file_path: ""
```

All debug messages will be logged to console.

### File Logging

```yaml
logging:
  level: INFO
  file_path: /var/log/pomclaw/app.log
```

Logs will be written to both console and the specified file in JSON format.

### Production Setup

```yaml
logging:
  level: WARN
  file_path: /var/log/pomclaw/pomclaw.log
```

Only warnings and errors are logged, reducing noise in production.

## Log Output Format

### Console Output
```
[2026-04-28T10:15:30Z] [INFO] [component] message {key=value, key2=value2}
```

### File Output (JSON)
```json
{
  "level": "INFO",
  "timestamp": "2026-04-28T10:15:30Z",
  "component": "gateway",
  "message": "Connection established",
  "fields": {"host": "localhost", "port": 18792},
  "caller": "/path/to/file.go:123 (function.name)"
}
```

## Command Line Override

The `--debug` flag still works and overrides the config file setting:

```bash
pomclaw gateway -f config.yaml --debug
```

This will set log level to DEBUG regardless of config file setting.

## Configuration in Different Environments

### Development (config-dev.yaml)

```yaml
logging:
  level: DEBUG
  file_path: ""  # Console only for faster debugging
```

### Testing (config-test.yaml)

```yaml
logging:
  level: INFO
  file_path: /tmp/pomclaw-test.log
```

### Production

```yaml
logging:
  level: WARN
  file_path: /var/log/pomclaw/production.log
```

## Implementing Logging in Code

The logger package provides several helper methods:

```go
import "github.com/pomclaw/pomclaw/pkg/logger"

// Simple logging
logger.Info("Application started")
logger.Debug("Debug information")
logger.Warn("Warning message")
logger.Error("An error occurred")

// Logging with component
logger.InfoC("gateway", "Server started")

// Logging with fields
logger.InfoF("Application started", map[string]interface{}{
    "port": 8080,
    "env": "production",
})

// Logging with component and fields
logger.InfoCF("gateway", "Connection established", map[string]interface{}{
    "client": "127.0.0.1",
    "protocol": "WebSocket",
})
```

## Log File Management

When file logging is enabled:

1. Log files are created with append mode
2. JSON-formatted logs are written line-by-line
3. Files are not automatically rotated (use external tools like `logrotate`)

Example logrotate configuration:

```
/var/log/pomclaw/*.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0640 pomclaw pomclaw
    sharedscripts
    postrotate
        systemctl reload pomclaw > /dev/null 2>&1 || true
    endscript
}
```

## Programmatic Control

You can also control logging at runtime (useful for testing):

```go
import "github.com/pomclaw/pomclaw/pkg/logger"

// Change log level at runtime
logger.SetLevel(logger.DEBUG)

// Enable file logging
logger.EnableFileLogging("/tmp/app.log")

// Disable file logging
logger.DisableFileLogging()

// Get current log level
currentLevel := logger.GetLevel()
```
