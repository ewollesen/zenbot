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

package memory

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ewollesen/zenbot/queue"
)

type memQueue struct {
	mu sync.Mutex
	q  [][]byte
}

var _ queue.Queue = (*memQueue)(nil)

func New() *memQueue {
	return &memQueue{
		q: [][]byte{},
	}
}

func (q *memQueue) Clear() error {
	q.mu.Lock()
	q.q = [][]byte{}
	q.mu.Unlock()
	return nil
}

func (q *memQueue) DequeueN(n int) (
	removed [][]byte, num_left int, err error) {

	q.mu.Lock()
	defer q.mu.Unlock()

	last := min(len(q.q), n)
	removed = q.q[:last]
	q.q = q.q[last:]

	return removed, len(q.q), nil
}

func (q *memQueue) Enqueue(datum []byte) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	pos, err := q.position(datum)
	if err != nil {
		return -1, err
	}
	if pos >= 0 {
		return -1, queue.AlreadyEnqueued.NewWith(
			fmt.Sprintf("%+v in position %d", datum, pos+1),
			queue.SetPosition(pos))
	}

	q.q = append(q.q, datum)

	return len(q.q) - 1, nil
}

// TODO: consider returning an interator... is that doable in go?
func (q *memQueue) Iter(fn func(int, []byte) bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.q {
		if fn(i, item) {
			break
		}
	}

	return nil
}

func (q *memQueue) Position(datum []byte) (pos int, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.position(datum)
}

// ensure that you're holding q.mu before calling!
func (q *memQueue) position(datum []byte) (pos int, err error) {
	for pos, candidate := range q.q {
		if bytes.Equal(candidate, datum) {
			return pos, nil
		}
	}

	return -1, nil
}

func (q *memQueue) Remove(item []byte) (err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	pos, err := q.position(item)
	if err != nil {
		return err
	}
	if pos < 0 {
		return queue.NotFound.New("")
	}

	if pos == 0 {
		q.q = q.q[1:]
	} else if pos < len(q.q)-1 {
		q.q = append(q.q[:pos], q.q[pos+1:]...)
	} else {
		q.q = q.q[:pos]
	}

	return nil
}

func (q *memQueue) Size() (int, error) {
	return len(q.q), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
