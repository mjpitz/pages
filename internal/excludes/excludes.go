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

package excludes

import (
	"path"
	"regexp"
	"strings"
)

// Exclusion defines an abstraction for matching paths.
type Exclusion func(s string) bool

// AnyExclusion returns a matcher who returns true if any of the provided matchers match the string.
func AnyExclusion(exclusions ...Exclusion) Exclusion {
	return func(s string) bool {
		for _, exclusion := range exclusions {
			if exclusion(s) {
				return true
			}
		}

		return false
	}
}

// AssetExclusion returns a matcher who returns true when an asset file is requested.
func AssetExclusion() Exclusion {
	return func(s string) bool {
		return path.Ext(s) != ""
	}
}

// PrefixExclusion returns a matcher who returns true if the string matches the provided prefix.
func PrefixExclusion(prefix string) Exclusion {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}

// RegexExclusion returns a matcher who returns true if the string matches the provided regular expression pattern.
func RegexExclusion(pattern string) Exclusion {
	exp := regexp.MustCompile(pattern)

	return func(s string) bool {
		return exp.MatchString(s)
	}
}
