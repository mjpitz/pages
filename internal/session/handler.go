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
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mjpitz/pages/internal/geoip"
	"github.com/mjpitz/pages/internal/metrics"
)

type Request struct {
	ID       string
	FullName string
}

func Handler() *Handle {
	return &Handle{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

type Handle struct {
	upgrader websocket.Upgrader
}

func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	defer func() { _ = conn.Close() }()

	if err != nil {
		// log
		return
	}

	path := r.URL.Path
	geoInfo := geoip.Extract(r.Context())

	start := time.Now()
	defer func() {
		metrics.PageSessionDuration.WithLabelValues(path, geoInfo.CountryCode).Observe(time.Since(start).Seconds())
	}()

	for {
		req := Request{}

		err = conn.ReadJSON(&req)
		if err != nil {
			// log
			return
		}
	}
}
