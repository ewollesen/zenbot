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

import "github.com/spacemonkeygo/spacelog"

var (
	logger = spacelog.GetLogger()
)

type OverwatchAPI interface {
	SkillRank(platform, region, battle_tag string) (
		sr int, img_url string, err error)
}

func CheckPlatform(platform string) {
	switch platform {
	case "pc", "psn", "xbl":
		// no op
	default:
		logger.Noticef("continuing with unexpected platform: %q", platform)
	}
}

func CheckRegion(region string) {
	switch region {
	case "us", "eu", "kr", "cn", "global":
		// no op
	default:
		logger.Noticef("continuing with unexpected region: %q", region)
	}
}

// TOOD: CheckBattleTag
