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
	"sync"
	"time"

	"github.com/spacemonkeygo/spacelog"

	"github.com/ewollesen/zenbot/ratelimiter"
)

var (
	logger = spacelog.GetLogger()
)

type rateLimiter struct {
	mu         sync.Mutex
	timestamps map[string]time.Time
	dur        time.Duration
}

var _ ratelimiter.RateLimiter = (*rateLimiter)(nil)

func New(dur time.Duration) *rateLimiter {
	return &rateLimiter{
		dur:        dur,
		timestamps: make(map[string]time.Time),
	}
}

func (l *rateLimiter) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamps = make(map[string]time.Time)
	return nil
}

func (l *rateLimiter) Limit(id string) (func() error, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	at, ok := l.timestamps[id]
	if ok && now.Before(at.Add(l.dur)) {
		return nil, ratelimiter.TooSoon.New("")
	}

	return l.trigger(id, now), nil
}

func (l *rateLimiter) trigger(id string, now time.Time) func() error {
	return func() error {
		l.mu.Lock()
		defer l.mu.Unlock()

		l.timestamps[id] = now

		return nil
	}
}
