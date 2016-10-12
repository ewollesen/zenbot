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
	"strings"

	"github.com/ewollesen/discordgo"
	"github.com/spacemonkeygo/errors"
)

var (
	CommandNotFound = Error.NewClass("command not found",
		errors.NoCaptureStack())
)

var _ DiscordHandler = (*discordHandler)(nil)

func (h *discordHandler) Handle(s Session, m *discordgo.MessageCreate,
	argv ...string) (err error) {

	msg, err := h.CommandHandler.Handle(argv...)
	if err != nil {
		return err
	}

	return s.ChannelMessageSend(m.ChannelID, msg)
}

func (b *bot) RegisterCommand(name string, handler DiscordHandler) error {
	b.command_handlers_mu.Lock()
	b.command_handlers[name] = handler
	b.command_handlers_mu.Unlock()

	return nil
}

func (b *bot) handleCommand(s Session, m *discordgo.MessageCreate) error {
	argv := []string{}
	if isPrivateMessage(s, m) {
		argv = strings.Split(m.Content, " ")
	} else {
		argv = strings.Split(m.Content[len(*commandPrefix):], " ")
	}
	cmd := strings.ToLower(argv[0])

	b.command_handlers_mu.Lock()
	handler, ok := b.command_handlers[cmd]
	b.command_handlers_mu.Unlock()

	if !ok {
		return CommandNotFound.New(cmd)
	}

	return handler.Handle(s, m, argv...)
}
