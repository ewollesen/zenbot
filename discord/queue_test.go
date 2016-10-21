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
	"testing"

	"github.com/ewollesen/discordgo"

	memorycache "github.com/ewollesen/zenbot/cache/memory"
	"github.com/ewollesen/zenbot/overwatch/mockoverwatch"
	memoryqueue "github.com/ewollesen/zenbot/queue/memory"
	"github.com/ewollesen/zenbot/ratelimiter/mocklimiter"
)

func TestHandleClear(t *testing.T) {
	test := newDiscordTest(t)

	c := memorycache.New()
	qh := newQueueHandler(newBattleTagQueue(memoryqueue.New()),
		NewBattleTagCache(c), mockoverwatch.NewRandom())
	s := test.mockSession()
	m := test.testMessage("!queue clear")
	test.AssertNil(qh.Handle(s, m, "!queue", "clear"))
}

func TestIsPermitted(t *testing.T) {
	test := newDiscordTest(t)
	s := test.mockSession()
	m := test.testMessage("!queue clear")

	granted, err := isPermitted(s, m, discordgo.PermissionKickMembers)
	test.AssertNil(err)
	test.Assert(!granted)

	s.grantPermission(discordgo.PermissionKickMembers)
	granted, err = isPermitted(s, m, discordgo.PermissionKickMembers)
	test.AssertNil(err)
	test.Assert(granted)
}

func TestClearEnqueueRateLimits(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!queue clear")

	calls := 0
	f := func(s Session, m *discordgo.MessageCreate) error {
		calls++
		return nil
	}
	rl := mocklimiter.New()
	qh.enqueue_rl = rl
	qh.clearEnqueueRateLimits(f)(s, m)
	test.AssertEqual(calls, 1)
	test.AssertEqual(rl.Clears, 1)
}

func TestEnqueueRateLimited(t *testing.T) {
	test := newDiscordTest(t)
	qh := newQueueHandler(newBattleTagQueue(memoryqueue.New()),
		NewBattleTagCache(memorycache.New()), mockoverwatch.NewRandom())
	s := test.mockSession()
	m := test.testMessage("!enqueue example#1234")

	called := 0
	f := func(s Session, m *discordgo.MessageCreate) (bool, error) {
		called++
		return true, nil
	}

	test.AssertNil(qh.enqueueRateLimited(s, m, f))
	test.AssertEqual(called, 1)

	test.AssertNil(qh.enqueueRateLimited(s, m, f))
	test.AssertEqual(called, 1)
	test.AssertContainsRe(s.sends, "You may enqueue at most once every")

	test.AssertNil(qh.enqueue_rl.Clear())

	test.AssertNil(qh.enqueueRateLimited(s, m, f))
	test.AssertEqual(called, 2)
}

func TestHandleClearUnsafe(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!enqueue example#1234")

	test.enqueue(qh.wrapBattleTag(s, m, string(testBattleTag)))

	test.AssertNil(qh.handleClearUnsafe(s, m))
	size, err := qh.q.Size()
	test.AssertNil(err)
	test.AssertEqual(size, 0)
}

func TestHandleDequeue(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!dequeue")
	s.setMember(testGuildId, testUserId, &discordgo.Member{
		Nick: "nick-without-battletag",
	})

	test.AssertNil(qh.handleDequeue(s, m))
	test.AssertContainsRe(s.sends,
		"No BattleTag specified. Try `!dequeue example#1234`.")

	s.setMember(testGuildId, testUserId, &discordgo.Member{
		Nick: testNick,
	})
	test.enqueue(qh.wrapBattleTag(s, m, string(testBattleTag)))

	test.AssertNil(qh.handleDequeue(s, m))
	size, err := qh.q.Size()
	test.AssertNil(err)
	test.AssertEqual(size, 0)

	test.AssertContains(s.sends,
		"Dequeued \"example#1234\" from the scrimmages queue.")

	s.clearSends()
	test.AssertNil(qh.handleDequeue(s, m))
	test.AssertContains(s.sends,
		"BattleTag \"example#1234\" was not found in "+
			"the scrimmages queue.")
}

func TestHandleEnqueueUnlimited(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!enqueue")
	s.setMember(testGuildId, testUserId, &discordgo.Member{
		Nick: "nick-without-battletag",
	})
	success, err := qh.handleEnqueueUnlimited(s, m)
	test.AssertNil(err)
	test.Assert(!success)
	test.AssertContainsRe(s.sends,
		"No BattleTag specified. Try `!enqueue example#1234`.")

	m = test.testMessage("!enqueue example#1234")
	success, err = qh.handleEnqueueUnlimited(s, m)
	test.AssertNil(err)
	test.Assert(success)
	test.AssertContains(s.sends,
		"Enqueued \"example#1234\" in the scrimmages queue "+
			"in position 1.")
}

func TestHandleAddUnsafe(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!queue add")

	err := qh.handleAddUnsafe(s, m)
	test.AssertNil(err)
	test.AssertContains(s.sends,
		"No BattleTag specified. Try `!queue add example#1234`.")

	s.clearSends()
	m = test.testMessage("!queue add example#1234")

	err = qh.handleAddUnsafe(s, m)
	test.AssertNil(err)
	test.AssertContains(s.sends,
		"Enqueued \"example#1234\" in the scrimmages queue "+
			"in position 1.")

	s.clearSends()
	err = qh.handleAddUnsafe(s, m)
	test.AssertNil(err)
	test.AssertContainsRe(s.sends, "BattleTag.*already enqueued.*position 1.")
}

func TestHandleKickUnsafe(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	s.setMember(testGuildId, testUserId, &discordgo.Member{
		Nick: "nick-without-battletag",
	})

	m := test.testMessage("!queue kick")
	test.AssertNil(qh.handleKickUnsafe(s, m))
	test.AssertContains(s.sends,
		"No BattleTag specified. Try `!queue kick example#1234`.")

	s.clearSends()
	m = test.testMessage("!queue kick example#1234")
	test.AssertNil(qh.handleKickUnsafe(s, m))
	test.AssertContains(s.sends,
		"BattleTag \"example#1234\" was not found in the "+
			"scrimmages queue.")

	test.enqueue(qh.wrapBattleTag(s, m, string(testBattleTag)))
	s.clearSends()
	m = test.testMessage("!queue kick example#1234")
	test.AssertNil(qh.handleKickUnsafe(s, m))
	test.AssertContains(s.sends,
		"Kicked \"example#1234\" from the scrimmages queue.")
}

func TestHandleList(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()
	m := test.testMessage("!queue list")

	test.AssertNil(qh.handleList(s, m))
	test.AssertContains(s.sends, "The scrimmages queue is empty.")

	test.enqueue(qh.wrapBattleTag(s, m, string(testBattleTag)))
	s.clearSends()
	test.AssertNil(qh.handleList(s, m))
	test.AssertContains(s.sends, "The scrimmages queue contains "+
		"1 BattleTags: example#1234.")

	m.Author.ID = "test-user-456"
	test.enqueue(qh.wrapBattleTag(s, m, "example#5678"))
	s.clearSends()
	test.AssertNil(qh.handleList(s, m))
	test.AssertContains(s.sends, "The scrimmages queue contains "+
		"2 BattleTags: example#1234, example#5678.")
}

func TestTakeUnsafe(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()

	m := test.testMessage("!queue take")
	test.AssertNil(qh.handleTakeUnsafe(s, m))
	test.AssertContainsRe(s.sends, "Took 0 BattleTags.* 0 BattleTags remain")

	m = test.testMessage("!queue take -5")
	test.AssertNil(qh.handleTakeUnsafe(s, m))
	test.AssertContains(s.sends, "Specified number of BattleTags to "+
		"take (-5) must be > 0.")

	users := generateUsers(15)

	test.enqueue(users[0:5]...)
	m = test.testMessage("!queue take")
	test.AssertNil(qh.handleTakeUnsafe(s, m))
	test.AssertContainsRe(s.sends, "Took 5 BattleTags.* 0 BattleTags remain")

	test.enqueue(users[0:13]...)
	m = test.testMessage("!queue take")
	test.AssertNil(qh.handleTakeUnsafe(s, m))
	test.AssertContainsRe(s.sends, "Took 12 BattleTags.* 1 BattleTags remain")

	test.enqueue(users[0:3]...)
	m = test.testMessage("!queue take 2")
	test.AssertNil(qh.handleTakeUnsafe(s, m))
	test.AssertContainsRe(s.sends, "Took 2 BattleTags.* 2 BattleTags remain")
	test.AssertContainsRe(s.sends, ": "+users[12].BattleTag+", "+users[0].BattleTag)
}

func TestHelp(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()

	m := test.testMessage("!queue help")
	test.AssertNil(qh.Handle(s, m, "queue", "help"))
	test.AssertContainsRe(s.sends, "Manipulates the scrimmages queue")

	m = test.testMessage("!queue")
	test.AssertNil(qh.Handle(s, m, "queue"))
	test.AssertContainsRe(s.sends, "Manipulates the scrimmages queue")
}

func TestDoubleEnqueue(t *testing.T) {
	test, qh := newQueueTest(t)
	s := test.mockSession()

	m := test.testMessage("!enqueue example#1234")
	success, err := qh.handleEnqueueUnlimited(s, m)
	test.AssertNil(err)
	test.Assert(success)
	test.AssertContains(s.sends,
		"Enqueued \"example#1234\" in the scrimmages queue "+
			"in position 1.")

	s.clearSends()
	success, err = qh.handleEnqueueUnlimited(s, m)
	test.AssertNil(err)
	test.Assert(!success)
	test.AssertContainsRe(s.sends,
		"BattleTag \"example#1234\" is already enqueued .* position 1.")

	test.Logf("Still need to handle the double enqueue issue")
	test.SkipNow()
	s.clearSends()
	m = test.testMessage("!enqueue example#5678")
	success, err = qh.handleEnqueueUnlimited(s, m)
	test.AssertNil(err)
	test.Assert(success)
	test.AssertContainsRe(s.sends,
		"BattleTag \"example#1234\" is already enqueued .* position 1.")
}

//
// Helpers
//

type queueTest struct {
	*discordTest
	handler *queueHandler
}

func newQueueTest(t *testing.T) (*queueTest, *queueHandler) {
	qh := newQueueHandler(newBattleTagQueue(memoryqueue.New()),
		NewBattleTagCache(memorycache.New()), mockoverwatch.NewRandom())

	return &queueTest{
		discordTest: newDiscordTest(t),
		handler:     qh,
	}, qh
}

func (t *queueTest) enqueue(btags ...*userBattleTag) {
	for _, btag := range btags {
		pre_size, err := t.handler.q.Size()
		t.AssertNil(err)

		pos, err := t.handler.q.Enqueue(btag)
		t.AssertNil(err)
		t.AssertEqual(pos, pre_size)
		size, err := t.handler.q.Size()
		t.AssertNil(err)
		t.AssertEqual(size, pre_size+1)
	}
}

func msgFromUserBattleTag(u *userBattleTag) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID: u.UserId,
			},
			ChannelID: testChannelId,
		},
	}
}

func generateUsers(n int) (users []*userBattleTag) {
	for i := 0; i < n; i++ {
		user := &userBattleTag{
			BattleTag: nextBattleTag(),
			UserId:    nextUserId(),
			GuildId:   testGuildId,
		}
		users = append(users, user)
	}

	return users
}

var userIds = 0

func nextUserId() string {
	userIds++
	return fmt.Sprintf("test-user-%03d", userIds)
}

var battleTags = 0

func nextBattleTag() string {
	battleTags++
	return fmt.Sprintf("example#%04d", battleTags)
}
