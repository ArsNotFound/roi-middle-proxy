package slgoproxy

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/elazarl/goproxy"
)

type LoggerAdapter struct {
	Logger *slog.Logger
}

func (la *LoggerAdapter) Printf(msg string, argv ...any) {
	// remove "[%03d] " in the start and newline at the end
	msg = msg[7 : len(msg)-1]

	// session is always first argument
	session := argv[0].(int64)

	level := slog.LevelError
	if strings.HasPrefix(msg, "INFO: ") {
		level = slog.LevelInfo
		msg = msg[6:] // remove "INFO: "
	} else if strings.HasPrefix(msg, "WARN: ") {
		level = slog.LevelWarn
		msg = msg[6:] // remove "WARN: "
	}

	formattedMsg := fmt.Sprintf(msg, argv[1:]...)
	la.Logger.Log(context.Background(), level, formattedMsg, slog.Int64("session", session))
}

var _ goproxy.Logger = (*LoggerAdapter)(nil)
