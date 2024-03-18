package logger

import (
	"log/slog"
)

type CustomLogger struct {
    logger *slog.Logger
    plugin string
    pluginName   string
}

func NewCustomLogger(plugin string, pluginName string, handler slog.Handler) *CustomLogger {
    logger := slog.New(handler) // Create an underlying slog.Logger
    return &CustomLogger{
        logger: logger,
        plugin: plugin,
        pluginName: pluginName,
    }
}

func (cl *CustomLogger) Info(msg string, attrs ...slog.Attr) {
    // Convert attrs slice to a new slice of type []interface{}
    convertedAttrs := make([]interface{}, len(attrs))
    for i, attr := range attrs {
        convertedAttrs[i] = attr
    }
    cl.logger.Info(msg, append(append(convertedAttrs, slog.String("plugin", cl.plugin)), slog.String("pluginName", cl.pluginName))...)
}

func (cl *CustomLogger) Error(msg string, attrs ...slog.Attr) {
    // Convert attrs slice to a new slice of type []interface{}
    convertedAttrs := make([]interface{}, len(attrs))
    for i, attr := range attrs {
        convertedAttrs[i] = attr
    }
    cl.logger.Error(msg, append(append(convertedAttrs, slog.String("plugin", cl.plugin)), slog.String("pluginName", cl.pluginName))...)
}

func (cl *CustomLogger) Debug(msg string, attrs ...slog.Attr) {
    // Convert attrs slice to a new slice of type []interface{}
    convertedAttrs := make([]interface{}, len(attrs))
    for i, attr := range attrs {
        convertedAttrs[i] = attr
    }
    cl.logger.Debug(msg, append(append(convertedAttrs, slog.String("plugin", cl.plugin)), slog.String("pluginName", cl.pluginName))...)
}

