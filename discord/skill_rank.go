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
	"regexp"
	"strings"

	"github.com/ewollesen/discordgo"
	"github.com/ewollesen/zenbot/overwatch"
)

var (
	rankUrlRe  = regexp.MustCompile(`/(season-[0-9])/(rank-[0-9])\.png$`)
	rankLookup = map[string]map[string]string{
		"season-2": map[string]string{
			"rank-1": "bronze",
			"rank-2": "silver",
			"rank-3": "gold",
			"rank-4": "platinum",
			"rank-5": "diamond",
			"rank-6": "master!",
			"rank-7": "grandmaster!!!",
		},
	}

	skillRankHelpMsg = strings.TrimSpace(strings.Join([]string{
		"Looks up the Skill Rank for a BattleTag. BattleTags are CaSe-SeNsiTiVe! Ranks are cached, and therefore may be slightly out of date.",
		"`!sr example#1234` - looks up the skill rank for example#1234 (PC, US-only for now)",
		"`!sr help` - displays this help message",
	}, "\n"))
)

type skillRankHandler struct {
	overwatch overwatch.OverwatchAPI
}

var _ DiscordHandler = (*skillRankHandler)(nil)

func imageUrlToName(img_url string) string {
	matches := rankUrlRe.FindStringSubmatch(img_url)
	if len(matches) != 3 {
		logger.Debugf("expected 3 matches for rank img lookup, got: %d",
			len(matches))
		return "unknown"
	}

	name := rankLookup[matches[1]][matches[2]]
	if name == "" {
		return "unknown"
	}
	return name
}

func (sr *skillRankHandler) Handle(s Session, m *discordgo.MessageCreate,
	argv ...string) (err error) {

	sub_cmd := "help"
	if len(argv) > 1 {
		sub_cmd = argv[1]
	}
	switch sub_cmd {
	case "help":
		err = reply(s, m, skillRankHelpMsg)
	default:
		err = sr.handleSkillRank(s, m, sub_cmd)
	}

	return err
}

func (sr *skillRankHandler) Help(argv ...string) string {
	return "lookup the skill rank for the given BattleTag (PC, US only for now)"
}

func newSkillRankHandler(ow overwatch.OverwatchAPI) DiscordHandler {
	return &skillRankHandler{
		overwatch: ow,
	}
}

func (sr *skillRankHandler) handleSkillRank(s Session,
	m *discordgo.MessageCreate, btag string) (err error) {

	rank, img_url, err := sr.overwatch.SkillRank("pc", "us", btag)
	if err != nil {
		logger.Errore(err)
		return reply(s, m, "Error looking up skill rank for %q",
			btag)
	}
	return reply(s, m, "Skill rank for %q: %d (%s).", btag, rank,
		imageUrlToName(img_url))
}
