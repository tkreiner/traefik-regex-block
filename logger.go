package traefik_regex_block

import (
	"fmt"
	"github.com/zerodha/logf"
	"strings"
	"time"
)

type pluginLogger struct {
	logger *logf.Logger
}

func newPluginLogger(logLevel string, pluginName string) *pluginLogger {
	parsedLogLevel, err := logf.LevelFromString(strings.ToLower(logLevel))
	if err != nil {
		parsedLogLevel = logf.InfoLevel
	}
	log := logf.New(logf.Opts{
		EnableColor:     false,
		Level:           parsedLogLevel,
		EnableCaller:    false,
		TimestampFormat: time.RFC3339Nano,
		DefaultFields:   []any{"plugin", "traefik-regex-block", "pluginName", pluginName},
	})

	log.Info(fmt.Sprintf("Setting log level to %s", strings.ToUpper(parsedLogLevel.String())))

	return &pluginLogger{
		logger: &log,
	}
}

func (l *pluginLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *pluginLogger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *pluginLogger) Error(msg string) {
	l.logger.Error(msg)
}
