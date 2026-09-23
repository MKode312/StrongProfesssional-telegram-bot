package slogpretty

import (
	"context"
	"encoding/json"
	"io"
	stdlog "log"
	"log/slog"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	log   *stdlog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(out io.Writer) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		log:     stdlog.New(out, "", 0),
	}
}

func (h *PrettyHandler) Handle(_ context.Context, record slog.Record) error {
	level := record.Level.String() + ":"
	switch record.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]any, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.Any()
		return true
	})
	for _, attr := range h.attrs {
		fields[attr.Key] = attr.Value.Any()
	}

	var fieldsJSON []byte
	if len(fields) > 0 {
		var err error
		fieldsJSON, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	h.log.Println(record.Time.Format("[15:05:05.000]"), level, color.CyanString(record.Message), color.WhiteString(string(fieldsJSON)))
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{Handler: h.Handler, log: h.log, attrs: append(h.attrs, attrs...)}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{Handler: h.Handler.WithGroup(name), log: h.log}
}
