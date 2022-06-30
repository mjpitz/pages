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

package git

import (
	"context"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/mjpitz/myago/zaputil"
)

// Config encapsulates the elements that can be configured about the git service.
type Config struct {
	URL      string `json:"url"      usage:"the git url used to clone the repository"`
	Branch   string `json:"branch"   usage:"the name of the git branch to clone"`
	Tag      string `json:"tag"      usage:"the name of the git tag to clone"`
	Username string `json:"username" usage:"the username used to authenticate with the git service"`
	Password string `json:"password" usage:"the password used to authenticate with the git service"`
}

// NewService constructs a Service that manages the underlying git repository.
func NewService(config Config) (*Service, error) {
	temp, err := os.MkdirTemp(os.TempDir(), "pages-*")
	if err != nil {
		return nil, err
	}

	options := &git.CloneOptions{
		URL: config.URL,
	}

	if config.Username != "" && config.Password != "" {
		options.Auth = &http.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}
	}

	switch {
	case config.Tag != "":
		options.ReferenceName = plumbing.NewTagReferenceName(config.Tag)
	case config.Branch != "":
		options.ReferenceName = plumbing.NewBranchReferenceName(config.Branch)
	}

	return &Service{
		options: options,
		Store:   memory.NewStorage(),
		FS:      osfs.New(temp),
	}, nil
}

// Service encapsulates operations that can be performed against the target git repository.
type Service struct {
	options    *git.CloneOptions
	Store      *memory.Storage
	FS         billy.Filesystem
	Repository *git.Repository
}

// Load initializes the git repository given the provided options. This _should_ only be called once.
func (s *Service) Load(ctx context.Context) (err error) {
	zaputil.Extract(ctx).Info("cloning", zap.String("url", s.options.URL))

	s.Repository, err = git.CloneContext(ctx, s.Store, s.FS, s.options)
	if err != nil {
		return errors.Wrap(err, "failed to clone repository")
	}

	return nil
}

// Sync pulls the underlying repository to ensure it's up-to-date.
func (s *Service) Sync(ctx context.Context) error {
	zaputil.Extract(ctx).Info("synchronizing", zap.String("url", s.options.URL))

	wt, err := s.Repository.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed to obtain worktree")
	}

	err = wt.PullContext(ctx, &git.PullOptions{
		ReferenceName: s.options.ReferenceName,
		SingleBranch:  s.options.SingleBranch,
		Depth:         s.options.Depth,
		Auth:          s.options.Auth,
		Force:         true,
	})

	switch {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
	case err != nil:
		zaputil.Extract(ctx).Error("failed to pull", zap.Error(err))
	}

	return nil
}
