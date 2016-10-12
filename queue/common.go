// Copyright 2016 Eric Wollesen <ericw at xmtp dot net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"fmt"

	"github.com/spacemonkeygo/errors"
)

var (
	errorPosKey = errors.GenSym()

	Error           = errors.NewClass("queue")
	NotFound        = Error.NewClass("not found")
	AlreadyEnqueued = Error.NewClass("already enqueued")
)

type Queue interface {
	Clear() error
	DequeueN(n int) ([][]byte, int, error)
	Enqueue(datum []byte) (int, error)
	Iter(fn func(index int, datum []byte) (stop bool)) error
	Position(datum []byte) (int, error)
	Remove(datum []byte) error
	Size() (int, error)
}

func SetPosition(pos int) errors.ErrorOption {
	return errors.SetData(errorPosKey, pos)
}

func GetPosition(err error) (position int) {
	val, ok := errors.GetData(err, errorPosKey).(int)
	if !ok {
		fmt.Printf("not ok: %+v\n", val)
		return -69
	}
	return val
}
