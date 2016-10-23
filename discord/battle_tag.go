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

package discord

import (
	"encoding/json"
	"regexp"

	"github.com/ewollesen/zenbot/cache"
	"github.com/ewollesen/zenbot/queue"
)

var (
	btagRe     = regexp.MustCompile("^\\pL[\\pL\\pN]{2,11}#\\d{1,7}$")
	btagTextRe = regexp.MustCompile("\\b\\pL[\\pL\\pN]{2,11}#\\d{1,7}\\b")
)

type BattleTagCache struct {
	c cache.Cache
}

func NewBattleTagCache(c cache.Cache) *BattleTagCache {
	return &BattleTagCache{
		c: c,
	}
}

func (c *BattleTagCache) Clear() error {
	return c.c.Clear()
}

func (c *BattleTagCache) Get(key string) (string, error) {
	value_bytes, err := c.c.Get(key)
	if err != nil {
		return "", err
	}
	return string(value_bytes), nil
}

func (c *BattleTagCache) Iter(fn func(key string, btag string) bool) {
	c.c.Iter(func(key string, value []byte) bool {
		return fn(key, string(value))
	})
}

func (c *BattleTagCache) Set(key string, value string) error {
	return c.c.Set(key, []byte(value))
}

type BattleTagQueue struct {
	q queue.Queue
}

func newBattleTagQueue(q queue.Queue) *BattleTagQueue {
	return &BattleTagQueue{q: q}
}

func (q *BattleTagQueue) Clear() error {
	return q.q.Clear()
}

func (q *BattleTagQueue) DequeueN(n int) (
	taken []*userBattleTag, num_left int, err error) {

	takens_bytes, num_left, err := q.q.DequeueN(n)
	if err != nil {
		return nil, -1, err
	}

	for _, taken_bytes := range takens_bytes {
		tq := &userBattleTag{}
		err := json.Unmarshal(taken_bytes, tq)
		if err != nil {
			return nil, -1, err
		}
		taken = append(taken, tq)
	}

	return taken, num_left, nil
}

func (q *BattleTagQueue) Enqueue(datum *userBattleTag) (int, error) {
	datum_bytes, err := json.Marshal(datum)
	if err != nil {
		return -1, err
	}

	return q.q.Enqueue(datum_bytes)
}

func (q *BattleTagQueue) Iter(fn func(int, *userBattleTag) bool) error {
	return q.q.Iter(func(index int, datum []byte) bool {
		tq := &userBattleTag{}
		logger.Errore(json.Unmarshal(datum, tq))
		return fn(index, tq)
	})
}

func (q *BattleTagQueue) Position(ubt *userBattleTag) (int, error) {
	tq_bytes, err := json.Marshal(ubt)
	if err != nil {
		return -1, err
	}
	return q.q.Position(tq_bytes)
}

func (q *BattleTagQueue) Remove(ubt *userBattleTag) error {
	ubt_bytes, err := json.Marshal(ubt)
	if err != nil {
		return err
	}
	return q.q.Remove(ubt_bytes)
}

func (q *BattleTagQueue) Size() (int, error) {
	return q.q.Size()
}
