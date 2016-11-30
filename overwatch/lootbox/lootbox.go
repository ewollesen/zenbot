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

package lootbox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/ewollesen/zenbot/overwatch"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"
)

var _ overwatch.RegionalOverwatchAPI = (*lootBox)(nil)

var (
	logger = spacelog.GetLogger()

	Error = errors.NewClass("lootbox")
)

type lootBox struct {
	host     string
	official overwatch.OfficialAPI
}

func New(official overwatch.OfficialAPI, host string) *lootBox {
	return &lootBox{host: host, official: official}
}

type profile struct {
	Data *struct {
		Competitive *struct {
			Rank *string `json:"rank,omitempty"`
		} `json:"competitive,omitempty"`
	} `json:"data,omitempty"`
	StatusCode *int    `json:"statusCode,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// curl -X GET --header 'Accept: application/json' 'https://api.lootbox.eu/pc/us/encoded-1148/profile'
// {
// 	"data": {
// 		"username": "encoded",
// 		"level": 133,
// 		"games": {
// 			"quick": {
// 				"wins": "468"
// 			},
// 			"competitive": {
// 				"wins": "8",
// 				"lost": 8,
// 				"played": "16"
// 			}
// 		},
// 		"playtime": {
// 			"quick": "107 hours",
// 			"competitive": "3 hours"
// 		},
// 		"avatar": "https://blzgdapipro-a.akamaihd.net/game/unlocks/0x02500000000009E7.png",
// 		"competitive": {
// 			"rank": "1892",
// 			"rank_img": "https://blzgdapipro-a.akamaihd.net/game/rank-icons/season-2/rank-2.png"
// 		},
// 		"levelFrame": "https://blzgdapipro-a.akamaihd.net/game/playerlevelrewards/0x0250000000000925_Border.png",
// 		"star": "https://blzgdapipro-a.akamaihd.net/game/playerlevelrewards/0x0250000000000925_Rank.png"
// 	}
// }

// curl -s 'https://api.lootbox.eu/pc/us/encoded-1149/profile' | jq .
// 	{
// 	"statusCode": 404,
// 	"error": "Found no user with the BattleTag: encoded-1149"
// }

func (l *lootBox) SkillRank(platform, region, battle_tag string) (
	sr int, err error) {

	json_bytes, err := l.get("profile", platform, region, battle_tag)
	if err != nil {
		return overwatch.SkillRankError, err
	}
	logger.Debugf("raw json: %s", string(json_bytes))

	profile := &profile{}
	err = json.Unmarshal(json_bytes, profile)
	if err != nil {
		return overwatch.SkillRankError, err
	}

	if profile.Error != nil {
		if profile.StatusCode != nil &&
			*profile.StatusCode == http.StatusNotFound {
			return overwatch.SkillRankError, overwatch.BattleTagNotFound.New(battle_tag)
		}
		return overwatch.SkillRankError, Error.New(*profile.Error)
	}

	if unranked(profile) {
		return overwatch.SkillRankError, overwatch.BattleTagUnranked.New(battle_tag)
	}

	sr64, err := strconv.ParseInt(*profile.Data.Competitive.Rank, 10, 32)
	if err != nil {
		return overwatch.SkillRankError, err
	}

	return int(sr64), nil
}

func (l *lootBox) get(path string, platform, region, battle_tag string) (
	result []byte, err error) {

	resp, err := http.Get(l.buildUrl(platform, region, battle_tag, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (l *lootBox) escapeBattleTag(in string) (out string) {
	return strings.Replace(in, "#", "-", -1)
}

func (l *lootBox) buildUrl(platform, region, battle_tag, path string) string {
	overwatch.CheckPlatform(platform)
	overwatch.CheckRegion(region)

	return fmt.Sprintf("%s/%s/%s/%s/%s", l.host, platform, region,
		l.escapeBattleTag(battle_tag), path)
}

func (l *lootBox) IsValidBattleTag(platform, region, battle_tag string) (
	bool, error) {

	return l.official.IsValidBattleTag(platform, region, battle_tag)
}

func unranked(pr *profile) bool {
	return pr.Data.Competitive == nil ||
		pr.Data.Competitive.Rank == nil ||
		*pr.Data.Competitive.Rank == ""
}
