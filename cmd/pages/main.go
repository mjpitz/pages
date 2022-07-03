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

package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"code.pitz.tech/mya/pages/internal/commands"

	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/lifecycle"
	"github.com/mjpitz/myago/zaputil"
)

var version = ""
var commit = ""
var date = time.Now().Format(time.RFC3339)

type GlobalConfig struct {
	Log zaputil.Config `json:"log"`
}

func main() {
	compiled, _ := time.Parse(time.RFC3339, date)

	cfg := &GlobalConfig{
		Log: zaputil.DefaultConfig(),
	}

	app := &cli.App{
		Name:      "pages",
		Usage:     "Provides hosting and anonymized prometheus metrics for front-end web applications",
		UsageText: "pages [options] <command>",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Flags:     flagset.ExtractPrefix("pages", cfg),
		Commands: []*cli.Command{
			commands.Host,
			commands.Version,
		},
		Before: func(ctx *cli.Context) error {
			ctx.Context = zaputil.Setup(ctx.Context, cfg.Log)
			ctx.Context = lifecycle.Setup(ctx.Context)

			return nil
		},
		After: func(ctx *cli.Context) error {
			lifecycle.Resolve(ctx.Context)
			lifecycle.Shutdown(ctx.Context)

			return nil
		},
		Compiled:             compiled,
		Copyright:            fmt.Sprintf("Copyright %d The pages authors - All Rights Reserved\n", compiled.Year()),
		HideVersion:          true,
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Metadata: map[string]interface{}{
			"arch":       runtime.GOARCH,
			"compiled":   date,
			"go_version": strings.TrimPrefix(runtime.Version(), "go"),
			"os":         runtime.GOOS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
