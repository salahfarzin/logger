# Context Propagation for Distributed Systems

This package includes helpers in `context.go` for propagating loggers with context, which is useful in distributed systems (e.g., microservices, gRPC, HTTP servers):

- `logger.WithLogger(ctx, l)` — Returns a new context with the provided logger attached.
- `logger.FromContext(ctx)` — Retrieves the logger from context, or falls back to the global logger if not found.

**Example:**

```go
import (
    "context"
    "github.com/salahfarzin/logger"
)

func handler(ctx context.Context) {
    log := logger.FromContext(ctx)
    log.Info("handling request")
}

func main() {
    logger.Init()
    ctx := logger.WithLogger(context.Background(), logger.Get())
    handler(ctx)
}
```

This allows you to pass request-scoped loggers (with trace IDs, user info, etc.) through your application using context, which is a best practice for distributed systems.
