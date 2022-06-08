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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespace = "pages"
	page      = "page"

	// PageViewCount tracks page views by how often they're loaded.
	PageViewCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: page,
			Name:      "view_count",
			Help:      "the number of times a given page has been viewed and by what referrer",
		},
		[]string{"path", "referrer", "country"},
	)

	PageSessionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: page,
			Name:      "session_seconds",
			Help:      "how long someone spent on a given page",
		},
		[]string{"path", "country"},
	)
)
