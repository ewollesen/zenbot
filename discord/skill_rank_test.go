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

package discord

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/ewollesen/zenbot/overwatch/mockoverwatch"
)

func TestHandleTeams(t *testing.T) {
	test := newDiscordTest(t)

	cmdv := append([]string{"teams"}, mockoverwatch.TestBattleTags...)
	srh := newSkillRankHandler(mockoverwatch.NewRandom())
	s := test.mockSession()
	m := test.testMessage("!" + strings.Join(cmdv, " "))
	rand.Seed(13)

	test.AssertNil(srh.Handle(s, m, cmdv...))
	test.AssertContainsRe(s.sends, "Team 1 \\(avg\\. 2627\\.5\\): testuser10#1010  testuser3#3333  testuser4#4444  testuser5#5555  testuser7#7777  testuser9#9999")
	test.AssertContainsRe(s.sends, "Team 2 \\(avg\\. 2627\\.5\\): testuser1#1111  testuser11#1111  testuser12#1212  testuser2#2222  testuser6#6666  testuser8#8888")

}
