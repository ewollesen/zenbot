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

	"github.com/ewollesen/discordgo"
)

func reply(s Session, m *discordgo.MessageCreate,
	template string, args ...interface{}) error {

	return s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(template, args...))
}

// TODO: make me a real private message
func replyPrivate(s Session, m *discordgo.MessageCreate,
	template string, args ...interface{}) error {

	dm_channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		logger.Errore(err)
		return nil
	}

	return s.ChannelMessageSend(dm_channel.ID, fmt.Sprintf(template, args...))
}

func authorNick(s Session, m *discordgo.MessageCreate) string {
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		logger.Errore(err)
		return ""
	}

	member, err := s.Member(channel.GuildID, m.Author.ID)
	if err != nil {
		logger.Errore(err)
		return ""
	}

	if member.Nick == "" {
		logger.Warne(err)
		return m.Author.Username
	}

	return member.Nick
}

func userKey(s Session, m *discordgo.MessageCreate) string {
	// We no longer include the guild id here, because pm's don't have a
	// guild id.
	return m.Author.ID
}
