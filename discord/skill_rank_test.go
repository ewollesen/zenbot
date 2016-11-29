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
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/ewollesen/discordgo"
	memorycache "github.com/ewollesen/zenbot/cache/memory"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/overwatch/mockoverwatch"
)

func TestHandleTeams(t *testing.T) {
	test := newDiscordTest(t)

	cmdv := append([]string{"teams"}, mockoverwatch.TestBattleTags...)
	btc := NewBattleTagCache(memorycache.New())
	srh := newSkillRankHandler(btc, overwatch.NewGlobal(mockoverwatch.NewRandom()))
	s := test.mockSession()
	for i, _ := range mockoverwatch.TestBattleTags {
		test_user_id := fmt.Sprintf("test-user-%03d", 123+i)
		s.setMember(testGuildId, test_user_id, &discordgo.Member{
			Nick: test_user_id,
		})
	}
	m := test.testMessage("!" + strings.Join(cmdv, " "))
	rand.Seed(13)
	test.AssertNil(srh.Handle(s, m, cmdv...))
	test.AssertContainsRe(s.sends, "Team \\d \\(avg\\. 2627\\.5\\): testuser10#1010  testuser3#3333  testuser4#4444  testuser5#5555  testuser7#7777  testuser9#9999")
	test.AssertContainsRe(s.sends, "Team \\d \\(avg\\. 2627\\.5\\): testuser1#1111  testuser11#1111  testuser12#1212  testuser2#2222  testuser6#6666  testuser8#8888")
}

func TestHandleTeamsReplaceMentions(t *testing.T) {
	test := newDiscordTest(t)

	btc := NewBattleTagCache(memorycache.New())
	srh := newSkillRankHandler(btc, overwatch.NewGlobal(mockoverwatch.NewRandom()))
	s := test.mockSession()
	for i, _ := range mockoverwatch.TestBattleTags {
		test_user_id := fmt.Sprintf("test-user-%03d", 123+i)
		s.setMember(testGuildId, test_user_id, &discordgo.Member{
			Nick: test_user_id,
		})
	}

	btc.Set("1234", "foobar#4321")
	btc.Set("5678", "bazquux#5678")
	m := test.testMessage("!teams <@1234> <@!5678> <@!9012> <@3456> example#1234")
	srh.handleTeams(s, m)
	test.AssertContainsRe(s.sends, `Team \d \(avg\. 2512\.0\): bazquux#5678  example#1234`)
	test.AssertContainsRe(s.sends, `Team \d \(avg\. 2856\.0\): foobar#4321`)
}

func TestHandleReplaceMentions(t *testing.T) {
	test := newDiscordTest(t)

	btc := NewBattleTagCache(memorycache.New())
	srh := newSkillRankHandler(btc, overwatch.NewGlobal(mockoverwatch.NewRandom()))
	btc.Set("1234", "example#1234")

	test.AssertEqual(srh.replaceMentions(
		"this is a <@1234> of the <@5678> with extra foobar#1111"),
		"this is a example#1234 of the <@5678> with extra foobar#1111")

	test.Assert(true)
}
