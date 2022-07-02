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

(() => {
    // Event Queue

    let events = window._pagesEvents || [];
    window._pagesEvents = events;

    // Public API

    window.pages = {
        // observe - Queues a set of metrics (a measurement) for reporting to the server. The measurement is timestamped
        // since the batch is reported asynchronously. `metrics` should be an array of key-value pairs. Each value in
        // the array should be an object with a `key` and a `value` attribute. For example:
        // `{ key: "metric_name", value: 36.2 }`
        observe({metrics}) {
            events.push({
                Timestamp: Date.now(),
                Metrics: metrics.map((val) => {
                    // format keys so callers don't need to know the go specifics
                    val.Key = val.Key ? val.Key : val.key;
                    val.Value = val.Value ? val.Value : val.value;

                    delete val.key;
                    delete val.value;

                    return val;
                }),
            });
        },
    };

    // report - Takes a snapshot of the current events list and sends it along to the server for processing. If no
    // events are available, a heartbeat is sent instead. If no error occurs, we update the event slice to remove the
    // elements we emit.
    const report = ({sessionId, ws}) => {
        const checkpoint = events.length;

        let msg = {ID: sessionId, FullName: "record", Data: events.slice(0, checkpoint)};
        if (checkpoint === 0) {
            msg = {ID: sessionId, FullName: "heartbeat"};
        }

        ws.send(JSON.stringify(msg));

        if (checkpoint > 0) {
            events = events.slice(checkpoint);
            window._pagesEvents = events;
        }

        setTimeout(() => report({sessionId, ws}), 5000);
    };

    // connect - Establishes a new WebSocket to the backend. When the socket is closed, it will attempt to reconnect to
    // the backend. Once open, the system asynchronously reports on the metrics being emit.
    const connect = ({sessionId}) => {
        const secure = window.location.protocol === "https:";
        const protocol = secure ? "wss:" : "ws:";
        const sessionPrefix = "/_session";
        const domain = window.location.host + (secure ? ":443" : "");
        const path = window.location.pathname;

        const ws = new WebSocket(`${protocol}//${domain}${sessionPrefix}${path}`);

        ws.onclose = () => connect({sessionId});
        ws.onopen = () => report({sessionId, ws});
    };

    // kick everything off
    connect({sessionId: ulid()});
})()
