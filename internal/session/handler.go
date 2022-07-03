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
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"code.pitz.tech/mya/pages/internal/geoip"
	"code.pitz.tech/mya/pages/internal/metrics"

	"github.com/mjpitz/myago/zaputil"
)

type Metric struct {
	Key   string
	Value any // float, bool, int, ...
}

type Measurements struct {
	Timestamp uint64 // seconds
	Metrics   []Metric
}

type Request struct {
	ID       string
	FullName string
	Data     []Measurements
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
	ctx := r.Context()
	log := zaputil.Extract(ctx)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade connection", zap.Error(err))
		return
	}

	defer func() { _ = conn.Close() }()

	u, err := url.Parse(r.RequestURI)
	if err != nil {
		return
	}

	domain := u.Hostname()
	switch {
	case r.Header.Get("Host") != "":
		domain = r.Header.Get("Host")
	case r.Header.Get("X-Forwarded-Host") != "":
		domain = r.Header.Get("X-Forwarded-Host")
	}

	path := u.Path
	geoInfo := geoip.Extract(ctx)

	active := metrics.PageSessionsActive.WithLabelValues(domain, path, geoInfo.CountryCode)
	active.Inc()
	defer func() { active.Dec() }()

	start := time.Now()
	defer func() {
		metrics.PageSessionDuration.WithLabelValues(domain, path, geoInfo.CountryCode).Observe(time.Since(start).Seconds())
	}()

	for {
		req := Request{}

		err = conn.ReadJSON(&req)
		if err != nil {
			return
		}
	}
}
