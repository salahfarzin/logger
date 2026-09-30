package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// The default encoder config must emit caller and function, or
// zap.AddCaller() is silently a no-op.
func TestDefaultEncoderEmitsCallerAndFunction(t *testing.T) {
	var buf bytes.Buffer
	cfg := mergeEncoderConfig(defaultEncoderConfig(), zapcore.EncoderConfig{})
	l := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(&buf), zapcore.DebugLevel), zap.AddCaller())

	l.Warn("boom")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if c, _ := entry["caller"].(string); c == "" {
		t.Fatalf("caller missing: %s", buf.String())
	}
	if f, _ := entry["func"].(string); f == "" {
		t.Fatalf("func missing: %s", buf.String())
	}
}
