# pages

`pages` provides web application / static site hosting with built-in support for simple analytics via [Prometheus][] and
[Grafana][].

[Prometheus]: https://prometheus.io
[Grafana]: https://grafana.com

**Motivation**

As I work to reduce the number of Google services from my life, I found myself wanting a simpler alternative to my 
current hosting solution. Right now, my hosting is provided by GitHub pages and my analytics is provided by Google (as 
one did in th 2010s). But as we start to enforce more privacy rights through systems like GDPR, I find myself wanting a
simpler solution that doesn't require cookie disclosures.

## Quickstart with Docker

The easiest way to get started is with Docker. Your web content will be exposed on port `8080` while [metrics](#metrics)
will be available on `8081`.

```shell
docker run --rm -it \
  -e PAGES_GIT_URL=<your repo> \
  -e PAGES_GIT_BRANCH=gh-pages \
  -p 8080:8080 \
  -p 8081:8081 \
  img.pitz.tech/mya/pages
```

**Syncing the git repository**

You can instruct the server to reload the Git branch by curling this `/_admin/sync` endpoint.

```shell
curl -X POST http://localhost:8080/_admin/sync
```

This endpoint is exposed publicly so your CI solution can issue the command to cause the servers to update. 

## Metrics

All metrics are prefixed with the `pages` namespace. This makes it easy to narrow down to the specific metrics for the
system.

### pages_page_view_count

The number of page views for a given path and their associated referrer.

```text
# HELP pages_page_view_count the number of times a given page has been viewed and by what referrer
# TYPE pages_page_view_count counter
pages_page_view_count{country="",path="/charts/",referrer="http://localhost:8080/blog/"} 1
```

### pages_page_session_seconds

Histogram of how long users spend on a page.

```text
# HELP pages_page_session_seconds how long someone spent on a given page
# TYPE pages_page_session_seconds histogram
pages_page_session_seconds_bucket{country="",path="/",le="0.005"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.01"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.025"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.05"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.1"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.25"} 0
pages_page_session_seconds_bucket{country="",path="/",le="0.5"} 0
pages_page_session_seconds_bucket{country="",path="/",le="1"} 0
pages_page_session_seconds_bucket{country="",path="/",le="2.5"} 1
pages_page_session_seconds_bucket{country="",path="/",le="5"} 1
pages_page_session_seconds_bucket{country="",path="/",le="10"} 1
pages_page_session_seconds_bucket{country="",path="/",le="+Inf"} 1
pages_page_session_seconds_sum{country="",path="/"} 1.855976549
pages_page_session_seconds_count{country="",path="/"} 1
```
