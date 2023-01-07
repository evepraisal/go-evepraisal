package accesslog

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type contextKey int

var (
	ctxLoggerKey contextKey
)

type LogRecord struct {
	Time                                      time.Time
	Ip, Method, Uri, Protocol, Username, Host string
	Status                                    int
	Size                                      int64
	ElapsedTime                               time.Duration
	RequestHeader                             http.Header
	CustomRecords                             map[string]string
}

type LoggingWriter struct {
	http.ResponseWriter
	logRecord LogRecord
}

func (r *LoggingWriter) Write(p []byte) (int, error) {
	if r.logRecord.Status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		r.logRecord.Status = http.StatusOK
	}
	written, err := r.ResponseWriter.Write(p)
	r.logRecord.Size += int64(written)
	return written, err
}

func (r *LoggingWriter) WriteHeader(status int) {
	r.logRecord.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// SetCustomLogRecord and GetCustomLogRecord functions provide accessors to the logRecord.CustomRecords.
// You can use it to store arbitrary strings that are relevant to this request.
//
// Alternative method would be to store the value in context.
// Which doesn't work when you want to retrieve the value from a HTTP middleware that is earlier in the middleware chain, eg: accesslog, recovery.
//
// w.(accesslogger.LoggingWriter).SetCustomLogRecord("X-User-Id", "3")
func (r *LoggingWriter) SetCustomLogRecord(key, value string) {
	if r.logRecord.CustomRecords == nil {
		r.logRecord.CustomRecords = map[string]string{}
	}
	r.logRecord.CustomRecords[key] = value
}

// w.(accesslogger.LoggingWriter).GetCustomLogRecord("X-User-Id")
func (r *LoggingWriter) GetCustomLogRecord(key string) string{
	return r.logRecord.CustomRecords[key]
}

// http.CloseNotifier interface
func (r *LoggingWriter) CloseNotify() <-chan bool {
	if w, ok := r.ResponseWriter.(http.CloseNotifier); ok {
		return w.CloseNotify()
	}
	return make(chan bool)
}

func (r *LoggingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
}

// http.Flusher
func (r *LoggingWriter) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

// http.Pusher
func (r *LoggingWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.ResponseWriter.(http.Pusher)
	if ok {
		return pusher.Push(target, opts)
	}
	return fmt.Errorf("ResponseWriter doesn't support Pusher interface")
}

// WrapWriter interface
func (r *LoggingWriter) WrappedWriter() http.ResponseWriter {
	return r.ResponseWriter
}

type Logger interface {
	Log(record LogRecord)
}

type ContextLogger interface {
	Logger
	LogContext(context.Context, LogRecord)
}

type LoggingHandler struct {
	handler   http.Handler
	logger    ContextLogger
	logBefore bool
}

type wrapLogger struct {
	Logger
}

func (wl *wrapLogger) LogContext(ctx context.Context, l LogRecord) {
	wl.Log(l)
}

func contextLogger(l Logger) ContextLogger {
	if cl, ok := l.(ContextLogger); ok {
		return cl
	}
	return &wrapLogger{l}
}

func NewLoggingHandler(handler http.Handler, logger Logger) http.Handler {
	return &LoggingHandler{
		handler:   handler,
		logger:    contextLogger(logger),
		logBefore: false,
	}
}

func NewAroundLoggingHandler(handler http.Handler, logger Logger) http.Handler {
	return &LoggingHandler{
		handler:   handler,
		logger:    contextLogger(logger),
		logBefore: true,
	}
}

func NewLoggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := NewLoggingHandler(next, logger)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}

func NewAroundLoggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := NewAroundLoggingHandler(next, logger)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}

// readIp return the real ip when behide nginx or apache
func (h *LoggingHandler) realIp(r *http.Request) string {
	// Check if behide nginx or apache
	xRealIP := r.Header.Get("X-Real-Ip")
	if xRealIP != "" {
		return xRealIP
	}

	xForwardedFor := r.Header.Get("X-Forwarded-For")
	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		if address != "" {
			return address
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func GetLoggingWriter(ctx context.Context) *LoggingWriter {
	iface := ctx.Value(ctxLoggerKey)
	if l, ok := iface.(*LoggingWriter); ok {
		return l
	}
	return nil
}

func (h *LoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ip := h.realIp(r)
	username := "-"
	if r.URL.User != nil {
		if name := r.URL.User.Username(); name != "" {
			username = name
		}
	}

	startTime := time.Now()
	writer := &LoggingWriter{
		ResponseWriter: rw,
		logRecord: LogRecord{
			Time:          startTime.UTC(),
			Ip:            ip,
			Method:        r.Method,
			Uri:           r.RequestURI,
			Username:      username,
			Protocol:      r.Proto,
			Host:          r.Host,
			Status:        0,
			Size:          0,
			ElapsedTime:   time.Duration(0),
			RequestHeader: r.Header,
		},
	}

	if h.logBefore {
		writer.SetCustomLogRecord("at", "before")
		h.logger.LogContext(r.Context(), writer.logRecord)
	}

	ctx := context.WithValue(r.Context(), ctxLoggerKey, writer)
	r = r.WithContext(ctx)
	h.handler.ServeHTTP(writer, r)
	finishTime := time.Now()

	writer.logRecord.Time = finishTime.UTC()
	writer.logRecord.ElapsedTime = finishTime.Sub(startTime)

	if h.logBefore {
		writer.SetCustomLogRecord("at", "after")
	}
	h.logger.LogContext(r.Context(), writer.logRecord)
}
