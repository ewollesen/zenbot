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
	"strings"

	"net/http"

	"github.com/ewollesen/zenbot/blizzard"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/spacemonkeygo/spacelog"
)

var _ overwatch.OverwatchAPI = (*blizzardScrape)(nil)

var (
	logger = spacelog.GetLogger()
)

type blizzardScrape struct {
	client *http.Client
}

func New() *blizzardScrape {
	return &blizzardScrape{
		client: &http.Client{},
	}
}

func (b *blizzardScrape) SkillRank(platform, battle_tag string) (
	sr int, err error) {
	return overwatch.SkillRankError, fmt.Errorf("not implemented")
}

func (b *blizzardScrape) IsValidBattleTag(platform, region, battle_tag string) (
	bool, error) {

	if !blizzard.WellFormedBattleTag(battle_tag) {
		return false, nil
	}

	resp, err := b.client.Head(b.buildUrl(platform, region, battle_tag))
	if err != nil {
		return false, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return false, nil
	case http.StatusOK:
		return true, nil
	default:
		return false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

func (b *blizzardScrape) escapeBattleTag(in string) (out string) {
	return strings.Replace(in, "#", "-", -1)
}

func (b *blizzardScrape) buildUrl(platform, region, battle_tag string) string {
	overwatch.CheckPlatform(platform)
	overwatch.CheckRegion(region)
	return fmt.Sprintf("https://playoverwatch.com/en-us/career/%s/%s/%s",
		platform, region, b.escapeBattleTag(battle_tag))
}
