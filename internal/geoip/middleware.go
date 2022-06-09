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

package geoip

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mjpitz/myago"
)

var key = myago.ContextKey("geoip.info")

func Extract(ctx context.Context) Info {
	val := ctx.Value(key)
	v, ok := val.(Info)

	if val == nil || !ok {
		return Info{}
	}

	return v
}

func Middleware(geoip Interface) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), key, geoip.Lookup(clientIP))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
