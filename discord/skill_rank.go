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
	"sort"
	"strings"

	"github.com/ewollesen/discordgo"
	"github.com/ewollesen/zenbot/blizzard"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/partition"
	"github.com/ewollesen/zenbot/util"
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

	TooManyLookupFailures = Error.NewClass("too many skill rank lookup failures")
	BTagNotFound          = Error.NewClass("couldn't find a BattleTag for rank")
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

	cmd := argv[0]
	switch cmd {
	case "sr":
		sub_cmd := "help"
		if len(argv) > 1 {
			sub_cmd = argv[1]
		}
		switch sub_cmd {
		case "help":
			reply(s, m, skillRankHelpMsg)
		default:
			err = sr.handleSkillRank(s, m, sub_cmd)
		}
	case "teams":
		err = sr.handleTeams(s, m)
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
		reply(s, m, "Error looking up skill rank for %s "+
			"(remember, BattleTags are CaSe-SeNsItIvE!)", btag)
		return err
	}
	reply(s, m, "Skill rank for %s: %d (%s).", btag, rank,
		imageUrlToName(img_url))
	return nil
}

func (sr *skillRankHandler) handleTeams(s Session,
	m *discordgo.MessageCreate) (err error) {

	return sr.replyPartition(s, m, blizzard.FindBattleTags(m.Content))
}

// TODO: DRY up with queueHandler's version
func (sr *skillRankHandler) replyPartition(s Session,
	m *discordgo.MessageCreate, btags []string) error {

	return replyPartition(s, m, sr.overwatch, btags)
}

func averageRank(ranks []int) int {
	sum := 0
	for _, rank := range ranks {
		sum += rank
	}
	return sum / len(ranks)
}

type rankBtagPair struct {
	Rank      int
	BattleTag string
}

// TODO: DRY up with the queueHandler's version
func partitionBattleTags(ow overwatch.OverwatchAPI, btags []string) (
	team_one, team_two []*rankBtagPair, err error) {

	ranks := make(map[int][]string)
	all_ranks := []int{}
	failures := 0
	no_ranks := []string{}

	// Optimization: parallelize
	for _, btag := range btags {
		rank, _, err := ow.SkillRank("pc", "us", btag)
		if err != nil {
			logger.Errore(err)
			failures++
			if failures >= len(btags)/4 {
				return nil, nil, TooManyLookupFailures.Wrap(err)
			}
			no_ranks = append(no_ranks, btag)
			continue
		}

		all_ranks = append(all_ranks, rank)

		_, ok := ranks[rank]
		if ok {
			ranks[rank] = append(ranks[rank], btag)
		} else {
			ranks[rank] = []string{btag}
		}
	}

	if len(no_ranks) > 0 {
		// Loop through the BattleTags for which we couldn't look up a
		// skill rank, and give them the average of the other players.
		average_rank := averageRank(all_ranks)
		for _, btag := range no_ranks {
			logger.Debugf("assigning average rank (%d) for btag %s",
				average_rank, btag)
			all_ranks = append(all_ranks, average_rank)
			_, ok := ranks[average_rank]
			if ok {
				ranks[average_rank] =
					append(ranks[average_rank], btag)
			} else {
				ranks[average_rank] = []string{btag}
			}
		}
	}

	if len(all_ranks) != len(btags) {
		logger.Warnf("something is weird, I don't have a rank for "+
			"each battle tag (%d vs %d)", len(all_ranks), len(btags))
	}

	logger.Debugf("all_ranks: %v", all_ranks)

	team_one_ranks, team_two_ranks := partition.Partition(all_ranks)
	for _, team_one_rank := range team_one_ranks {
		btags, ok := ranks[team_one_rank]
		if !ok || len(btags) == 0 {
			return nil, nil, BTagNotFound.New("%d", team_one_rank)
		}
		team_one = append(team_one, &rankBtagPair{
			BattleTag: btags[0],
			Rank:      team_one_rank,
		})
		ranks[team_one_rank] = btags[1:]
	}
	for _, team_two_rank := range team_two_ranks {
		btags, ok := ranks[team_two_rank]
		if !ok || len(btags) == 0 {
			return nil, nil, BTagNotFound.New("%d", team_two_rank)
		}
		team_two = append(team_two, &rankBtagPair{
			BattleTag: btags[0],
			Rank:      team_two_rank,
		})
		ranks[team_two_rank] = btags[1:]
	}

	return team_one, team_two, nil
}

func replyPartition(s Session, m *discordgo.MessageCreate,
	ow overwatch.OverwatchAPI, btags []string) error {

	team_one, team_two, err := partitionBattleTags(ow, btags)
	if err != nil {
		if TooManyLookupFailures.Contains(err) {
			replyPrivate(s, m, "I failed to look up Skill "+
				"Ranks for >= 25%% of the BattleTags listed, "+
				"so I'm giving up. Look up failures are often "+
				"caused by case-sensitivity errors in "+
				"BattleTags.")
		} else {
			replyPrivate(s, m, "Error partitioning into teams.")
		}
		return err
	}

	team_one_avg := 0.0
	team_one_btags := []string{}
	for _, pair := range team_one {
		team_one_avg += float64(pair.Rank)
		team_one_btags = append(team_one_btags, pair.BattleTag)
	}
	team_one_avg /= float64(len(team_one))

	team_two_avg := 0.0
	team_two_btags := []string{}
	for _, pair := range team_two {
		team_two_avg += float64(pair.Rank)
		team_two_btags = append(team_two_btags, pair.BattleTag)
	}
	team_two_avg /= float64(len(team_two))

	sort.Strings(team_one_btags)
	sort.Strings(team_two_btags)

	// Don't join with commas, they'll only cause copy pasta errors
	replyPrivate(s, m,
		`I suggest the following teams based on skill rank:
Team 1 (avg. %0.1f): %s
Team 2 (avg. %0.1f): %s`,
		team_one_avg, util.ToList(team_one_btags),
		team_two_avg, util.ToList(team_two_btags))
	return nil
}
