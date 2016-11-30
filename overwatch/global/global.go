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

package global

import (
	"github.com/ewollesen/zenbot/blizzard"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/spacemonkeygo/spacelog"
)

var (
	logger = spacelog.GetLogger()
)

type GlobalOverwatch struct {
	overwatch.RegionalOverwatchAPI
}

func (o *GlobalOverwatch) SkillRank(platform, battle_tag string) (
	sr int, err error) {

	if !blizzard.WellFormedBattleTag(battle_tag) {
		return overwatch.SkillRankError,
			overwatch.BattleTagInvalid.New(battle_tag)
	}

	// TODO parallelize
	for _, region := range overwatch.Regions {
		sr, err = o.RegionalOverwatchAPI.SkillRank(platform, region,
			battle_tag)
		if err != nil {
			logger.Infoe(err)
			continue
		}
		return sr, err
	}

	return overwatch.SkillRankError, err
}

func New(regional overwatch.RegionalOverwatchAPI) *GlobalOverwatch {
	return &GlobalOverwatch{regional}
}
