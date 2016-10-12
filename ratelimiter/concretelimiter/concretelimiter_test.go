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

package concretelimiter

import (
	"testing"
	"time"

	"github.com/ewollesen/zenbot/ratelimiter"
	"github.com/ewollesen/zenbot/zentest"
)

func TestClear(t *testing.T) {
	test, lim := newLimTest(t)
	test.setLimit("foo")
	test.AssertNil(lim.Clear())
	_, err := lim.Limit("foo")
	test.AssertNil(err)
}

func TestLimit(t *testing.T) {
	test, lim := newLimTest(t)

	f, err := lim.Limit("foo")
	test.AssertNil(err)

	f()
	f, err = lim.Limit("foo")
	test.AssertErrorContainedBy(err, ratelimiter.TooSoon)
}

type limTest struct {
	*zentest.ZenTest
	lim *rateLimiter
}

func newLimTest(t *testing.T) (*limTest, *rateLimiter) {
	lim := New(time.Second)
	return &limTest{
		lim:     lim,
		ZenTest: zentest.New(t),
	}, lim
}

func (t *limTest) setLimit(id string) {
	f, err := t.lim.Limit(id)
	t.AssertNil(err)
	f()
	_, err = t.lim.Limit(id)
	t.AssertErrorContainedBy(err, ratelimiter.TooSoon)
}
