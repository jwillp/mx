package mx

import (
	"context"
	"fmt"
	"github.com/logrusorgru/aurora/v4"
	"github.com/samber/lo"
	"hash/fnv"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// humanReadableLogHandler creates a development-focused log handler inspired by Spring Boot's logging format.
// It outputs logs with proper coloring and subsystem information.
//
// Format:
//
//	YYYY-MM-DD HH:MM:SS.SSS LEVEL [SCOPE] message key1=value1 key2=value2
//
// Example output (with colors):
//
//	2025-11-25 10:15:32.456 INFO  -- [PaymentService] : Payment processed successfully
//	2025-11-25 10:15:32.567 WARN  -- [AuthService]    : Token expiration coming up
//	2025-11-25 10:15:32.678 ERROR -- [DataService]    : Connection timeout
//	2025-11-25 10:15:32.789 INFO  -- [system]         : System initialization complete
//
// Each subsystem gets a unique color (256-color palette) based on its name.
// The "system" subsystem has no color (special case).
// Levels are left-aligned and padded to 5 characters.
// Subsystem names are left-aligned and padded.
type humanReadableLogHandler struct {
	colorizer *aurora.Aurora
	mu        *sync.Mutex
	w         io.Writer

	groups []string

	opts  slog.HandlerOptions
	attrs []slog.Attr
}

// NewHumanReadableLogHandler creates a new development-focused log handler with colored output.
func NewHumanReadableLogHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	return humanReadableLogHandler{
		w:         w,
		opts:      *opts,
		mu:        &sync.Mutex{},
		colorizer: aurora.New(aurora.WithColors(true)),
	}
}

func (h humanReadableLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h humanReadableLogHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	record.AddAttrs(h.attrs...)

	var attrs []slog.Attr
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	scopeName := h.findScope(ctx)
	logRecord := humanReadableLogRecord{
		colorizer: h.colorizer,
		timestamp: record.Time,
		level:     record.Level,
		scope:     scopeName,
		message:   record.Message,
		attrs:     attrs,
	}

	_, err := fmt.Fprintln(h.w, logRecord.String())
	return err
}

type humanReadableLogRecord struct {
	colorizer *aurora.Aurora

	timestamp time.Time
	level     slog.Level
	scope     string
	message   string
	attrs     []slog.Attr

	lineColor aurora.Color
}

func (r humanReadableLogRecord) String() string {
	var colorOverride *aurora.Color
	switch r.level {
	case slog.LevelDebug:
		colorOverride = lo.ToPtr(aurora.Color(0).Bold().Green())
	case slog.LevelWarn:
		colorOverride = lo.ToPtr(aurora.Color(0).Bold().Yellow())
	case slog.LevelError:
		colorOverride = lo.ToPtr(aurora.Color(0).Bold().Red())
	}

	timestamp := r.formatTimestamp(colorOverride)
	level := r.formatLevel(colorOverride)
	scope := r.formatScope(colorOverride)
	message := r.formatMessage(colorOverride)
	attributes := r.formatAttributes(colorOverride)

	line := fmt.Sprintf("%s %s -- %s : %s %s", timestamp, level, scope, message, attributes)
	if colorOverride != nil {
		line = r.colorizer.Colorize(line, *colorOverride).String()
	}

	return line
}

func (r humanReadableLogRecord) formatTimestamp(colorOverride *aurora.Color) string {
	timestamp := r.timestamp.Format("2006-01-02 15:04:05.000")

	if colorOverride == nil {
		timestamp = r.colorizer.Gray(12, timestamp).String()
	}

	return timestamp
}

func (r humanReadableLogRecord) formatAttributes(colorOverride *aurora.Color) string {
	attrs := lo.Map(r.attrs, func(attr slog.Attr, _ int) string {
		if attr.Key == logKeySubsystem {
			return ""
		}

		key := attr.Key
		value := attr.Value.String()

		if colorOverride == nil {
			key = r.colorizer.Magenta(attr.Key).String()
			value = r.colorizer.Gray(12, fmt.Sprint(attr.Value)).String()
		}
		return fmt.Sprintf("%s=%s", key, value)
	})
	attrs = lo.Filter(attrs, func(attr string, _ int) bool { return attr != "" })

	return strings.Join(attrs, " ")
}

func (r humanReadableLogRecord) formatLevel(colorOverride *aurora.Color) string {
	level := strings.ToUpper(r.level.String())
	level = fmt.Sprintf("%-5s", level)

	if colorOverride == nil {
		switch r.level {
		case slog.LevelDebug:
			level = r.colorizer.Green(level).String()
		case slog.LevelInfo:
			level = r.colorizer.Blue(level).String()
		case slog.LevelWarn:
			level = r.colorizer.Yellow(level).String()
		case slog.LevelError:
			level = r.colorizer.Red(level).String()
		default:
			level = r.colorizer.Gray(12, level).String()
		}
	}

	return level
}

func (r humanReadableLogRecord) formatScope(colorOverride *aurora.Color) string {
	scope := r.scope
	if colorOverride == nil && scope != "system" {
		hash := fnv.New32a()
		_, _ = hash.Write([]byte(scope))
		colorIndex := uint8((hash.Sum32() % 216) + 16)
		scope = r.colorizer.Index(aurora.ColorIndex(colorIndex), scope).String()
	}

	columnSize := 20
	padding := columnSize - len(r.scope) - 2

	return fmt.Sprintf("[%s]%s", scope, strings.Repeat(" ", padding))
}

func (r humanReadableLogRecord) formatMessage(colorOverride *aurora.Color) string {
	message := r.message

	if colorOverride == nil {
		switch r.level {
		case slog.LevelDebug:
			message = r.colorizer.Colorize(message, aurora.Color(0).Green()).String()
		case slog.LevelError:
			message = r.colorizer.Colorize(message, aurora.Color(0).Red()).String()
		case slog.LevelWarn:
			message = r.colorizer.Colorize(message, aurora.Color(0).Yellow()).String()
		}
	}

	if r.scope != "system" {
		message = "\t" + message
	}

	return message
}

//func (r humanReadableLogRecord) colorize() {
//
//	// Time
//
//	// Level
//	switch r.level {
//	case slog.LevelDebug:
//		r.level = r.colorizer.Green(levelStr).String()
//	case slog.LevelInfo:
//		r.level = r.colorizer.Blue(levelStr).String()
//	case slog.LevelWarn:
//		r.level = r.colorizer.Yellow(levelStr).String()
//	case slog.LevelError:
//		r.level = r.colorizer.Red(levelStr).String()
//	default:
//		r.level = r.colorizer.Gray(12, levelStr).String()
//	}
//
//	// Scope
//	if r.scope != "system" {
//		hash := fnv.New32a()
//		_, _ = hash.Write([]byte(r.scope))
//		colorIndex := uint8((hash.Sum32() % 216) + 16)
//		r.scope = r.colorizer.Index(aurora.ColorIndex(colorIndex), r.scope).String()
//	}
//}

func (h humanReadableLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h.attrs = append(h.attrs, attrs...)

	return h
}

func (h humanReadableLogHandler) WithGroup(name string) slog.Handler {
	h.groups = append(h.groups, name)
	return h
}

func (h humanReadableLogHandler) findScope(ctx context.Context) string {
	scope := "system"
	if subsystem, ok := ctx.Value(applicationSubsystemNameKey{}).(string); ok && subsystem != "" {
		scope = subsystem
	}
	return scope
}
