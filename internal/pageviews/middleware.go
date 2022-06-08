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

package pageviews

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mjpitz/pages/internal/excludes"
	"github.com/mjpitz/pages/internal/geoip"
	"github.com/mjpitz/pages/internal/metrics"
)

type opt struct {
	excludes []excludes.Exclusion
}

// Option provides a way to configure elements of the Middleware.
type Option func(*opt)

// Exclusions appends the provided rules to the excludes list. Any path that matches an exclusion will not be measured.
func Exclusions(exclusions ...excludes.Exclusion) Option {
	return func(o *opt) {
		o.excludes = append(o.excludes, exclusions...)
	}
}

// Middleware produces an HTTP middleware function that reports page views.
func Middleware(opts ...Option) mux.MiddlewareFunc {
	o := opt{}
	for _, opt := range opts {
		opt(&o)
	}

	exclude := excludes.AnyExclusion(o.excludes...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exclude(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			info := geoip.Extract(r.Context())

			d := &writer{w, http.StatusOK}
			defer func() {
				if d.statusCode != http.StatusNotFound {
					metrics.PageViewCount.WithLabelValues(r.URL.Path, r.Referer(), info.CountryCode).Inc()
				}
			}()

			next.ServeHTTP(d, r)
		})
	}
}

type writer struct {
	writer     http.ResponseWriter
	statusCode int
}

func (w *writer) Header() http.Header {
	return w.writer.Header()
}

func (w *writer) Write(bytes []byte) (int, error) {
	return w.writer.Write(bytes)
}

func (w *writer) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.writer.WriteHeader(statusCode)
}
