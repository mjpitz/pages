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

package internal

import (
	"net"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mjpitz/pages/internal/metrics"
)

// GeoIP provides an abstraction for looking up a country code from an IP address.
type GeoIP interface {
	Lookup(ip string) (countryCode string)
}

type emptyGeoIP struct{}

func (e emptyGeoIP) Lookup(ip string) (countryCode string) {
	return ""
}

// Matcher defines an abstraction for matching paths.
type Matcher func(s string) bool

// AnyMatcher returns a matcher who returns true if any of the provided matchers match the string.
func AnyMatcher(matchers ...Matcher) Matcher {
	return func(s string) bool {
		for _, matcher := range matchers {
			if matcher(s) {
				return true
			}
		}

		return false
	}
}

// AssetMatcher returns a matcher who returns true when an asset file is requested.
func AssetMatcher() Matcher {
	return func(s string) bool {
		return path.Ext(s) != ""
	}
}

// PrefixMatcher returns a matcher who returns true if the string matches the provided prefix.
func PrefixMatcher(prefix string) Matcher {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}

// RegexMatcher returns a matcher who returns true if the string matches the provided regular expression pattern.
func RegexMatcher(pattern string) Matcher {
	exp := regexp.MustCompile(pattern)

	return func(s string) bool {
		return exp.MatchString(s)
	}
}

// Middleware produces an HTTP middleware function that reports page views.
func Middleware(geoIP GeoIP, excludes ...Matcher) mux.MiddlewareFunc {
	exclude := AnyMatcher(excludes...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exclude(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			countryCode := geoIP.Lookup(clientIP)

			d := &writer{w, http.StatusOK}
			defer func() {
				if d.statusCode != http.StatusNotFound {
					metrics.PageViewCount.WithLabelValues(r.URL.Path, r.Referer(), countryCode).Inc()
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
