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

package redisqueue

import (
	"bytes"
	"fmt"

	"github.com/ewollesen/zenbot/queue"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"
	redis "gopkg.in/redis.v5"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

var _ queue.Queue = (*RedisQueue)(nil)

var (
	Error  = errors.NewClass("redisqueue")
	logger = spacelog.GetLogger()
)

func New(client *redis.Client, key string) *RedisQueue {
	return &RedisQueue{
		client: client,
		key:    key,
	}
}

func (q *RedisQueue) Clear() error {
	return q.client.LTrim(q.key, 1, 0).Err()
}

func (q *RedisQueue) DequeueN(n int) (taken [][]byte, num_left int, err error) {
	var size int64
	err = q.client.Watch(func(tx *redis.Tx) error {
		strs, err := tx.LRange(q.key, 0, int64(n-1)).Result()
		if err != nil {
			return err
		}

		size, err = tx.LLen(q.key).Result()
		if err != nil {
			return err
		}

		_, err = tx.Pipelined(func(pipe *redis.Pipeline) error {
			for _, str := range strs {
				taken = append(taken, []byte(str))
				err = pipe.LRem(q.key, 0, str).Err()
				if err != nil {
					return err
				}
			}
			return nil
		})
		return err
	}, q.key)
	if err == redis.TxFailedErr {
		logger.Warne(err)
		return q.DequeueN(n)
	}
	if err != nil {
		return nil, -1, err
	}

	return taken, int(size) - len(taken), nil
}

func (q *RedisQueue) Enqueue(datum []byte) (pos int, err error) {
	var new_len int64
	err = q.client.Watch(func(tx *redis.Tx) error {
		cur_pos, err := q.position(tx, datum)
		if err != nil {
			return err
		}
		if cur_pos >= 0 {
			return queue.AlreadyEnqueued.NewWith(
				fmt.Sprintf("%+v in position %d",
					datum, cur_pos+1),
				queue.SetPosition(pos))
		}

		new_len, err = tx.RPush(q.key, datum).Result()
		return err
	}, q.key)
	if err == redis.TxFailedErr {
		logger.Warne(err)
		return q.Enqueue(datum)
	}
	if err != nil {
		return -1, err
	}

	return int(new_len) - 1, nil
}

func (q *RedisQueue) Iter(fn func(int, []byte) bool) error {
	items, err := q.client.LRange(q.key, 0, -1).Result()
	if err != nil {
		return err
	}

	for i, item := range items {
		if fn(i, []byte(item)) {
			break
		}
	}

	return nil
}

func (q *RedisQueue) Position(datum []byte) (pos int, err error) {
	return q.position(q.client, datum)
}

func (q *RedisQueue) position(cmd redis.Cmdable, datum []byte) (
	pos int, err error) {

	items, err := cmd.LRange(q.key, 0, -1).Result()
	if err != nil {
		return -1, err
	}

	for i, item := range items {
		if bytes.Equal(datum, []byte(item)) {
			return i, nil
		}
	}

	return -1, nil
}

func (q *RedisQueue) Remove(datum []byte) error {
	return q.client.LRem(q.key, 0, datum).Err()
}

func (q *RedisQueue) Size() (int, error) {
	s, err := q.client.LLen(q.key).Result()
	return int(s), err
}
