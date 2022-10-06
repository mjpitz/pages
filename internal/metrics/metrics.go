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

	// by default, summaries give us counts and sums which we can use to compute an average (not great, but it can work)
	// in addition to the default information, we report on the following quantiles:
	defaultObjectives = map[float64]float64{
		0.25: 0, // lower quartile
		0.5:  0, // median
		0.75: 0, // upper quartile
		0.90: 0, // One 9
		0.99: 0, // Two 9s
	}

	// PageViewCount tracks page views by how often they're loaded.
	PageViewCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: page,
			Name:      "view_count",
			Help:      "the number of times a given page has been viewed and by what referrer",
		},
		[]string{"domain", "path", "referrer", "country"},
	)

	// PageSessionDuration tracks how long someone spends on an individual page.
	PageSessionDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  page,
			Name:       "session_seconds",
			Help:       "how long someone spent on a given page",
			Objectives: defaultObjectives,
		},
		[]string{"domain", "path", "country"},
	)

	// PageSessionsActive provides an approximation of the number of sessions that are currently observing the page.
	PageSessionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: page,
			Name:      "sessions_active",
			Help:      "the number of current sessions for a given page",
		},
		[]string{"domain", "path", "country"},
	)
)
