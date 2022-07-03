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
	"net/url"

	"github.com/gorilla/mux"

	"code.pitz.tech/mya/pages/internal/excludes"
	"code.pitz.tech/mya/pages/internal/geoip"
	"code.pitz.tech/mya/pages/internal/metrics"
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

			url, err := url.Parse(r.RequestURI)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			domain := url.Hostname()
			switch {
			case r.Header.Get("Host") != "":
				domain = r.Header.Get("Host")
			case r.Header.Get("X-Forwarded-Host") != "":
				domain = r.Header.Get("X-Forwarded-Host")
			}

			path := url.Path
			referrer := r.Referer()
			info := geoip.Extract(r.Context())

			d := &writer{w, http.StatusOK}
			defer func() {
				if d.statusCode != http.StatusNotFound {
					metrics.PageViewCount.WithLabelValues(domain, path, referrer, info.CountryCode).Inc()
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
