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
	"github.com/ewollesen/zenbot/blizzard"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/queue"
	"github.com/ewollesen/zenbot/ratelimiter"
	"github.com/ewollesen/zenbot/ratelimiter/concretelimiter"
	"github.com/ewollesen/zenbot/util"
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
		"`!queue kick example#1234` - removes a BattleTag from the scrimmages queue (admin-only)",
		"`!queue list` - lists the BattleTags in the scrimmages queue.",
		"`!queue teams` - splits the BattleTags into two teams by Skill Rank.",
		"`!queue take <n>` - takes the first <n> BattleTags from the scrimmages queue (default: 12, admin-only)",
		"`!queue help` - displays this help message",
	}, "\n")
	PermissionDenied = Error.NewClass("permission denied")
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
		err = h.enqueueRateLimited(s, m, h.handleEnqueueUnlimited)
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
			reply(s, m, helpMsg)
		}
	}

	return err
}

func (h *queueHandler) Help(argv ...string) string {
	term := strings.Join(argv, " ")
	wrap := func(msg string) string {
		return fmt.Sprintf("`!%s` - %s. See `!queue help` for more info",
			term, msg)
	}
	return wrap("manipulates the scrimmages queue")
}

func (h *queueHandler) handleClearUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	err = h.q.Clear()
	if err != nil {
		return err
	}

	reply(s, m, "Scrimmages queue cleared.")
	return nil
}

func (h *queueHandler) cacheBattleTag(s Session, m *discordgo.MessageCreate,
	btag string) error {

	return h.btags.Set(userKey(s, m), btag)
}

func (h *queueHandler) lookupBattleTag(s Session, m *discordgo.MessageCreate) (
	btag string, err error) {

	return h.btags.Get(userKey(s, m))
}

// Look for a BattleTag in the cache, or in the User's nick. No validation is
// performed.
func (h *queueHandler) discoverBattleTag(s Session, m *discordgo.MessageCreate) (
	btag string, err error) {

	btag, err = h.lookupBattleTag(s, m)
	if err == nil && btag != "" { // WARNING: FLIPPED
		return btag, nil
	}
	logger.Warne(err)

	nick := h.lookupNickOrUsername(s, m)
	if btag = blizzard.FirstBattleTag(nick); btag == "" {
		return "", overwatch.BattleTagNotFound.New(nick)
	}

	return btag, nil
}

// TODO: cache these? how do we invalidate the cache?
func (h *queueHandler) lookupNick(s Session, m *discordgo.MessageCreate) (
	nick string, err error) {

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return "", err
	}

	member, err := s.Member(channel.GuildID, m.Author.ID)
	if err != nil {
		return "", err
	}

	if member.Nick == "" {
		return m.Author.Username, nil
	}

	return member.Nick, nil
}

func (h *queueHandler) lookupNickOrUsername(s Session,
	m *discordgo.MessageCreate) string {

	nick, err := h.lookupNick(s, m)
	if err != nil {
		logger.Warne(err)
		nick = m.Author.Username
	}

	return nick
}

func (h *queueHandler) lookupSkillRank(btag string) {
	go func(btag string) {
		_, _, err := h.overwatch.SkillRank("pc", "us", btag)
		logger.Warne(err)
	}(btag)
}

func (h *queueHandler) handleDequeue(s Session,
	m *discordgo.MessageCreate) (err error) {

	// Just ignore any text after the command. Users may only dequeue
	// themselves. Admins can use the `!queue kick` command to remove people
	// other than themselves.

	nick := h.lookupNickOrUsername(s, m)
	btag, err := h.discoverBattleTag(s, m)
	if err != nil {
		reply(s, m, "Error finding BattleTag for %s.", nick)
		return err
	}

	err = h.q.Remove(h.wrapBattleTag(s, m, string(btag)))
	if err != nil {
		if queue.NotFound.Contains(err) {
			reply(s, m, "BattleTag %s was not found in the "+
				"scrimmages queue.", btag)
			return err
		}
		reply(s, m, "Error dequeueing %s from the scrimmages queue. "+
			"Please try again.", btag)
		return err
	}

	reply(s, m, "Dequeued %s (%s) from the scrimmages queue.", btag, nick)
	return nil
}

func (h *queueHandler) handleEnqueueUnlimited(s Session,
	m *discordgo.MessageCreate) (err error) {

	btag := ""
	args := strings.Split(m.Content, " ")
	nick := h.lookupNickOrUsername(s, m)

	if len(args) > 1 {
		text := strings.Join(args[1:], " ")
		if btag = blizzard.FirstBattleTag(text); btag == "" {
			reply(s, m, "Invalid BattleTag %q. "+
				"Try `!enqueue example#1234`.", text)
			return overwatch.BattleTagInvalid.New(text)
		}
	} else {
		btag, err = h.discoverBattleTag(s, m)
		if err != nil {
			if overwatch.BattleTagNotFound.Contains(err) {
				reply(s, m, "No BattleTag specified. "+
					"Try `!enqueue example#1234`.")
				return err
			}
			reply(s, m, "Error finding BattleTag for %s.", nick)
			return err
		}
	}

	valid, err := h.overwatch.IsValidBattleTag("pc", "us", btag)
	if err != nil {
		reply(s, m, "Error validating BattleTag %q. "+
			"Please try again.", btag)
		return err
	}
	if !valid {
		reply(s, m, "Invalid BattleTag %q. "+
			"Try `!enqueue example#1234`. "+
			"Remember, BattleTags are CaSe-SeNsItIvE!", btag)
		return overwatch.BattleTagInvalid.New(btag)
	}

	btag_in_cache, err := h.lookupBattleTag(s, m)
	if err != nil {
		reply(s, m,
			"Error enqueuing BattleTag %q. Please try again.", btag)
		return err
	}
	if btag_in_cache != btag {
		// Its possible that the user is already enqueued with a
		// different BattleTag, so let's check...
		pos, err := h.q.Position(h.wrapBattleTag(s, m, btag_in_cache))
		if err != nil {
			reply(s, m, "Error enqueueing BattleTag %q. "+
				"Please try again.")
			return err
		}
		if pos != -1 {
			reply(s, m, "%s is already enqueued in the scrimmages "+
				"queue with BattleTag %s. Please `!dequeue` "+
				"before re-enqueueing with a different "+
				"BattleTag.", nick, btag_in_cache)
			msg := fmt.Sprintf("%+v in position %d",
				btag_in_cache, pos+1)
			return queue.AlreadyEnqueued.NewWith(msg,
				queue.SetPosition(pos))
		}
	} // There's a race condition here, but not worth worrying about.

	pos, err := h.q.Enqueue(h.wrapBattleTag(s, m, btag))
	if err != nil {
		if queue.AlreadyEnqueued.Contains(err) {
			reply(s, m, "BattleTag %s (%s) is already enqueued "+
				"in the scrimmages queue in position %d.",
				btag, nick, queue.GetPosition(err)+1)
			return err
		}
		reply(s, m, "Error enqueueing %s (%s) into the scrimmages "+
			"queue. Please try again.", btag, nick)
		return err
	}

	logger.Errore(h.cacheBattleTag(s, m, btag))
	h.lookupSkillRank(btag)

	reply(s, m, "Enqueued %s (%s) in the scrimmages queue in position %d.",
		btag, nick, pos+1)
	return nil
}

func (h *queueHandler) handleAddUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	args := strings.Split(m.Content, " ")
	if len(args) < 3 {
		reply(s, m, "No BattleTag specified. "+
			"Try `!queue add example#1234`.")
		return nil
	}

	btags := blizzard.FindBattleTags(m.Content)
	if len(btags) == 0 {
		reply(s, m, "No valid BattleTags specified."+
			"Try `!queue add example#1234`.")
		return nil
	}

	added_btags := []string{}
	for _, btag := range btags {
		valid, err := h.overwatch.IsValidBattleTag("pc", "us", btag)
		if err != nil {
			reply(s, m, "Error validating BattleTag %q.")
			logger.Errore(err)
			continue
		}
		if !valid {
			reply(s, m, "Invalid BattleTag %q. "+
				"Try `!queue add example#1234`. "+
				"Remember, BattleTags are CaSe-SeNsItIvE!", btag)
			continue
		}

		_, err = h.q.Enqueue(h.wrapBattleTag(s, m, btag))
		if err != nil {
			if queue.AlreadyEnqueued.Contains(err) {
				reply(s, m, "BattleTag %q is already enqueued "+
					"in the scrimmages queue in position "+
					"%d.", btag, queue.GetPosition(err)+1)
				continue
			}
			reply(s, m, "Error adding %q to the scrimmages queue.", btag)
			logger.Errore(err)
			continue
		}

		added_btags = append(added_btags, btag)
		// DO NOT cache BattleTags added in this way.
		h.lookupSkillRank(btag)
	}

	reply(s, m, "Added %s to the scrimmages queue.",
		util.ToList(added_btags))
	return nil
}

func (h *queueHandler) handleKickUnsafe(s Session,
	m *discordgo.MessageCreate) (err error) {

	args := strings.Split(m.Content, " ")
	if len(args) < 3 {
		reply(s, m, "No BattleTag specified. "+
			"Try `!queue kick example#1234`.")
		return nil
	}

	btags := blizzard.FindBattleTags(m.Content)
	if len(btags) == 0 {
		reply(s, m, "No valid BattleTags specified."+
			"Try `!queue kick example#1234`.")
		return nil
	}

	// There's no need to validate the BattleTags, since the user is kicking
	// them. In theory, if the BattleTag is invalid, it won't be in the
	// queue, so no harm, no foul.
	kicked_btags := []string{}
	for _, btag := range btags {
		if err = h.q.Remove(h.wrapBattleTag(s, m, btag)); err != nil {
			if queue.NotFound.Contains(err) {
				reply(s, m, "BattleTag %q was not found in "+
					"the scrimmages queue.", btag)
				continue
			}
			logger.Warne(err)
			reply(s, m, "Error kicking %s from the scrimmages "+
				"queue.", btag)
			continue
		}
		kicked_btags = append(kicked_btags, btag)
	}

	reply(s, m, "Kicked %s from the scrimmages queue.",
		util.ToList(kicked_btags))
	return nil
}

func (h *queueHandler) handleList(s Session,
	m *discordgo.MessageCreate) (err error) {

	btags, err := h.queueBattleTags()
	if err != nil {
		return err
	}
	if len(btags) == 0 {
		reply(s, m, "The scrimmages queue is empty.")
		return nil
	}
	reply(s, m, "The scrimmages queue contains %d "+
		"BattleTags: %s", len(btags), util.ToList(btags))
	return nil
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
		reply(s, m, "Specified number of BattleTags to take "+
			"(%d) must be > 0.", num_to_take)
		return nil
	}

	taken, num_left, err := h.q.DequeueN(num_to_take)
	if err != nil {
		return err
	}

	if len(taken) == 0 {
		reply(s, m, "Took 0 BattleTags from the scrimmages "+
			"queue. %d BattleTags remain in the scrimmages queue.",
			num_left)
		return nil
	}

	btags := []string{}
	for _, btag := range taken {
		btags = append(btags, btag.BattleTag)
	}

	// TODO: move me to a wrapper?
	go func() { logger.Warne(h.replyPartition(s, m, btags)) }()

	reply(s, m, "Took %d BattleTags from the scrimmages queue: %s. "+
		"%d BattleTags remain in the scrimmages queue.",
		len(taken), util.ToList(btags), num_left)
	return nil
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
		reply(s, m, "The scrimmages queue is empty.")
		return nil
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
		reply(s, m, "Permission denied.")
		return PermissionDenied.New("")
	}

	return fn(s, m)
}

func (h *queueHandler) enqueueRateLimited(s Session, m *discordgo.MessageCreate,
	handler bareHandler) error {

	trigger, err := h.enqueue_rl.Limit(userKey(s, m))
	if err != nil {
		if ratelimiter.TooSoon.Contains(err) {
			reply(s, m,
				"You may enqueue at most once every %d "+
					"minutes, %s. Please try again later.",
				*enqueueRateLimit/time.Minute,
				mention(m.Author.ID))
			return err
		}
		reply(s, m, "Error checking enqueue rate limits. "+
			"Please try again.")
		return err
	}

	if err := handler(s, m); err != nil {
		return err
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
