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

package geoip

import (
	"net"

	"github.com/IncSW/geoip2"
	"github.com/pkg/errors"
)

// Config contains the set of configuration options needed to configure looking up a users location.
type Config struct {
	DB string `json:"db" usage:"path of the Maxmind GeoIP2Country database"`
}

// Open uses information from the Config to determine which type of lookup to use.
func (c Config) Open() (Interface, error) {
	if c.DB != "" {
		return CountryLite(c.DB)
	}

	return Empty{}, nil
}

type Info struct {
	CountryCode string
}

type Interface interface {
	Lookup(ip string) Info
}

type Empty struct{}

func (e Empty) Lookup(ip string) Info {
	return Info{}
}

func CountryLite(file string) (*Maxmind, error) {
	reader, err := geoip2.NewCountryReaderFromFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	return &Maxmind{reader}, nil
}

type Maxmind struct {
	reader *geoip2.CountryReader
}

func (m *Maxmind) Lookup(ip string) Info {
	record, err := m.reader.Lookup(net.ParseIP(ip))
	if err != nil {
		return Info{}
	}

	return Info{
		CountryCode: record.Country.ISOCode,
	}
}
