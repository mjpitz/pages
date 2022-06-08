#!/usr/bin/env bash
# Copyright (C) 2022  The pages authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

set -e -o pipefail

go mod download
go mod verify

go generate ./...

if [[ -z "${VERSION}" ]]; then
	goreleaser --snapshot --skip-publish --rm-dist
else
	goreleaser
fi

rm -rf "$(pwd)/pages"

os=$(uname | tr '[:upper:]' '[:lower:]')
arch="$(uname -m)"
if [[ "$arch" == "x86_64" ]]; then
	ln -s "$(pwd)/dist/pages_${os}_amd64_v1/pages" "$(pwd)/pages"
elif [[ "$arch" == "aarch64" ]]; then
	ln -s "$(pwd)/dist/pages_${os}_arm64/pages" "$(pwd)/pages"
fi
