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

package commands

import (
	"context"
	"mime"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"code.pitz.tech/mya/pages/internal"
	"code.pitz.tech/mya/pages/internal/git"

	"github.com/mjpitz/myago/config"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

type HostConfig struct {
	internal.ServerConfig
	Git      git.Config `json:"git"`
	SiteFile string     `json:"site_file" usage:"configure multiple sites using a single file"`
}

var (
	hostConfig = &HostConfig{
		ServerConfig: internal.ServerConfig{
			Public:  internal.BindConfig{Address: "0.0.0.0:8080"},
			Private: internal.BindConfig{Address: "0.0.0.0:8081"},
		},
	}

	Host = &cli.Command{
		Name:      "host",
		Usage:     "Host the web page content",
		UsageText: "pages host",
		Flags:     flagset.ExtractPrefix("pages", hostConfig),
		Action: func(ctx *cli.Context) error {
			log := zaputil.Extract(ctx.Context)

			// additional mime-types that need to be explicitly registered
			_ = mime.AddExtensionType(".woff2", "application/font-woff2")
			_ = mime.AddExtensionType(".woff", "application/font-woff")
			_ = mime.AddExtensionType(".ttf", "font/ttf")
			_ = mime.AddExtensionType(".yaml", "application/yaml")
			_ = mime.AddExtensionType(".yml", "application/yaml")
			_ = mime.AddExtensionType(".json", "application/json")

			server, err := internal.NewServer(ctx.Context, hostConfig.ServerConfig)
			if err != nil {
				return err
			}

			endpointConfig := git.EndpointConfig{
				Sites: make(map[string]*git.Config),
			}

			if hostConfig.SiteFile == "" {
				endpointConfig.Sites["*"] = &hostConfig.Git
			} else {
				err = config.Load(ctx.Context, &endpointConfig, hostConfig.SiteFile)
				if err != nil {
					return err
				}
			}

			endpoint, err := git.NewEndpoint(ctx.Context, endpointConfig)
			if err != nil {
				return err
			}
			defer endpoint.Close()

			{ // git endpoints
				server.AdminMux.HandleFunc("/sync", endpoint.Sync).Methods(http.MethodPost)
				server.PublicMux.PathPrefix("/").HandlerFunc(endpoint.Lookup).Methods(http.MethodGet)
			}

			log.Info("serving",
				zap.String("public", hostConfig.Public.Address),
				zap.String("private", hostConfig.Private.Address))

			group, c := errgroup.WithContext(ctx.Context)
			group.Go(server.ListenAndServe)
			group.Go(func() error {
				return endpoint.SyncLoop(ctx.Context)
			})

			<-c.Done()

			shutdownTimeout := 30 * time.Second
			timeout, cancelTimeout := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancelTimeout()

			_ = server.Shutdown(timeout)
			_ = group.Wait()

			err = c.Err()
			if err != context.Canceled {
				return err
			}

			return nil
		},
		HideHelpCommand: true,
	}
)
