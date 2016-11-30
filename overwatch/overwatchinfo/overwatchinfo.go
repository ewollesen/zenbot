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

package overwatchinfo

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

var _ overwatch.RegionalOverwatchAPI = (*overwatchInfo)(nil)

var (
	logger = spacelog.GetLogger()

	Error = errors.NewClass("overwatchinfo")
)

type overwatchInfo struct {
	client   *http.Client
	official overwatch.OfficialAPI
}

func New(official overwatch.OfficialAPI) *overwatchInfo {
	return &overwatchInfo{
		client:   &http.Client{},
		official: official,
	}
}

type profile struct {
	Data struct {
		CompetitivePlay struct {
			Rank string `json:"rank"`
		} `json:"competitive_play"`
	} `json:"data"`
	StatusCode int    `json:"statusCode,omitempty"`
	Error      string `json:"error,omitempty"`
}

//curl -X GET --header 'Accept: application/json' 'https://api.overwatchinfo.com/pc/us/PrinceMO-11110/profile'
func (l *overwatchInfo) SkillRank(platform, region, battle_tag string) (
	sr int, err error) {

	json_bytes, err := l.get("profile", platform, region, battle_tag)
	if err != nil {
		return -1, err
	}
	logger.Debugf("raw json: %s", string(json_bytes))

	profile := &profile{}
	err = json.Unmarshal(json_bytes, profile)
	if err != nil {
		return -1, err
	}

	if profile.Error != "" {
		return -1, Error.New(profile.Error)
	}

	sr64, err := strconv.ParseInt(profile.Data.CompetitivePlay.Rank, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(sr64), nil
}

func (l *overwatchInfo) get(path string, platform, region, battle_tag string) (
	result []byte, err error) {

	url := l.buildUrl(platform, region, battle_tag, path)
	logger.Debugf("GET %q", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (l *overwatchInfo) escapeBattleTag(in string) (out string) {
	return strings.Replace(in, "#", "-", -1)
}

func (l *overwatchInfo) buildUrl(platform, region, battle_tag, path string) string {
	overwatch.CheckPlatform(platform)
	overwatch.CheckRegion(region)

	return fmt.Sprintf("https://api.overwatchinfo.com/%s/%s/%s/%s",
		platform, region, l.escapeBattleTag(battle_tag), path)
}

func (l *overwatchInfo) IsValidBattleTag(platform, region, battle_tag string) (
	bool, error) {

	return l.official.IsValidBattleTag(platform, region, battle_tag)
}
