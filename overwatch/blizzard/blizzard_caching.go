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

package blizzard

import (
	"fmt"
	"strconv"

	"github.com/ewollesen/zenbot/cache"
)

type cachingBlizzardScrape struct {
	*blizzardScrape
	cache cache.Cache
}

func NewCaching(cache cache.Cache) *cachingBlizzardScrape {
	return &cachingBlizzardScrape{
		cache:          cache,
		blizzardScrape: New(),
	}
}

func (b *cachingBlizzardScrape) IsValidBattleTag(
	platform, region, battle_tag string) (bool, error) {

	value, err := b.cache.Fetch(b.key(platform, region, battle_tag), func() []byte {
		valid, err := b.blizzardScrape.IsValidBattleTag(platform, region, battle_tag)
		if err != nil {
			logger.Warne(err)
			return nil
		}
		if !valid {
			return nil
		}
		return []byte(fmt.Sprintf("%t", valid))
	})
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(string(value))
}

func (b *cachingBlizzardScrape) key(platform, region, btag string) string {
	return fmt.Sprintf("%s/%s/%s", platform, region, btag)
}
