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
	"testing"

	redis "gopkg.in/redis.v4"

	"github.com/ewollesen/zenbot/queue"
)

var keyPrefix = "test.zenbot.queue"

func TestClear(t *testing.T) {
	queue.CommonTestClear(t, New(redisTestClient(t), keyPrefix))
}

func TestDequeueN(t *testing.T) {
	queue.CommonTestDequeueN(t, New(redisTestClient(t), keyPrefix))
}

func TestEnqueue(t *testing.T) {
	queue.CommonTestEnqueue(t, New(redisTestClient(t), keyPrefix))
}

func TestIter(t *testing.T) {
	queue.CommonTestIter(t, New(redisTestClient(t), keyPrefix))
}

func TestPosition(t *testing.T) {
	queue.CommonTestPosition(t, New(redisTestClient(t), keyPrefix))
}

func TestRemove(t *testing.T) {
	queue.CommonTestRemove(t, New(redisTestClient(t), keyPrefix))
	queue.CommonTestRemoveFirst(t, New(redisTestClient(t), keyPrefix))
	queue.CommonTestRemoveMiddle(t, New(redisTestClient(t), keyPrefix))
	queue.CommonTestRemoveLast(t, New(redisTestClient(t), keyPrefix))
}

func TestSize(t *testing.T) {
	queue.CommonTestSize(t, New(redisTestClient(t), keyPrefix))
}

func redisTestClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if client == nil {
		t.SkipNow()
	}

	err := client.Del(keyPrefix).Err()
	if err != nil {
		t.SkipNow()
	}

	return client
}
