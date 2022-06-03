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

package lock

import (
	"context"
)

// New creates a Lock and populates its channel. Once constructed, callers can use the Lock(ctx) method to lock a
// portion of code. Unlike the sync.Mutex implementation, this allows a Lock operation to be deadlined or cancelled
// using the provided context.
func New() Lock {
	lock := Lock{ch: make(chan bool, 1)}
	lock.ch <- true

	return lock
}

// Lock encapsulates the channel used to provide locking and unlocking capabilities. We don't want this channel
// exposed as someone may interact with it in malicious ways.
type Lock struct {
	ch chan bool
}

// Lock attempts to obtain the lock given the provided context. This call can be time-bound using the
// context.WithDeadline method call to decorate the context with a halting point.
func (l Lock) Lock(ctx context.Context) (unlock func(), err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-l.ch:
		return func() { l.ch <- v }, nil
	}
}
