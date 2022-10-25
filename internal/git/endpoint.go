package git

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/mjpitz/myago/clocks"
	"github.com/mjpitz/myago/zaputil"
)

type EndpointConfig struct {
	Sites map[string]*Config `json:"sites"`
}

func NewEndpoint(ctx context.Context, multi EndpointConfig) (endpoint *Endpoint, err error) {
	clock := clocks.Extract(ctx)
	log := zaputil.Extract(ctx)

	endpoint = &Endpoint{
		sites: make(map[string]*entry),
	}

	for domain, cfg := range multi.Sites {
		log.Info("loading site",
			zap.String("domain", domain),
			zap.String("tag", cfg.Tag),
			zap.String("branch", cfg.Branch),
			zap.Duration("sync_interval", cfg.SyncInterval),
		)

		endpoint.sites[domain] = &entry{
			config: *cfg,
		}

		endpoint.sites[domain].service, err = NewService(*cfg)
		if err != nil {
			return nil, err
		}

		err = endpoint.sites[domain].service.Load(ctx)
		if err != nil {
			return nil, err
		}
	}

	for domain, cfg := range multi.Sites {
		endpoint.sites[domain].ticker = clock.NewTicker(cfg.SyncInterval)
	}

	return endpoint, nil
}

type entry struct {
	config  Config
	service *Service
	ticker  clockwork.Ticker
}

type Endpoint struct {
	sites map[string]*entry
}

func (e *Endpoint) lookupSite(r *http.Request) *entry {
	if len(e.sites) == 1 && e.sites["*"] != nil {
		return e.sites["*"]
	}

	url, err := url.Parse(r.RequestURI)
	if err != nil {
		return nil
	}

	domain := url.Hostname()
	switch {
	case r.Header.Get("Host") != "":
		domain = r.Header.Get("Host")
	case r.Header.Get("X-Forwarded-Host") != "":
		domain = r.Header.Get("X-Forwarded-Host")
	}

	return e.sites[domain]
}

func (e *Endpoint) Sync(w http.ResponseWriter, r *http.Request) {
	entry := e.lookupSite(r)

	if entry == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err := entry.service.Sync(r.Context())
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (e *Endpoint) Lookup(w http.ResponseWriter, r *http.Request) {
	entry := e.lookupSite(r)

	if entry == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// Lookup file
	values := r.URL.Query()

	if values.Get("go-get") == "1" {
		// peak for go-get requests
		_, name := path.Split(r.URL.Path)
		index := path.Join(r.URL.Path, "index.html")

		// if index.html exists, then use that
		info, err := entry.service.FS.Stat(index)
		if err == nil {
			file, err := entry.service.FS.Open(index)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			defer file.Close()

			http.ServeContent(w, r, name, info.ModTime(), file)
			return
		}
	}

	http.FileServer(HTTP(entry.service.FS)).ServeHTTP(w, r)
}

func (e *Endpoint) SyncLoop(ctx context.Context) error {
	clock := clocks.Extract(ctx)

	timer := clock.NewTicker(time.Second)
	defer timer.Stop()

	keys := make([]string, 0, len(e.sites))
	for key := range e.sites {
		keys = append(keys, key)
	}

	group := &errgroup.Group{}

	for {
		for i := 0; i < len(keys); i++ {
			select {
			case <-ctx.Done():
				// context cancelled / hit deadline
				return ctx.Err()

			case <-e.sites[keys[i]].ticker.Chan():
				service := e.sites[keys[i]].service

				// ticker expired, sync the site, check the next
				// sync happens in a background thread to avoid contention on this loop
				group.Go(func() error {
					return service.Sync(ctx)
				})

			case <-timer.Chan():
				// times up, check the next
				continue
			}
		}
	}
}

func (e *Endpoint) Close() error {
	for _, site := range e.sites {
		site.ticker.Stop()
	}

	return nil
}
