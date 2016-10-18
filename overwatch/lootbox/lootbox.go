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
	"github.com/spacemonkeygo/spacelog"
)

var _ overwatch.OverwatchAPI = (*lootBox)(nil)

var (
	logger = spacelog.GetLogger()
)

type lootBox struct{}

func New() *lootBox {
	return &lootBox{}
}

type profile struct {
	Data struct {
		Competitive struct {
			Rank    string `json:"rank"`
			RankImg string `json:"rank_img"`
		} `json:"competitive"`
	} `json:"data"`
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

func (l *lootBox) SkillRank(platform, region, battle_tag string) (
	sr int, img string, err error) {

	json_bytes, err := l.get("profile", platform, region, battle_tag)
	if err != nil {
		return -1, "", err
	}
	logger.Debugf("raw json: %s", string(json_bytes))

	profile := &profile{}
	err = json.Unmarshal(json_bytes, profile)
	if err != nil {
		return -1, "", err
	}

	sr64, err := strconv.ParseInt(profile.Data.Competitive.Rank, 10, 32)
	if err != nil {
		return -1, "", err
	}

	return int(sr64), profile.Data.Competitive.RankImg, nil
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

	return fmt.Sprintf("https://api.lootbox.eu/%s/%s/%s/%s",
		platform, region, l.escapeBattleTag(battle_tag), path)
}
