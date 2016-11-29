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

package owapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ewollesen/zenbot/blizzard"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/spacemonkeygo/spacelog"
)

var _ overwatch.OverwatchAPI = (*owApi)(nil)

var (
	logger = spacelog.GetLogger()
)

type owApi struct {
	official overwatch.OfficialAPI
}

func New(official overwatch.OfficialAPI) *owApi {
	return &owApi{official: official}
}

type stats struct {
	US *regionData `json:"us"`
	EU *regionData `json:"eu"`
	KR *regionData `json:"kr"`
}

type regionData struct {
	Stats struct {
		Competitive *struct {
			OverallStats *struct {
				CompRank *int `json:"comprank"`
			} `json:"overall_stats"`
		} `json:"competitive"`
	} `json:"stats"`
}

func (l *owApi) SkillRank(platform, battle_tag string) (
	sr int, img string, err error) {

	if !blizzard.WellFormedBattleTag(battle_tag) {
		return -1, "", overwatch.BattleTagInvalid.New(battle_tag)
	}

	json_bytes, err := l.get("stats", platform, battle_tag)
	if err != nil {
		return -1, "", err
	}
	logger.Debugf("raw json: %s", string(json_bytes))

	stats := &stats{}
	err = json.Unmarshal(json_bytes, stats)
	if err != nil {
		return -1, "", err
	}

	for _, region := range overwatch.Regions {
		sr = findRank(stats, region)
		if sr > 0 {
			logger.Infof("found %s's SR in region %s", battle_tag, region)
			return sr, "", nil
		}
	}

	return -1, "", overwatch.BattleTagNotFound.New(battle_tag)
}

func findRank(stats *stats, region string) int {
	var rd *regionData
	switch region {
	case overwatch.RegionEU:
		rd = stats.EU
	case overwatch.RegionKR:
		rd = stats.KR
	case overwatch.RegionUS:
		fallthrough
	default:
		rd = stats.US
	}
	if rd.Stats.Competitive == nil ||
		rd.Stats.Competitive.OverallStats == nil ||
		rd.Stats.Competitive.OverallStats.CompRank == nil {
		return -1
	}
	return *rd.Stats.Competitive.OverallStats.CompRank
}

func (l *owApi) get(path string, platform, battle_tag string) (
	result []byte, err error) {

	resp, err := http.Get(l.buildUrl(platform, battle_tag, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (l *owApi) escapeBattleTag(in string) (out string) {
	return strings.Replace(in, "#", "-", -1)
}

func (l *owApi) buildUrl(platform, battle_tag, path string) string {
	overwatch.CheckPlatform(platform)

	return fmt.Sprintf("%s/%s/%s?platform=%s", "https://owapi.net/api/v3/u",
		l.escapeBattleTag(battle_tag), path, platform)
}

func (l *owApi) IsValidBattleTag(platform, region, battle_tag string) (
	bool, error) {

	return l.official.IsValidBattleTag(platform, region, battle_tag)
}
