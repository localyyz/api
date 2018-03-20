package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/pressly/lg"
	"github.com/sirupsen/logrus"
)

// RequestLogger is a middleware for the github.com/sirupsen/logrus to log requests.
// It is equipt to handle recovery in case of panics and record the stack trace
// with a panic log-level.
func RequestLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	httpLogger := &lg.HTTPLogger{logger}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := httpLogger.NewLogEntry(r)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					entry.Panic(rec, debug.Stack())
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

					if lg.AlertFn != nil {
						lg.AlertFn(logrus.PanicLevel, fmt.Sprintf("\nPANIC: %+v\n. Check server log for stack trace.", rec))
					}
				}

				// Log the entry, the request is complete.
				entry.Write(ww.Status(), ww.BytesWritten(), t2.Sub(t1))
			}()

			r = r.WithContext(lg.WithLogEntry(r.Context(), entry))
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
