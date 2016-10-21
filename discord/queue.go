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
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ewollesen/discordgo"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/queue"
	"github.com/ewollesen/zenbot/ratelimiter"
	"github.com/ewollesen/zenbot/ratelimiter/concretelimiter"
)

var (
	enqueueRateLimit = flag.Duration("discord.enqueue_rate_limit",
		5*time.Minute, "minimum duration between enqueue attempts")
	helpMsg = strings.Join([]string{
		"Manipulates the scrimmages queue.",
		"`!dequeue` - removes your BattleTag from the scrimmages queue",
		"`!enqueue example#1234` - adds your BattleTag to the scrimmages queue",
		"`!queue add example#1234` - adds a BattleTag to the scrimmages queue (admin-only)",
		"`!queue clear` - clears the scrimmages queue (admin-only)",
		"`!queue kick example#1234` - removes a BattleTag from the scrimmages queue (admni-only)",
		"`!queue list` - lists the BattleTags in the scrimmages queue.",
		"`!queue teams` - splits the BattleTags into two teams by Skill Rank.",
		"`!queue take <n>` - takes the first <n> BattleTags from the scrimmages queue (default: 12, admin-only)",
		"`!queue help` - displays this help message",
	}, "\n")
	BTagNotFound = Error.NewClass("couldn't find a BattleTag for rank")
)

type queueHandler struct {
	q          *BattleTagQueue
	btags      *BattleTagCache
	enqueue_rl ratelimiter.RateLimiter
	overwatch  overwatch.OverwatchAPI
}

var _ DiscordHandler = (*queueHandler)(nil)

func newQueueHandler(q *BattleTagQueue, b *BattleTagCache,
	o overwatch.OverwatchAPI) *queueHandler {

	return &queueHandler{
		btags:      b,
		q:          q,
		enqueue_rl: concretelimiter.New(*enqueueRateLimit),
		overwatch:  o,
	}
}

func (h *queueHandler) Handle(s Session, m *discordgo.MessageCreate,
	argv ...string) (err error) {

	cmd := argv[0]
	switch cmd {
	case "dequeue":
		err = h.handleDequeue(s, m)
	case "enqueue":
		err = h.enqueueRateLimited(s, m,
			h.handleEnqueueLookupSkillRank(h.handleEnqueueUnlimited))
	case "queue":
		sub_cmd := "help"
		if len(argv) > 1 {
			sub_cmd = strings.ToLower(argv[1])
		}
		switch sub_cmd {
		case "add":
			err = h.auth2KickRequired(s, m, h.handleAddUnsafe)
		case "clear":
			err = h.auth2KickRequired(s, m,
				h.clearEnqueueRateLimits(h.handleClearUnsafe))
		case "kick", "remove":
			err = h.auth2KickRequired(s, m, h.handleKickUnsafe)
		case "list":
			err = h.handleList(s, m)
		case "partition", "teams":
			err = h.handleQueuePartition(s, m)
		case "take", "pick":
			err = h.auth2KickRequired(s, m, h.handleTakeUnsafe)
		default:
			err = reply(s, m, helpMsg)
		}
	}

	return err
}

func (h *queueHandler) Help(argv ...string) string {
	return "manipulates the scrimmages queue"
}

func (h *queueHandler) handleClearUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	err = h.q.Clear()
	if err != nil {
		return err
	}

	return reply(s, m, "Scrimmages queue cleared.")
}

func (h *queueHandler) handleDequeue(s Session,
	m *discordgo.MessageCreate) (err error) {

	// TODO: don't allow passing someone else's battle tag on the command line
	btag, err := findBattleTag(h.overwatch, h.btags, s, m)
	if err != nil {
		if overwatch.BattleTagNotFound.Contains(err) {
			return reply(s, m, "No BattleTag specified. "+
				"Try `!dequeue example#1234`. "+
				"Remember, BattleTags are CaSe-SeNsItIvE!")
		}
		if overwatch.BattleTagInvalid.Contains(err) {
			return reply(s, m, "Invalid BattleTag. "+
				"Try `!dequeue example#1234`.")
		}
		logger.Errore(err)
		return reply(s, m, "Error parsing BattleTag.")
	}

	err = h.q.Remove(h.wrapBattleTag(s, m, string(btag)))
	if err != nil {
		if queue.NotFound.Contains(err) {
			return reply(s, m, "BattleTag %q was not found in the "+
				"scrimmages queue.", btag)
		}
		return err
	}

	return reply(s, m, "Dequeued %q from the scrimmages queue.", btag)
}

func (h *queueHandler) handleEnqueueUnlimited(s Session,
	m *discordgo.MessageCreate) (success bool, err error) {

	btag, err := h.lookupSkillRankWrapper(findBattleTag(h.overwatch, h.btags, s, m))
	if err != nil {
		if overwatch.BattleTagNotFound.Contains(err) {
			return false, reply(s, m, "No BattleTag specified. "+
				"Try `!enqueue example#1234`. "+
				"Remember, BattleTags are CaSe-SeNsItIvE!")
		}
		if overwatch.BattleTagInvalid.Contains(err) {
			return false, reply(s, m, "Invalid BattleTag. "+
				"Try `!enqueue example#1234`. "+
				"Remember, BattleTags are CaSe-SeNsItIvE!")
		}
		logger.Errore(err)
		return false, reply(s, m, "Error parsing BattleTag.")
	}

	logger.Warne(h.btags.Set(userKey(s, m), btag))

	pos, err := h.q.Enqueue(h.wrapBattleTag(s, m, string(btag)))
	if err != nil {
		if queue.AlreadyEnqueued.Contains(err) {
			return false, reply(s, m, "BattleTag %q is already enqueued "+
				"in the scrimmages queue in position %d.",
				btag, queue.GetPosition(err)+1)
		}
		return false, err
	}

	return true, reply(s, m, "Enqueued %q in the scrimmages queue in position %d.",
		btag, pos+1)
}

func (h *queueHandler) handleEnqueueLookupSkillRank(fn successHandler) successHandler {
	return func(s Session, m *discordgo.MessageCreate) (bool, error) {
		success, real_err := fn(s, m)
		if success {
			btag, err := h.btags.Get(userKey(s, m))
			if err != nil {
				logger.Warne(err)
				return success, real_err
			}
			_, _, err = h.overwatch.SkillRank("pc", "us", btag)
			logger.Warne(err)
		}

		return success, real_err
	}
}

func (h *queueHandler) handleAddUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	btags := []string{}
	for _, btag := range parseAllBattleTags(m.Content) {
		if btag == "" {
			continue
		}

		btags = append(btags, h.lookupSkillRank(btag))
	}

	if len(btags) == 0 {
		return reply(s, m, "No BattleTag specified. "+
			"Try `!queue add example#1234`.")
	}

	for _, btag := range btags {
		pos, err := h.q.Enqueue(h.wrapBattleTag(s, m, btag))
		if err != nil {
			if queue.AlreadyEnqueued.Contains(err) {
				logger.Warne(reply(s, m, "BattleTag %q is already enqueued "+
					"in the scrimmages queue in position %d.",
					btag, queue.GetPosition(err)+1))
				continue
			}
			logger.Errore(err)
			return reply(s, m, "Error adding %q to the scrimmages queue.", btag)
		}
		logger.Errore(reply(s, m, "Enqueued %q in the scrimmages queue in position %d.",
			btag, pos+1))
	}

	return nil
}

func (h *queueHandler) handleKickUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	btag := parseBattleTag(m.Content)
	if btag == "" {
		return reply(s, m, "No BattleTag specified. "+
			"Try `!queue kick example#1234`.")
	}

	err = h.q.Remove(h.wrapBattleTag(s, m, btag))
	if err != nil {
		if queue.NotFound.Contains(err) {
			return reply(s, m, "BattleTag %q was not found in the "+
				"scrimmages queue.", btag)
		}
		return err
	}

	return reply(s, m, "Kicked %q from the scrimmages queue.", btag)
}

func (h *queueHandler) handleList(s Session,
	m *discordgo.MessageCreate) (err error) {

	btags, err := h.queueBattleTags()
	if err != nil {
		return err
	}
	if len(btags) == 0 {
		return reply(s, m, "The scrimmages queue is empty.")
	}
	return reply(s, m, "The scrimmages queue contains %d "+
		"BattleTags: %s.", len(btags), strings.Join(btags, ", "))
}

func (h *queueHandler) handleTakeUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {
	num_to_take := 12

	argv := strings.Split(m.Content, " ")
	if len(argv) > 2 {
		n, err := strconv.ParseInt(argv[2], 10, 32)
		logger.Warne(err)
		if err == nil {
			num_to_take = int(n)
		}
	}

	if num_to_take <= 0 {
		return reply(s, m, "Specified number of BattleTags to take "+
			"(%d) must be > 0.", num_to_take)
	}

	taken, num_left, err := h.q.DequeueN(num_to_take)
	if err != nil {
		return err
	}

	if len(taken) == 0 {
		return reply(s, m, "Took 0 BattleTags from the scrimmages "+
			"queue. %d BattleTags remain in the scrimmages queue.",
			num_left)
	}

	btags := []string{}
	for _, btag := range taken {
		btags = append(btags, btag.BattleTag)
	}

	// TODO: move me to a wrapper
	go func() { logger.Warne(h.replyPartition(s, m, btags)) }()

	return reply(s, m, "Took %d BattleTags from the scrimmages queue: %s. "+
		"%d BattleTags remain in the scrimmages queue.",
		len(taken), strings.Join(btags, ", "), num_left)
}

func (h *queueHandler) queueBattleTags() (btags []string, err error) {
	err = h.q.Iter(func(index int, btag *userBattleTag) bool {
		btags = append(btags, btag.BattleTag)
		return false
	})

	return btags, err
}

func (h *queueHandler) handleQueuePartition(s Session,
	m *discordgo.MessageCreate) (err error) {

	btags, err := h.queueBattleTags()
	if err != nil {
		return err
	}
	if len(btags) == 0 {
		return reply(s, m, "The scrimmages queue is empty.")
	}
	go func() { logger.Warne(h.replyPartition(s, m, btags)) }()

	return nil
}

func (h *queueHandler) auth2KickRequired(s Session,
	m *discordgo.MessageCreate, fn bareHandler) error {

	granted, err := isPermitted(s, m, discordgo.PermissionKickMembers)
	if err != nil {
		logger.Errore(err)
		return err
	}

	if !granted {
		return reply(s, m, "Permission denied.")
	}

	return fn(s, m)
}

func (h *queueHandler) enqueueRateLimited(s Session,
	m *discordgo.MessageCreate, fn successHandler) error {

	trigger, err := h.enqueue_rl.Limit(userKey(s, m))
	if err != nil {
		if ratelimiter.TooSoon.Contains(err) {
			return reply(s, m,
				"You may enqueue at most once every %d "+
					"minutes, %s. Please try again later.",
				*enqueueRateLimit/time.Minute,
				mention(m.Author.ID))
		}
		return err
	}

	success, err := fn(s, m)
	if err != nil {
		return err
	}
	if !success {
		return nil
	}

	return trigger()
}

type userBattleTag struct {
	BattleTag string
	GuildId   string
	UserId    string
}

func (h *queueHandler) wrapBattleTag(s Session, m *discordgo.MessageCreate,
	btag string) *userBattleTag {

	ch, err := s.Channel(m.ChannelID)
	if err != nil {
		logger.Errore(err)
		return nil
	}

	return &userBattleTag{
		BattleTag: btag,
		GuildId:   ch.GuildID,
		UserId:    m.Author.ID,
	}
}

type bareHandler func(Session, *discordgo.MessageCreate) error
type successHandler func(Session, *discordgo.MessageCreate) (bool, error)

func (h *queueHandler) clearEnqueueRateLimits(fn bareHandler) bareHandler {
	return func(s Session, m *discordgo.MessageCreate) (err error) {
		if err = fn(s, m); err != nil {
			return err
		}

		return h.enqueue_rl.Clear()
	}
}

func mention(user_id string) string {
	return fmt.Sprintf("<@!%s>", user_id)
}

func (h *queueHandler) replyPartition(s Session,
	m *discordgo.MessageCreate, btags []string) error {

	return replyPartition(s, m, h.overwatch, btags)
}

func (h *queueHandler) lookupSkillRank(btag string) string {
	if !validBattleTag(btag) {
		return btag
	}

	go func() {
		_, _, err := h.overwatch.SkillRank("pc", "us", btag)
		if err != nil {
			logger.Warne(err)
			return
		}
	}()

	return btag
}

func (h *queueHandler) lookupSkillRankWrapper(btag string, err error) (
	string, error) {

	if err != nil {
		return btag, err
	}

	return h.lookupSkillRank(btag), nil
}
