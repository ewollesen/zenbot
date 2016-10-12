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

	"github.com/ewollesen/zenbot/cache"

	"github.com/ewollesen/discordgo"
)

type cachingSession struct {
	cache cache.Cache
	*session
}

var _ Session = (*cachingSession)(nil)

func newCachingSession(session *discordgo.Session, cache cache.Cache) *cachingSession {
	return &cachingSession{
		cache:   cache,
		session: newSession(session),
	}
}

func (s *cachingSession) Channel(channel_id string) (
	ch *discordgo.Channel, err error) {

	cache_hit := true
	ch_bytes, err := s.cache.Fetch("channels-"+channel_id, func() []byte {
		cache_hit = false
		ch, err := s.session.Channel(channel_id)
		if err != nil {
			logger.Warne(err)
			return nil
		}
		ch_bytes, err := json.Marshal(ch)
		if err != nil {
			logger.Warne(err)
			return nil
		}
		return ch_bytes
	})
	if err != nil {
		return nil, err
	}

	ch = &discordgo.Channel{}
	err = json.Unmarshal(ch_bytes, ch)
	if err != nil {
		logger.Warne(err)
		return nil, err
	}

	logger.Debugf("cache_hit for channel %q: %t", channel_id, cache_hit)

	return ch, nil
}
