# Usage

## 1. Configure Environment

Set log output and level using environment variables, e.g. in a `.env` file:

```
LOG_PATH=app.log         # or leave empty for stdout
LOG_LEVEL=debug          # debug, info, warn, error
```

Load your `.env` file at the start of your application (optional, but recommended):

```go
import "github.com/joho/godotenv"

func main() {
    _ = godotenv.Load() // loads .env file if present
    logger.Init()
    defer logger.Sync() // flush logs on shutdown
    logger.Info("Logger initialized")
}
```

## 2. Logging

```go
logger.Info("something happened", zap.String("user", "alice"))
logger.Error("something failed", err, zap.String("user", "bob"))
logger.Debug("debug details", zap.Int("count", 42))

// Add context fields
reqLogger := logger.With(zap.String("trace_id", "abc-123"))
reqLogger.Info("request started")
```

## 3. Log Output

- If `LOG_PATH` is not set, logs go to stdout.
- If `LOG_PATH` is set to a file path, logs are written to that file.
- Make sure the directory exists and is writable.

## 4. Log Levels

Set `LOG_LEVEL` to one of: `debug`, `info`, `warn`, `error`.
