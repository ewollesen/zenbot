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

package mocklimiter

import "github.com/ewollesen/zenbot/ratelimiter"

type mockRateLimiter struct {
	Clears int
	Limits int
}

var _ ratelimiter.RateLimiter = (*mockRateLimiter)(nil)

func New() *mockRateLimiter {
	return &mockRateLimiter{}
}

func (l *mockRateLimiter) Clear() error {
	l.Clears++
	return nil
}

func (l *mockRateLimiter) Limit(id string) (func() error, error) {
	return func() error {
		l.Limits++
		return nil
	}, nil
}
