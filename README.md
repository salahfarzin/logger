

# Logger

This package provides a structured logger for Go, built on top of [zap](https://github.com/uber-go/zap). It offers a simple API for structured, JSON logging in your applications.

- Structured, JSON-formatted logs
- Log levels: debug, info, warn, error
- Configurable log output (stdout or file)
- Context fields (e.g., trace IDs)
- Compatible with standard log interface (Print, Printf)

# Installation

Add to your `go.mod`:

```
go get github.com/salahfarzin/logger
```

If you want to load environment variables from a `.env` file:

```
go get github.com/joho/godotenv
```

## Documentation
- [Usage](docs/usage.md)
- [API](docs/api.md)
- [Context Propagation for Distributed Systems](docs/context.md)
- [Notes](docs/notes.md)