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
	"testing"

	"github.com/ewollesen/zenbot/zentest"
)

func TestFetch(t *testing.T) {
	test := zentest.New(t)
	mc := New()

	value, err := mc.Fetch("foo", func() []byte {
		return []byte("bar")
	})
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))

	value, err = mc.Fetch("foo", func() []byte {
		return []byte("bar2")
	})
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))
}

func TestGet(t *testing.T) {
	test := zentest.New(t)
	mc := New()

	value, err := mc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte(nil)))

	test.AssertNil(mc.Set("foo", []byte("bar")))
	value, err = mc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))
}

func TestSet(t *testing.T) {
	test := zentest.New(t)
	mc := New()

	test.AssertNil(mc.Set("foo", []byte("bar")))
	value, err := mc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar")))

	test.AssertNil(mc.Set("foo", []byte("bar2")))
	value, err = mc.Get("foo")
	test.AssertNil(err)
	test.Assert(bytes.Equal(value, []byte("bar2")))
}
