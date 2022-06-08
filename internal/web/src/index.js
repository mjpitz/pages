/**
 * Copyright (C) 2022  The pages authors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

import {ulid} from 'ulid';

const sessionPrefix = "/_session";
const domain = window.location.host;

async function main() {
    const ws = new WebSocket(`ws://${domain}${sessionPrefix}${window.location.pathname}`);
    const message = `{"ID":"${ulid()}","FullName":"ping"}`

    let id = null;
    ws.onopen = () => id = setInterval(() => ws.send(message), 5000);
    ws.onclose = () => clearInterval(id);
}

main().catch(err => console.error(err))
