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
	Rank     int    `json:"rank"`
	ImageUrl string `json:"image_url"`
}

func NewCaching(overwatch OverwatchAPI, cache cache.Cache) *cachingOverwatch {
	return &cachingOverwatch{
		OverwatchAPI: overwatch,
		cache:        cache,
	}
}

func (c *cachingOverwatch) SkillRank(platform, battle_tag string) (
	sr int, img_url string, err error) {

	cache_hit := true
	var inner_err error
	val_bytes, err := c.cache.Fetch(
		c.key(platform, battle_tag, "skill_rank_w_image"),
		func() []byte {
			cache_hit = false
			logger.Debugf("skill rank cache miss for %q", battle_tag)
			r, img_url, err := c.OverwatchAPI.SkillRank(
				platform, battle_tag)
			if err != nil {
				logger.Errore(err)
				if BattleTagUnranked.Contains(err) {
					inner_err = err
				}
				return nil
			}
			r_bytes, err := json.Marshal(&skillRankBlob{
				Rank: r, ImageUrl: img_url})
			if err != nil {
				logger.Errore(err)
				return nil
			}

			return r_bytes
		})
	logger.Debugf("inner_err: %+v", inner_err)
	if inner_err != nil && BattleTagUnranked.Contains(inner_err) {
		return -1, "", inner_err
	}
	if err != nil {
		return -1, "", err
	}
	if cache_hit {
		logger.Debugf("skill rank cache hit for %q", battle_tag)
	}

	blob := &skillRankBlob{}
	err = json.Unmarshal(val_bytes, blob)
	if err != nil {
		return -1, "", err
	}

	return blob.Rank, blob.ImageUrl, nil
}

func (c *cachingOverwatch) key(pieces ...string) string {
	return strings.Join(pieces, "-")
}
