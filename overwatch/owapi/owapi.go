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
	owApiRegions = []string{overwatch.RegionUS, overwatch.RegionEU, overwatch.RegionKR}

	logger = spacelog.GetLogger()
)

type owApi struct {
	host     string
	official overwatch.OfficialAPI
}

func New(official overwatch.OfficialAPI, host string) *owApi {
	return &owApi{host: host, official: official}
}

type stats struct {
	US     *regionData `json:"us,omitempty"`
	EU     *regionData `json:"eu,omitempty"`
	KR     *regionData `json:"kr,omitempty"`
	CN     *regionData `json:"cn,omitempty"`
	Global *regionData `json:"global,omitempty"`
}

type regionData struct {
	Stats struct {
		Competitive *struct {
			OverallStats *struct {
				CompRank *int `json:"comprank,omitempty"`
			} `json:"overall_stats,omitempty"`
		} `json:"competitive,omitempty"`
	} `json:"stats"`
}

func (l *owApi) SkillRank(platform, battle_tag string) (
	sr int, err error) {

	if !blizzard.WellFormedBattleTag(battle_tag) {
		return overwatch.SkillRankError, overwatch.BattleTagInvalid.New(battle_tag)
	}

	json_bytes, err := l.get("stats", platform, battle_tag)
	if err != nil {
		return overwatch.SkillRankError, err
	}
	logger.Debugf("raw json: %s", string(json_bytes))

	stats := &stats{}
	err = json.Unmarshal(json_bytes, stats)
	if err != nil {
		return overwatch.SkillRankError, err
	}

	for _, region := range owApiRegions {
		sr = findRank(stats, region)
		if sr > 0 {
			logger.Infof("found %s's SR in region %s", battle_tag, region)
			return sr, nil
		}
	}

	return overwatch.SkillRankError, overwatch.BattleTagUnranked.New(battle_tag)
}

func findRank(stats *stats, region string) int {
	var rd *regionData
	switch region {
	case overwatch.RegionEU:
		rd = stats.EU
	case overwatch.RegionKR:
		rd = stats.KR
	case overwatch.RegionCN:
		rd = stats.CN
	case overwatch.RegionGlobal:
		rd = stats.Global
	default:
		fallthrough
	case overwatch.RegionUS:
		rd = stats.US
	}

	if rd == nil {
		logger.Debugf("no data in %s region", region)
		return overwatch.SkillRankError
	}

	if rd.Stats.Competitive == nil ||
		rd.Stats.Competitive.OverallStats == nil ||
		rd.Stats.Competitive.OverallStats.CompRank == nil {
		logger.Debugf("unranked in %s", region)
		return overwatch.SkillRankError
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, overwatch.BattleTagNotFound.New(battle_tag)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("received status code %d, continuing anyway",
			resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (l *owApi) escapeBattleTag(in string) (out string) {
	return strings.Replace(in, "#", "-", -1)
}

func (l *owApi) buildUrl(platform, battle_tag, path string) string {
	overwatch.CheckPlatform(platform)

	return fmt.Sprintf("%s/api/v3/u/%s/%s?platform=%s", l.host,
		l.escapeBattleTag(battle_tag), path, platform)
}

func (l *owApi) IsValidBattleTag(platform, region, battle_tag string) (
	bool, error) {

	return l.official.IsValidBattleTag(platform, region, battle_tag)
}
