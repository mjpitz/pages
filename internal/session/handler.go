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
	"go.uber.org/zap"

	"github.com/mjpitz/myago/zaputil"
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
	log := zaputil.Extract(r.Context())

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade connection", zap.Error(err))
		return
	}

	defer func() { _ = conn.Close() }()

	path := r.URL.Path
	geoInfo := geoip.Extract(r.Context())

	active := metrics.PageSessionsActive.WithLabelValues(path, geoInfo.CountryCode)
	active.Inc()
	defer func() { active.Dec() }()

	start := time.Now()
	defer func() {
		metrics.PageSessionDuration.WithLabelValues(path, geoInfo.CountryCode).Observe(time.Since(start).Seconds())
	}()

	for {
		req := Request{}

		err = conn.ReadJSON(&req)
		if err != nil {
			return
		}
	}
}
