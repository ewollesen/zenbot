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
	"bytes"
	"fmt"
	"regexp"
	"testing"

	redis "gopkg.in/redis.v5"

	"github.com/ewollesen/zenbot/zentest"
)

var keyPrefix = "test.zenbot.cache"

func TestFetch(t *testing.T) {
	test := zentest.New(t)
	rc := New(redisTestClient(t), keyPrefix, 0)

	value, err := rc.Fetch("foo", func() []byte {
		return []byte("bar")
	})
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))

	value, err = rc.Fetch("foo", func() []byte {
		return []byte("bar2")
	})
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))
}

func TestGet(t *testing.T) {
	test := zentest.New(t)
	rc := New(redisTestClient(t), keyPrefix, 0)

	value, err := rc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte(nil)))

	test.AssertNil(rc.Set("foo", []byte("bar")))
	value, err = rc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))
}

func TestSet(t *testing.T) {
	test := zentest.New(t)
	rc := New(redisTestClient(t), keyPrefix, 0)

	test.AssertNil(rc.Set("foo", []byte("bar")))
	value, err := rc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))

	test.AssertNil(rc.Set("foo", []byte("bar2")))
	value, err = rc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar2")))
}

func redisTestClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if client == nil {
		t.SkipNow()
	}

	var cursor uint64
	var n int
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(cursor, "", 10).Result()
		if err != nil {
			t.SkipNow()
		}
		n += len(keys)
		if cursor == 0 {
			break
		}

		for _, key := range keys {
			match, err := regexp.MatchString(fmt.Sprintf("^%s", keyPrefix), key)
			if err != nil {
				t.SkipNow()
			}
			if match {
				err := client.Del(key).Err()
				if err != nil {
					t.SkipNow()
				}
			}
		}
	}

	return client
}
