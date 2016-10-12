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

import "github.com/ewollesen/discordgo"

type session struct {
	*discordgo.Session
}

func (s *session) Channel(channel_id string) (*discordgo.Channel, error) {
	return s.Session.Channel(channel_id)
}

func (s *session) ChannelMessageSend(channel_id, msg string) error {
	_, err := s.Session.ChannelMessageSend(channel_id, msg)
	return err
}

func (s *session) Member(guild_id, user_id string) (*discordgo.Member, error) {
	return s.Session.State.Member(guild_id, user_id)
}

func (s *session) User(user_id string) (*discordgo.User, error) {
	return s.Session.User(user_id)
}

func (s *session) UserChannelPermissions(user_id, channel_id string) (
	int, error) {

	return s.Session.UserChannelPermissions(user_id, channel_id)
}

func newSession(s *discordgo.Session) *session {
	return &session{
		Session: s,
	}
}
