// Copyright (C) 2022  The pages authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package session

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"code.pitz.tech/mya/pages/internal/excludes"
)

func script(path string) string {
	return fmt.Sprintf(`<script async type='text/javascript' src='%s'></script>`, path)
}

type opt struct {
	jsPath   string
	excludes []excludes.Exclusion
}

// Option provides a way to configure elements of the Middleware.
type Option func(*opt)

// JavaScriptPath configures the URL for the JS snippet.
func JavaScriptPath(path string) Option {
	return func(o *opt) {
		o.jsPath = path
	}
}

// Exclusions appends the provided rules to the excludes list. Any path that matches an exclusion will not be measured.
func Exclusions(exclusions ...excludes.Exclusion) Option {
	return func(o *opt) {
		o.excludes = append(o.excludes, exclusions...)
	}
}

// Middleware injects a JavaScript snippet that establishes a session with the server.
func Middleware(opts ...Option) mux.MiddlewareFunc {
	o := opt{}
	for _, opt := range opts {
		opt(&o)
	}

	exclude := excludes.AnyExclusion(o.excludes...)
	injection := script(o.jsPath)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exclude(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			buffer := &bufferedResponseWriter{
				header: http.Header{},
				status: http.StatusOK,
				buffer: bytes.NewBuffer(nil),
			}

			next.ServeHTTP(buffer, r)

			for key := range buffer.header {
				w.Header().Set(key, buffer.header.Get(key))
			}

			contents := buffer.buffer.Bytes()

			if strings.Contains(buffer.header.Get("Content-Type"), "text/html") {
				contents = bytes.TrimSpace(contents)
				contents = bytes.TrimSuffix(contents, []byte("</html>"))
				contents = bytes.TrimSpace(contents)
				contents = bytes.TrimSuffix(contents, []byte("</body>"))

				contents = append(contents, []byte(injection)...)
				contents = append(contents, []byte("</body></html>")...)
			}

			w.Header().Set("Content-Length", strconv.Itoa(len(contents)))
			w.WriteHeader(buffer.status)

			_, _ = w.Write(contents)
		})
	}
}

type bufferedResponseWriter struct {
	header http.Header
	status int
	buffer *bytes.Buffer
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *bufferedResponseWriter) Write(data []byte) (n int, err error) {
	return w.buffer.Write(data)
}

var _ http.ResponseWriter = &bufferedResponseWriter{}
