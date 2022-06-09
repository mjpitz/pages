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
	"net/http"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
	"github.com/mjpitz/pages/internal"
	"github.com/mjpitz/pages/internal/git"
)

type HostConfig struct {
	internal.ServerConfig
	Git git.Config `json:"git"`
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

			gitService := git.NewService(hostConfig.Git)
			err := gitService.Load(ctx.Context)
			if err != nil {
				return err
			}

			server, err := internal.NewServer(ctx.Context, hostConfig.ServerConfig)
			if err != nil {
				return err
			}

			{
				server.AdminMux.
					HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
						err = gitService.Sync(r.Context())
						if err != nil {
							log.Error(err.Error())
							http.Error(w, "", http.StatusInternalServerError)
							return
						}
					}).
					Methods(http.MethodPost)
			}

			server.PublicMux.PathPrefix("/").Handler(http.FileServer(git.HTTP(gitService.FS))).Methods(http.MethodGet)

			log.Info("serving",
				zap.String("public", hostConfig.Public.Address),
				zap.String("private", hostConfig.Private.Address))

			group, c := errgroup.WithContext(ctx.Context)
			group.Go(server.ListenAndServe)

			<-c.Done()

			_ = server.Shutdown(context.Background())
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
