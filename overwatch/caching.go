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

package overwatch

import (
	"encoding/json"
	"strings"

	"github.com/ewollesen/zenbot/cache"
)

var _ OverwatchAPI = (*cachingOverwatch)(nil)

type cachingOverwatch struct {
	OverwatchAPI
	cache cache.Cache
}

type skillRankBlob struct {
	Rank int `json:"rank"`
}

func NewCaching(overwatch OverwatchAPI, cache cache.Cache) *cachingOverwatch {
	return &cachingOverwatch{
		OverwatchAPI: overwatch,
		cache:        cache,
	}
}

func (c *cachingOverwatch) SkillRank(platform, battle_tag string) (
	sr int, err error) {

	cache_hit := true
	val_bytes, err := c.cache.Fetch(
		c.key("skillRank", platform, battle_tag),
		func() []byte {
			cache_hit = false
			logger.Debugf("skill rank cache miss for %q", battle_tag)
			r, err := c.OverwatchAPI.SkillRank(platform, battle_tag)
			if err != nil {
				// Is it desirable to cache unranked battle
				// tags? To reduce traffic if nothing else? If
				// so, this should be re-vamped. Inspecting the
				// error was super ugly due to the caching.
				logger.Errore(err)
				return nil
			}

			r_bytes, err := json.Marshal(&skillRankBlob{Rank: r})
			if err != nil {
				logger.Errore(err)
				return nil
			}

			return r_bytes
		})
	if err != nil {
		return SkillRankError, err
	}
	if cache_hit {
		logger.Debugf("skill rank cache hit for %q", battle_tag)
	}

	blob := &skillRankBlob{}
	err = json.Unmarshal(val_bytes, blob)
	if err != nil {
		return SkillRankError, err
	}

	return blob.Rank, nil
}

func (c *cachingOverwatch) key(pieces ...string) string {
	return strings.Join(pieces, "-")
}
