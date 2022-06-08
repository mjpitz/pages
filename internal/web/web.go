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

package web

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

//go:generate npm install
//go:generate npm run build

//go:embed dist/pages.js
var pagesjs string

func Handler() http.HandlerFunc {
	start := time.Now()

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "pages.js", start, strings.NewReader(pagesjs))
	}
}
