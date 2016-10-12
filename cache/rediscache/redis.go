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

package rediscache

import (
	"fmt"
	"regexp"

	"github.com/ewollesen/zenbot/cache"
	"github.com/spacemonkeygo/spacelog"
	redis "gopkg.in/redis.v4"
)

var (
	logger = spacelog.GetLogger()
)

type rediscache struct {
	client *redis.Client
	key    string
}

var _ cache.Cache = (*rediscache)(nil)

func New(client *redis.Client, key string) *rediscache {
	return &rediscache{
		client: client,
		key:    key,
	}
}

func (c *rediscache) Clear() (err error) {
	c.Iter(func(key string, value []byte) bool {
		err = c.client.Del(key).Err()
		if err != nil {
			return true
		}
		return false
	})

	return err
}

func (c *rediscache) Fetch(key string, fn func() []byte) (
	val []byte, err error) {

	value, err := c.client.Get(c.buildKey(key)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	val = []byte(value)
	if value == "" {
		val = fn()
		logger.Warne(c.Set(key, val))
	}

	return val, nil
}

func (c *rediscache) Get(key string) ([]byte, error) {
	return c.get(c.buildKey(key))
}

func (c *rediscache) Iter(fn func(key string, value []byte) bool) {
	var cursor uint64
	var n int

outer:
	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(cursor, "", 10).Result()
		if err != nil {
			logger.Warne(err)
			continue
		}

		n += len(keys)
		if cursor == 0 {
			break
		}

		for _, k := range keys {
			match, err := regexp.MatchString(fmt.Sprintf("^%s", c.key), k)
			if err != nil {
				logger.Warne(err)
				continue
			}
			if match {
				value, err := c.get(k)
				if err != nil {
					logger.Warne(err)
					continue
				}
				if fn(k, value) {
					break outer
				}
			}
		}
	}
}

func (c *rediscache) get(built_key string) (value []byte, err error) {
	str_value, err := c.client.Get(built_key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return []byte(str_value), nil
}

func (c *rediscache) Set(key string, value []byte) error {
	return c.client.Set(c.buildKey(key), value, 0).Err()
}

func (c *rediscache) buildKey(key string) string {
	return c.key + "-" + key
}
