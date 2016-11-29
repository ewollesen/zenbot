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

package mockoverwatch

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ewollesen/zenbot/overwatch"
)

var (
	TestBattleTags = []string{
		"us/testuser1#1111",
		"us/testuser2#2222",
		"us/testuser3#3333",
		"us/testuser4#4444",
		"us/testuser5#5555",
		"us/testuser6#6666",
		"us/testuser7#7777",
		"us/testuser8#8888",
		"us/testuser9#9999",
		"us/testuser10#1010",
		"us/testuser11#1111",
		"us/testuser12#1212",
	}
	testBattleTagsTeamOne = []string{
		"us/testuser9#9999",
		"us/testuser8#8888",
		"us/testuser7#7777",
		"us/testuser6#6666",
		"us/testuser4#4444",
		"us/testuser11#1111",
	}
	testBattleTagsTeamTwo = []string{
		"us/testuser5#5555",
		"us/testuser3#3333",
		"us/testuser2#2222",
		"us/testuser12#1212",
		"us/testuser1#1111",
		"us/testuser10#1010",
	}
)

type mockOverwatch struct {
	bad_btags []string
}

var _ overwatch.RegionalOverwatchAPI = (*mockOverwatch)(nil)

func (ow *mockOverwatch) SkillRank(platform, region, btag string) (
	rank int, img_url string, err error) {

	switch region + "/" + btag {
	case "us/testuser1#1111":
		return 2000, "", nil
	case "us/testuser2#2222":
		return 2056, "", nil
	case "us/testuser3#3333":
		return 3056, "", nil
	case "us/testuser4#4444":
		return 4056, "", nil
	case "us/testuser5#5555":
		return 3656, "", nil
	case "us/testuser6#6666":
		return 2468, "", nil
	case "us/testuser7#7777":
		return 2562, "", nil
	case "us/testuser8#8888":
		return 1265, "", nil
	case "us/testuser9#9999":
		return 3129, "", nil
	case "us/testuser10#1010":
		return 2654, "", nil
	case "us/testuser11#1111":
		return 2296, "", nil
	case "us/testuser12#1212":
		return 2307, "", nil
	case "eu/testuser13#1313":
		return 3183, "", nil
	case "eu/foundeu#2222":
		return 4998, "", nil
	case "us/foundus#1111":
		return 4999, "", nil
	case "us/foundeu#2222":
		return -1, "", overwatch.BattleTagNotFound.New("")
	case "eu/foundus#1111":
		return -1, "", overwatch.BattleTagNotFound.New("")
	default:
		return -1, "", fmt.Errorf("invalid battle tag")
	}
}

func (ow *mockOverwatch) SkillRankGlobal(platform, btag string) (
	rank int, img_url string, err error) {

	return ow.SkillRank(platform, "us", btag)
}

func (ow *mockOverwatch) IsValidBattleTag(platform, region, btag string) (
	bool, error) {

	for _, candidate := range ow.bad_btags {
		if candidate == btag {
			return false, nil
		}
	}

	return true, nil
}

func (ow *mockOverwatch) SetInvalidBattleTag(btag string) {
	ow.bad_btags = append(ow.bad_btags, btag)
}

func New() *mockOverwatch {
	return &mockOverwatch{}
}

type mockRandomOverwatch struct {
	*mockOverwatch
	ranks_mu sync.Mutex
	ranks    map[string]int
}

func (ow *mockRandomOverwatch) SkillRank(platform, region, btag string) (
	rank int, img_url string, err error) {

	ow.ranks_mu.Lock()
	defer ow.ranks_mu.Unlock()

	rank, ok := ow.ranks[region+"/"+btag]
	if !ok {
		rank = int(rand.NormFloat64()*500 + 2500)
		ow.ranks[region+"/"+btag] = rank
	}
	return rank, "", nil
}

func NewRandom() *mockRandomOverwatch {
	return &mockRandomOverwatch{
		mockOverwatch: New(),
		ranks:         make(map[string]int),
	}
}
