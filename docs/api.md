# API

- `logger.Init()` — Initialize the logger (call once, early in main)
- `logger.Sync()` — Flush logs (call on shutdown)
- `logger.Info(msg, fields...)`
- `logger.Error(msg, err, fields...)`
- `logger.Debug(msg, fields...)`
- `logger.With(fields...)` — Get a logger with extra context
- `logger.Printf(format, v...)`, `logger.Print(v...)` — std log compatibility

## Example .env

```
LOG_PATH=logs/app.log
LOG_LEVEL=info
```
