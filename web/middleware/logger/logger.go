package logger

// A custom chi logger middleware, that adds the User-Agent to the logs.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"golang.org/x/net/context"
)

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white.
//
// Logger prints a request ID if one is provided.
func Logger(next chi.Handler) chi.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(ctx)
		prefix := requestPrefix(reqID, r)
		lw := wrapWriter(w)

		t1 := time.Now()
		defer func() {
			t2 := time.Now()
			printRequest(prefix, reqID, lw, t2.Sub(t1))
		}()

		next.ServeHTTPC(ctx, lw, r)
	}

	return chi.HandlerFunc(fn)
}

func requestPrefix(reqID string, r *http.Request) *bytes.Buffer {
	buf := &bytes.Buffer{}

	rScheme := "http://"
	if r.TLS != nil {
		rScheme = "https://"
	}
	rURL := fmt.Sprintf("%s%s%s", rScheme, r.Host, r.RequestURI)

	if reqID != "" {
		cW(buf, nYellow, "[%s] ", reqID)
	}
	cW(buf, nCyan, "\"")
	cW(buf, bMagenta, "%s ", r.Method)
	cW(buf, nCyan, "%s %s\" ", rURL, r.Proto)
	buf.WriteString("from ")
	buf.WriteString(r.RemoteAddr)
	buf.WriteString(" - ")
	buf.WriteString(r.Header.Get("User-Agent"))
	buf.WriteString(" - ")

	return buf
}

func printRequest(buf *bytes.Buffer, reqID string, w writerProxy, dt time.Duration) {
	status := w.Status()
	if status == middleware.StatusClientClosedRequest {
		cW(buf, bRed, "[disconnected]")
	} else {
		switch {
		case status == basicWriterUnknownStatus:
			cW(buf, bBlue, "200") // Implicit code sent by net pkg.
		case status < 200:
			cW(buf, bBlue, "%03d", status)
		case status < 300:
			cW(buf, bGreen, "%03d", status)
		case status < 400:
			cW(buf, bCyan, "%03d", status)
		case status < 500:
			cW(buf, bYellow, "%03d", status)
		default:
			cW(buf, bRed, "%03d", status)
		}
	}

	cW(buf, bBlue, " %dB", w.BytesWritten())

	buf.WriteString(" in ")
	if dt < 500*time.Millisecond {
		cW(buf, nGreen, "%s", dt)
	} else if dt < 5*time.Second {
		cW(buf, nYellow, "%s", dt)
	} else {
		cW(buf, nRed, "%s", dt)
	}

	log.Print(buf.String())
}

// writerProxy is a proxy around an http.ResponseWriter that allows you to hook
// into various parts of the response process.
type writerProxy interface {
	http.ResponseWriter
	// Status returns the HTTP status of the request, or 0 if one has not
	// yet been sent.
	Status() int
	// BytesWritten returns the total number of bytes sent to the client.
	BytesWritten() int
	// Tee causes the response body to be written to the given io.Writer in
	// addition to proxying the writes through. Only one io.Writer can be
	// tee'd to at once: setting a second one will overwrite the first.
	// Writes will be sent to the proxy before being written to this
	// io.Writer. It is illegal for the tee'd writer to be modified
	// concurrently with writes.
	Tee(io.Writer)
	// Unwrap returns the original proxied target.
	Unwrap() http.ResponseWriter
}

// WrapWriter wraps an http.ResponseWriter, returning a proxy that allows you to
// hook into various parts of the response process.
func wrapWriter(w http.ResponseWriter) writerProxy {
	_, cn := w.(http.CloseNotifier)
	_, fl := w.(http.Flusher)
	_, hj := w.(http.Hijacker)
	_, rf := w.(io.ReaderFrom)

	bw := basicWriter{ResponseWriter: w}
	if cn && fl && hj && rf {
		return &fancyWriter{bw}
	}
	return &bw
}

// basicWriter wraps a http.ResponseWriter that implements the minimal
// http.ResponseWriter interface.
type basicWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	bytes       int
	tee         io.Writer
}

func (b *basicWriter) WriteHeader(code int) {
	if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}
func (b *basicWriter) Write(buf []byte) (int, error) {
	b.WriteHeader(http.StatusOK)
	n, err := b.ResponseWriter.Write(buf)
	if b.tee != nil {
		_, err2 := b.tee.Write(buf[:n])
		// Prefer errors generated by the proxied writer.
		if err == nil {
			err = err2
		}
	}
	b.bytes += n
	return n, err
}
func (b *basicWriter) maybeWriteHeader() {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
}

const basicWriterUnknownStatus = -1

func (b *basicWriter) Status() int {
	if !b.wroteHeader {
		return basicWriterUnknownStatus
	}
	return b.code
}
func (b *basicWriter) BytesWritten() int {
	return b.bytes
}
func (b *basicWriter) Tee(w io.Writer) {
	b.tee = w
}
func (b *basicWriter) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}

// fancyWriter is a writer that additionally satisfies http.CloseNotifier,
// http.Flusher, http.Hijacker, and io.ReaderFrom. It exists for the common case
// of wrapping the http.ResponseWriter that package http gives you, in order to
// make the proxied object support the full method set of the proxied object.
type fancyWriter struct {
	basicWriter
}

func (f *fancyWriter) CloseNotify() <-chan bool {
	cn := f.basicWriter.ResponseWriter.(http.CloseNotifier)
	return cn.CloseNotify()
}
func (f *fancyWriter) Flush() {
	fl := f.basicWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}
func (f *fancyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.basicWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}
func (f *fancyWriter) ReadFrom(r io.Reader) (int64, error) {
	if f.basicWriter.tee != nil {
		return io.Copy(&f.basicWriter, r)
	}
	rf := f.basicWriter.ResponseWriter.(io.ReaderFrom)
	f.basicWriter.maybeWriteHeader()
	return rf.ReadFrom(r)
}

var _ http.CloseNotifier = &fancyWriter{}
var _ http.Flusher = &fancyWriter{}
var _ http.Hijacker = &fancyWriter{}
var _ io.ReaderFrom = &fancyWriter{}
