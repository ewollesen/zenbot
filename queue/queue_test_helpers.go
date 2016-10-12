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
	"encoding/json"
	"testing"

	"github.com/spacemonkeygo/spacelog"

	"github.com/ewollesen/zenbot/zentest"
)

var (
	logger = spacelog.GetLogger()
)

func CommonTestClear(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	test.AssertQueueSize(qut, 0)
	pos, err := qut.Enqueue(newQueueable("foo1", "bar1"))
	test.AssertNil(err)
	test.AssertEqual(pos, 0)
	test.AssertQueueSize(qut, 1)
	test.AssertNil(qut.Clear())
	test.AssertQueueSize(qut, 0)
	test.AssertNil(qut.Clear())
	test.AssertQueueSize(qut, 0)
}

func CommonTestDequeueN(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	qut.Enqueue(newQueueable("foo1", "bar1"))
	qut.Enqueue(newQueueable("foo2", "bar2"))
	qut.Enqueue(newQueueable("foo3", "bar3"))
	qut.Enqueue(newQueueable("foo4", "bar4"))
	qut.Enqueue(newQueueable("foo5", "bar5"))
	qut.Enqueue(newQueueable("foo6", "bar6"))

	removed, num_left, err := qut.DequeueN(1)
	test.AssertNil(err)
	test.AssertEqual(len(removed), 1)
	test.AssertEqual(removed[0].Key_, "foo1")
	test.AssertQueueSize(qut, 5)
	test.AssertEqual(num_left, 5)

	removed, num_left, err = qut.DequeueN(2)
	test.AssertNil(err)
	test.AssertEqual(len(removed), 2)
	test.AssertEqual(removed[0].Key_, "foo2")
	test.AssertEqual(removed[1].Key_, "foo3")
	test.AssertQueueSize(qut, 3)
	test.AssertEqual(num_left, 3)

	removed, num_left, err = qut.DequeueN(3)
	test.AssertNil(err)
	test.AssertEqual(len(removed), 3)
	test.AssertEqual(removed[0].Key_, "foo4")
	test.AssertEqual(removed[1].Key_, "foo5")
	test.AssertEqual(removed[2].Key_, "foo6")
	test.AssertQueueSize(qut, 0)
	test.AssertEqual(num_left, 0)

	removed, num_left, err = qut.DequeueN(4)
	test.AssertNil(err)
	test.AssertEqual(len(removed), 0)
	test.AssertQueueSize(qut, 0)
	test.AssertEqual(num_left, 0)
}

func CommonTestEnqueue(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	pos, err := qut.Enqueue(newQueueable("foo1", "bar1"))
	test.AssertNil(err)
	test.AssertEqual(pos, 0)
	test.AssertQueueSize(qut, 1)

	pos, err = qut.Enqueue(newQueueable("foo2", "bar2"))
	test.AssertNil(err)
	test.AssertQueueSize(qut, 2)

	pos, err = qut.Enqueue(newQueueable("foo2", "bar2"))
	test.AssertErrorContainedBy(err, AlreadyEnqueued)
	test.AssertQueueSize(qut, 2)
}

func CommonTestIter(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	q1 := newQueueable("foo1", "bar1")
	q2 := newQueueable("foo2", "bar2")
	q3 := newQueueable("foo3", "bar3")
	q4 := newQueueable("foo4", "bar4")
	q5 := newQueueable("foo5", "bar5")
	q6 := newQueueable("foo6", "bar6")

	qut.Enqueue(q1)
	qut.Enqueue(q2)
	qut.Enqueue(q3)
	qut.Enqueue(q4)
	qut.Enqueue(q5)
	qut.Enqueue(q6)

	num := 0
	f := func(pos int, datum *TestQueueable) bool {
		num++
		return false
	}

	test.AssertNil(qut.Iter(f))
	test.AssertEqual(num, 6)
	test.AssertQueueContains(qut, q1, q2, q3, q4, q5, q6)
}

func CommonTestPosition(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	q1 := newQueueable("foo", "bar")
	pos, err := qut.Position(q1)
	test.AssertNil(err)
	test.AssertEqual(pos, -1)

	qut.Enqueue(q1)
	pos, err = qut.Position(q1)
	test.AssertNil(err)
	test.AssertEqual(pos, 0)

	q2 := newQueueable("foo2", "bar2")
	qut.Enqueue(q2)
	pos, err = qut.Position(q2)
	test.AssertNil(err)
	test.AssertEqual(pos, 1)
}

func CommonTestRemove(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	test.AssertQueueSize(qut, 0)
	queueable := newQueueable("foo", "bar")
	qut.Enqueue(queueable)
	test.AssertQueueSize(qut, 1)
	err := qut.Remove(queueable)
	test.AssertNil(err)
	test.AssertEmpty(qut)
}

func CommonTestRemoveFirst(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	q1 := newQueueable("foo1", "bar1")
	q2 := newQueueable("foo2", "bar2")
	q3 := newQueueable("foo3", "bar3")

	qut.Enqueue(q1)
	qut.Enqueue(q2)
	qut.Enqueue(q3)
	test.AssertQueueSize(qut, 3)
	err := qut.Remove(q1)
	test.AssertNil(err)
	test.AssertQueueSize(qut, 2)
	test.AssertQueueContains(qut, q3, q2)
	test.AssertQueueDoesNotContain(qut, q1)
}

func CommonTestRemoveMiddle(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	q1 := newQueueable("foo1", "bar1")
	q2 := newQueueable("foo2", "bar2")
	q3 := newQueueable("foo3", "bar3")

	qut.Enqueue(q1)
	qut.Enqueue(q2)
	qut.Enqueue(q3)
	test.AssertQueueSize(qut, 3)
	err := qut.Remove(q2)
	test.AssertNil(err)
	test.AssertQueueSize(qut, 2)
	test.AssertQueueContains(qut, q3, q1)
	test.AssertQueueDoesNotContain(qut, q2)
}

func CommonTestRemoveLast(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	q1 := newQueueable("foo1", "bar1")
	q2 := newQueueable("foo2", "bar2")
	q3 := newQueueable("foo3", "bar3")

	qut.Enqueue(q1)
	qut.Enqueue(q2)
	qut.Enqueue(q3)
	test.AssertQueueSize(qut, 3)
	err := qut.Remove(q3)
	test.AssertNil(err)
	test.AssertQueueSize(qut, 2)
	test.AssertQueueContains(qut, q1, q2)
	test.AssertQueueDoesNotContain(qut, q3)
}

func CommonTestSize(t *testing.T, queue_under_test Queue) {
	test := newQueueTest(t)
	qut := NewTestQueue(queue_under_test)

	_, err := qut.Enqueue(newQueueable("foo", "bar"))
	test.AssertNil(err)
	test.AssertQueueSize(qut, 1)
}

//
// Helpers
//

type TestQueueable struct {
	Key_  string `json:"key"`
	Value string `json:"value"`
}

func newQueueable(key, value string) *TestQueueable {
	return &TestQueueable{
		Key_:  key,
		Value: value,
	}
}

type queueTest struct {
	*zentest.ZenTest
}

func newQueueTest(t *testing.T) *queueTest {
	return &queueTest{
		ZenTest: zentest.New(t),
	}
}

func (t *queueTest) AssertEmpty(queue *TestQueue) {
	t.AssertQueueSize(queue, 0)
}

func (t *queueTest) AssertQueueSize(queue *TestQueue, expected int) {
	actual, err := queue.Size()
	t.AssertNil(err)
	t.AssertEqual(actual, expected)
}

func (t *queueTest) AssertQueueContains(queue *TestQueue, data ...*TestQueueable) {
	found := 0
	for _, datum := range data {
		queue.Iter(func(i int, x *TestQueueable) bool {
			if x.Key_ == datum.Key_ && x.Value == datum.Value {
				found++
				return true
			}
			return false
		})
	}
	t.AssertEqual(found, len(data))
}

func (t *queueTest) AssertQueueDoesNotContain(queue *TestQueue,
	datum *TestQueueable) {

	queue.Iter(func(i int, x *TestQueueable) bool {
		if x.Key_ == datum.Key_ && x.Value == datum.Value {
			t.Assert(false)
		}
		return false
	})
}

type TestQueue struct {
	q Queue
}

func NewTestQueue(q Queue) *TestQueue {
	return &TestQueue{
		q: q,
	}
}

func (q *TestQueue) Clear() error {
	return q.q.Clear()
}

func (q *TestQueue) DequeueN(n int) (
	taken []TestQueueable, num_left int, err error) {

	takens_bytes, num_left, err := q.q.DequeueN(n)
	if err != nil {
		return nil, -1, err
	}

	for _, taken_bytes := range takens_bytes {
		tq := TestQueueable{}
		err := json.Unmarshal(taken_bytes, &tq)
		if err != nil {
			return nil, -1, err
		}
		taken = append(taken, tq)
	}

	return taken, num_left, nil
}

func (q *TestQueue) Enqueue(datum *TestQueueable) (int, error) {
	datum_bytes, err := json.Marshal(datum)
	if err != nil {
		return -1, err
	}

	return q.q.Enqueue(datum_bytes)
}

func (q *TestQueue) Iter(fn func(int, *TestQueueable) bool) error {
	return q.q.Iter(func(index int, datum []byte) bool {
		tq := &TestQueueable{}
		logger.Errore(json.Unmarshal(datum, tq))
		return fn(index, tq)
	})
}

func (q *TestQueue) Position(tq *TestQueueable) (int, error) {
	tq_bytes, err := json.Marshal(tq)
	if err != nil {
		return -1, err
	}
	return q.q.Position(tq_bytes)
}

func (q *TestQueue) Remove(tq *TestQueueable) error {
	tq_bytes, err := json.Marshal(tq)
	if err != nil {
		return err
	}
	return q.q.Remove(tq_bytes)
}

func (q *TestQueue) Size() (int, error) {
	return q.q.Size()
}
