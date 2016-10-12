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
)

type debugHandler struct {
	btags *BattleTagCache
	q     *BattleTagQueue
}

var _ DiscordHandler = (*discordHandler)(nil)

func newDebugHandler(q *BattleTagQueue, b *BattleTagCache) *debugHandler {
	return &debugHandler{
		q:     q,
		btags: b,
	}
}

func (h *debugHandler) Handle(s Session, m *discordgo.MessageCreate,
	argv ...string) (err error) {

	cmd := argv[0]
	switch cmd {
	case "debug":
		sub_cmd := "help"
		if len(argv) > 1 {
			sub_cmd = strings.ToLower(argv[1])
		}
		switch sub_cmd {
		case "cache":
			err = h.handleCache(s, m)
		case "clear":
			err = h.handleClear(s, m)
		default:
			err = reply(s, m, "no help written yet")
		}
	}

	return err
}

func (h *debugHandler) Help(argv ...string) string {
	return "no help written yet"
}

func (h *debugHandler) handleCache(s Session, m *discordgo.MessageCreate) (
	err error) {

	logger.Warne(reply(s, m, "Cache dump starting:"))
	h.btags.Iter(func(key string, value string) bool {
		logger.Warne(reply(s, m, "key: %s => %s", key, value))
		return false
	})
	logger.Warne(reply(s, m, "Cache dump completed."))
	return nil
}

func (h *debugHandler) handleClear(s Session, m *discordgo.MessageCreate) (
	err error) {

	err = h.btags.Clear()
	if err != nil {
		return reply(s, m, "Error clearing cache: %#v", err)
	}
	return reply(s, m, "Cache cleared.")
}
