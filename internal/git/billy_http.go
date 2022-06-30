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
	"io/fs"
	"net/http"
	"os"
	"sync"

	"github.com/go-git/go-billy/v5"
)

// HTTP translates a billy.Filesystem into an http.FileSystem that can be used with the http.FileServer.
func HTTP(fs billy.Filesystem) http.FileSystem {
	return http.FS(&httpFS{fs: fs})
}

type httpFS struct {
	fs billy.Filesystem
}

func (f *httpFS) Open(name string) (fs.File, error) {
	fileInfo, err := f.fs.Stat(name)
	if err != nil {
		return nil, err
	}

	return &httpFile{
		once:     sync.Once{},
		fs:       f.fs,
		name:     name,
		fileInfo: fileInfo,
	}, nil
}

type httpFile struct {
	once sync.Once

	fs   billy.Filesystem
	name string

	fileInfo os.FileInfo
	file     billy.File
}

func (f *httpFile) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *httpFile) Stat() (fs.FileInfo, error) {
	return f.fileInfo, nil
}

func (f *httpFile) init() {
	f.file, _ = f.fs.Open(f.name)
}

func (f *httpFile) Read(bytes []byte) (int, error) {
	f.once.Do(f.init)
	return f.file.Read(bytes)
}

func (f *httpFile) Close() error {
	if f.file != nil {
		return f.file.Close()
	}

	return nil
}
