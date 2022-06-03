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

FROM alpine:3.12

WORKDIR /opt/pages
RUN apk add -U tzdata ca-certificates gzip

ARG TARGETOS
ARG TARGETARCH
COPY dist/pages_${TARGETOS}_${TARGETARCH}.tar.gz /opt/pages/bin/pages.tar.gz
RUN tar x -C bin -z -f bin/pages.tar.gz && \
    rm bin/pages.tar.gz && \
    mv bin/LICENSE LICENSE

ENTRYPOINT [ "/opt/pages/bin/pages", "host" ]
