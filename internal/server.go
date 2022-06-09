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
	"context"
	"net"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	"github.com/mjpitz/myago/auth"
	basicauth "github.com/mjpitz/myago/auth/basic"
	httpauth "github.com/mjpitz/myago/auth/http"
	"github.com/mjpitz/myago/headers"
	"github.com/mjpitz/myago/livetls"
	"github.com/mjpitz/pages/internal/excludes"
	"github.com/mjpitz/pages/internal/geoip"
	"github.com/mjpitz/pages/internal/pageviews"
	"github.com/mjpitz/pages/internal/session"
	"github.com/mjpitz/pages/internal/web"
)

// AdminConfig encapsulates configuration for the administrative endpoints.
type AdminConfig struct {
	Prefix   string `json:"prefix" usage:"configure the prefix to use for admin endpoints" default:"/_admin" hidden:"true"`
	Username string `json:"username" usage:"specify the username used to authenticate requests with the admin endpoints" default:"admin"`
	Password string `json:"password" usage:"specify the password used to authenticate requests with the admin endpoints"`
}

// BindConfig defines the set of configuration options for setting up a server.
type BindConfig struct {
	Address string `json:"address" usage:"configure the bind address for the server"`
}

// ServerConfig defines configuration for a public and private interface.
type ServerConfig struct {
	Admin   AdminConfig    `json:"admin"`
	GeoIP   geoip.Config   `json:"geoip"`
	Session session.Config `json:"session"`
	TLS     livetls.Config `json:"tls"`
	Public  BindConfig     `json:"public"`
	Private BindConfig     `json:"private"`
}

// NewServer constructs a Server from it's associated configuration.
func NewServer(ctx context.Context, config ServerConfig) (*Server, error) {
	tls, err := livetls.New(ctx, config.TLS)
	if err != nil {
		return nil, err
	}

	ipdb, err := config.GeoIP.Open()
	if err != nil {
		return nil, err
	}

	private := mux.NewRouter()
	private.Handle("/metrics", promhttp.Handler())

	exclusions := []excludes.Exclusion{
		excludes.AssetExclusion(),
		excludes.PrefixExclusion(config.Admin.Prefix),
		excludes.PrefixExclusion(config.Session.Prefix),
	}

	public := mux.NewRouter()
	public.Use(
		func(next http.Handler) http.Handler { return headers.HTTP(next) },
		geoip.Middleware(ipdb),
		pageviews.Middleware(
			pageviews.Exclusions(exclusions...),
		),
	)

	if config.Session.Enable {
		public.Use(
			session.Middleware(
				session.Exclusions(exclusions...),
				session.JavaScriptPath(path.Join(config.Session.Prefix, "pages.js")),
			),
		)
	}

	admin := public.PathPrefix(config.Admin.Prefix).Subrouter()

	if config.Admin.Password != "" {
		authenticate := basicauth.Static(config.Admin.Username, config.Admin.Password)
		required := auth.Required()

		admin.Use(func(next http.Handler) http.Handler {
			return httpauth.Handler(next, authenticate, required)
		})
	}

	{
		var handler http.Handler = session.Handler()
		handler = http.StripPrefix(config.Session.Prefix, handler)

		session := public.PathPrefix(config.Session.Prefix).Subrouter()
		session.HandleFunc("/pages.js", web.Handler()).Methods(http.MethodGet)
		session.Handle("/", handler)
	}

	return &Server{
		AdminMux: admin,

		PublicMux: public,
		Public: &http.Server{
			TLSConfig: tls,
			Addr:      config.Public.Address,
			Handler:   promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, public),
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		},

		PrivateMux: private,
		Private: &http.Server{
			TLSConfig: tls,
			Addr:      config.Private.Address,
			Handler:   promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, private),
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		},
	}, nil
}

// Server hosts a Public and Private HTTP server.
type Server struct {
	AdminMux   *mux.Router
	PublicMux  *mux.Router
	Public     *http.Server
	PrivateMux *mux.Router
	Private    *http.Server
}

// Shutdown closes the underlying Public and Private HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	_ = s.Public.Shutdown(ctx)
	_ = s.Private.Shutdown(ctx)

	return nil
}

// ListenAndServe starts underlying Public and Private HTTP servers.
func (s *Server) ListenAndServe() error {
	var group errgroup.Group
	group.Go(s.Public.ListenAndServe)
	group.Go(s.Private.ListenAndServe)

	return group.Wait()
}
