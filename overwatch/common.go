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
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"
)

var (
	logger = spacelog.GetLogger()

	Error            = errors.NewClass("overwatch")
	BattleTagInvalid = Error.NewClass("BattleTag invalid",
		errors.NoCaptureStack())
	BattleTagNotFound = Error.NewClass("no BattleTag found",
		errors.NoCaptureStack())
	BattleTagUnranked = Error.NewClass("no skill ranking found",
		errors.NoCaptureStack())
)

const (
	RegionUS     = "us"
	RegionEU     = "eu"
	RegionKR     = "kr"
	RegionCN     = "cn"
	RegionGlobal = "global"

	PlatformPC  = "pc"
	PlatformPSN = "psn"
	PlatformXBL = "xbl"

	SkillRankError = -1
)

var (
	Regions   = []string{RegionUS, RegionEU, RegionKR, RegionCN, RegionGlobal}
	Platforms = []string{PlatformPC, PlatformPSN, PlatformXBL}
)

type OfficialAPI interface {
	IsValidBattleTag(platform, region, battle_tag string) (
		bool, error)
}

type OverwatchAPI interface {
	SkillRank(platform, battle_tag string) (sr int, err error)
	OfficialAPI
}

type RegionalOverwatchAPI interface {
	SkillRank(platform, region, battle_tag string) (sr int, err error)
	OfficialAPI
}

func CheckPlatform(platform string) {
	switch platform {
	case PlatformPC, PlatformPSN, PlatformXBL:
		// no op
	default:
		logger.Noticef("continuing with unexpected platform: %q",
			platform)
	}
}

func CheckRegion(region string) {
	switch region {
	case RegionUS, RegionEU, RegionKR, RegionCN, RegionGlobal:
		// no op
	default:
		logger.Noticef("continuing with unexpected region: %q", region)
	}
}

func RankToDivision(rank int) string {
	switch {
	case rank < 1500:
		return "bronze"
	case rank < 2000:
		return "silver"
	case rank < 2500:
		return "gold"
	case rank < 3000:
		return "platinum"
	case rank < 3500:
		return "diamond"
	case rank < 4000:
		return "master!"
	case rank >= 4000:
		return "grandmaster!!!"
	default:
		return "unknown"
	}
}

// TOOD: CheckBattleTag
