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
	"github.com/ewollesen/discordgo"
	"github.com/ewollesen/zenbot/commands"
)

type Message interface {
	ChannelId() string
}

type Session interface {
	Channel(channel_id string) (*discordgo.Channel, error)
	ChannelMessageSend(channel_id, message string) error
	Member(guild_id, user_id string) (*discordgo.Member, error)
	User(user_id string) (*discordgo.User, error)
	UserChannelCreate(user_id string) (*discordgo.Channel, error)
	UserChannelPermissions(user_id, channel_id string) (int, error)
}

type DiscordHandler interface {
	Handle(s Session, m *discordgo.MessageCreate, argv ...string) error
	commands.CommandWithHelp
}

type discordHandler struct {
	commands.CommandHandler
}
