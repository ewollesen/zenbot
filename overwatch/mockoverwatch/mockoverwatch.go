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
)

var (
	TestBattleTags = []string{
		"testuser1#1111",
		"testuser2#2222",
		"testuser3#3333",
		"testuser4#4444",
		"testuser5#5555",
		"testuser6#6666",
		"testuser7#7777",
		"testuser8#8888",
		"testuser9#9999",
		"testuser10#1010",
		"testuser11#1111",
		"testuser12#1212",
	}
	testBattleTagsTeamOne = []string{
		"testuser9#9999",
		"testuser8#8888",
		"testuser7#7777",
		"testuser6#6666",
		"testuser4#4444",
		"testuser11#1111",
	}
	testBattleTagsTeamTwo = []string{
		"testuser5#5555",
		"testuser3#3333",
		"testuser2#2222",
		"testuser12#1212",
		"testuser1#1111",
		"testuser10#1010",
	}
)

type mockOverwatch struct{}

func (ow *mockOverwatch) SkillRank(platform, region, btag string) (
	rank int, img_url string, err error) {

	switch btag {
	case "testuser1#1111":
		return 2000, "", nil
	case "testuser2#2222":
		return 2056, "", nil
	case "testuser3#3333":
		return 3056, "", nil
	case "testuser4#4444":
		return 4056, "", nil
	case "testuser5#5555":
		return 3656, "", nil
	case "testuser6#6666":
		return 2468, "", nil
	case "testuser7#7777":
		return 2562, "", nil
	case "testuser8#8888":
		return 1265, "", nil
	case "testuser9#9999":
		return 3129, "", nil
	case "testuser10#1010":
		return 2654, "", nil
	case "testuser11#1111":
		return 2296, "", nil
	case "testuser12#1212":
		return 2307, "", nil
	default:
		return -1, "", fmt.Errorf("invalid battle tag")
	}
}

func New() *mockOverwatch {
	return &mockOverwatch{}
}

type mockRandomOverwatch struct {
	ranks_mu sync.Mutex
	ranks    map[string]int
}

func (ow *mockRandomOverwatch) SkillRank(platform, region, btag string) (
	rank int, img_url string, err error) {

	ow.ranks_mu.Lock()
	defer ow.ranks_mu.Unlock()

	rank, ok := ow.ranks[btag]
	if !ok {
		rank = int(rand.NormFloat64()*500 + 2500)
		ow.ranks[btag] = rank
	}
	return rank, "", nil
}

func NewRandom() *mockRandomOverwatch {
	return &mockRandomOverwatch{
		ranks: make(map[string]int),
	}
}
