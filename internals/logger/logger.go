package logger

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func Logf(ctx context.Context, flow, identifier, format string, args ...interface{}) {
	reqID := "no-req-id"
	if ctx != nil {
		if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
			reqID = id
		}
	}
	if identifier == "" {
		identifier = "-"
	}

	pc, file, _, ok := runtime.Caller(1)
	fileFunc := "unknown:unknown"
	if ok {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			file = parts[len(parts)-1]
		}
		f := runtime.FuncForPC(pc)
		funcName := "unknown"
		if f != nil {
			funcParts := strings.Split(f.Name(), "/")
			funcName = funcParts[len(funcParts)-1]
			if idx := strings.LastIndex(funcName, "."); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}
		fileFunc = fmt.Sprintf("%s:%s", file, funcName)
	}

	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] [%s] [%s] - %s", flow, identifier, reqID, fileFunc, msg)
}
