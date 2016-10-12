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
	"fmt"
	"sync"

	"github.com/ewollesen/zenbot/cache"
)

type memcache struct {
	mu sync.Mutex
	m  map[string][]byte
}

var _ cache.Cache = (*memcache)(nil)

func New() *memcache {
	return &memcache{
		m: make(map[string][]byte),
	}
}

func (c *memcache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[string][]byte)
	return nil
}

func (c *memcache) Fetch(key string, fn func() []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.m[key]
	if !ok {
		value = fn()
		if value == nil {
			fmt.Printf("returning nil\n")
			return nil, nil
		}
		c.m[key] = value
	}

	return value, nil
}

func (c *memcache) Get(key string) (value []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.m[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (c *memcache) Iter(fn func(key string, value []byte) bool) {
	for k, v := range c.m {
		if fn(k, v) {
			break
		}
	}
}

func (c *memcache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = value
	return nil
}
