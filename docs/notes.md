# Notes

- This logger is based on zap's production configuration (JSON output).
- Do not log sensitive data.
- Always call `logger.Sync()` before your program exits to flush logs.
