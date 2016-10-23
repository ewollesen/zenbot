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
	"testing"

	"github.com/ewollesen/discordgo"

	"github.com/ewollesen/zenbot/zentest"
)

var (
	testAuthor = &discordgo.User{
		ID: testUserId,
	}
	testBattleTag        = "example#1234"
	testChannelId        = "test-channel-123"
	testGuildId          = "test-guild-123"
	testNick             = "example#1234 [tank]"
	testPrivateChannelId = "test-private-channel-123"
	testUserId           = "test-user-123"
)

type discordTest struct {
	*zentest.ZenTest
}

func newDiscordTest(t *testing.T) *discordTest {
	rand.Seed(13)

	return &discordTest{
		ZenTest: zentest.New(t),
	}
}

type mockSession struct {
	permissions int
	channels    map[string]*discordgo.Channel
	members     map[string]*discordgo.Member
	sends       []string
}

var _ Session = (*mockSession)(nil)

func (s *mockSession) Channel(channel_id string) (*discordgo.Channel, error) {
	return s.channels[channel_id], nil
}

func (s *mockSession) ChannelMessageSend(channel_id, msg string) error {
	s.sends = append(s.sends, msg)
	return nil
}

func (s *mockSession) Member(guild_id, user_id string) (
	*discordgo.Member, error) {
	m, ok := s.members[guild_id+"-"+user_id]
	if ok {
		return m, nil
	}
	return nil, fmt.Errorf("No member set for guild:user %s:%s",
		guild_id, user_id)
}

func (s *mockSession) User(user_id string) (*discordgo.User, error) {
	return nil, nil
}

func (s *mockSession) UserChannelCreate(user_id string) (
	*discordgo.Channel, error) {

	return &discordgo.Channel{
		ID:        testPrivateChannelId,
		IsPrivate: true,
	}, nil
}

func (s *mockSession) UserChannelPermissions(user_id, channel_id string) (
	int, error) {

	return s.permissions, nil
}

func (s *mockSession) grantPermission(perm int) {
	s.permissions = s.permissions | perm
}

func (s *mockSession) setChannel(channel_id string, ch *discordgo.Channel) {
	s.channels[channel_id] = ch
}

func (s *mockSession) setMember(guild_id, user_id string, u *discordgo.Member) {
	s.members[guild_id+"-"+user_id] = u
}

func (s *mockSession) clearSends() {
	s.sends = []string{}
}

func (t *discordTest) mockSession() *mockSession {
	s := &mockSession{
		channels: make(map[string]*discordgo.Channel),
		members:  make(map[string]*discordgo.Member),
		sends:    []string{},
	}
	s.setChannel(testChannelId, &discordgo.Channel{
		GuildID: testGuildId,
	})
	s.setMember(testGuildId, testUserId, &discordgo.Member{
		Nick: testNick,
	})

	return s
}

func (t *discordTest) testMessage(msg string) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author:    testAuthor,
			ChannelID: testChannelId,
			Content:   msg,
		},
	}
}
