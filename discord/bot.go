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
	"os"
	"strings"
	"sync"
	"time"

	redis "gopkg.in/redis.v5"

	"github.com/ewollesen/discordgo"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"

	"github.com/ewollesen/zenbot/cache"
	memorycache "github.com/ewollesen/zenbot/cache/memory"
	"github.com/ewollesen/zenbot/cache/rediscache"
	"github.com/ewollesen/zenbot/commands"
	"github.com/ewollesen/zenbot/httpapi"
	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/overwatch/lootbox"
	"github.com/ewollesen/zenbot/queue"
	memoryqueue "github.com/ewollesen/zenbot/queue/memory"
	"github.com/ewollesen/zenbot/queue/redisqueue"
)

var (
	Error  = errors.NewClass("discord")
	logger = spacelog.GetLogger()

	commandPrefix = flag.String("discord.command_prefix", "!",
		"String that prefixes bot commands")
	discordToken  = flag.String("discord.token", "", "Discord bot API token")
	game          = flag.String("discord.game", "!queue help", "Game being played")
	redisKeySpace = flag.String("discord.redis_keyspace", "discord",
		"redis keyspace prefix")
)

type bot struct {
	command_handlers_mu sync.Mutex
	command_handlers    map[string]DiscordHandler

	handler_callbacks []func()
	user_id           string

	session_cache cache.Cache

	oauth_mu     sync.Mutex
	oauth_states map[string]string
}

func New(redis_client *redis.Client) *bot {
	b := &bot{
		command_handlers: make(map[string]DiscordHandler),
		oauth_states:     make(map[string]string),
	}

	b.RegisterCommand("ping", &discordHandler{commands.Pong})

	var q queue.Queue
	var c, owc cache.Cache
	if redis_client != nil {
		logger.Infof("using redis queue and cache")
		q = redisqueue.New(redis_client, *redisKeySpace+".queues.scrimmages")
		c = rediscache.New(redis_client, *redisKeySpace+".caches.battletags", 0)
		owc = rediscache.New(redis_client, *redisKeySpace+".caches.overwatch", time.Hour*2)
	} else {
		logger.Infof("using memory queue and cache")
		q = memoryqueue.New()
		c = memorycache.New()
		owc = memorycache.New()
	}

	b.session_cache = memorycache.New()

	btq := newBattleTagQueue(q)
	btc := NewBattleTagCache(c)
	qh := newQueueHandler(btq, btc)
	b.RegisterCommand("dequeue", qh)
	b.RegisterCommand("enqueue", qh)
	b.RegisterCommand("queue", qh)

	dh := newDebugHandler(btq, btc)
	b.RegisterCommand("debug", dh)

	cow := overwatch.NewCaching(lootbox.New(), owc)
	srh := newSkillRankHandler(cow)
	b.RegisterCommand("sr", srh)

	return b
}

func (b *bot) Run(quit chan os.Signal) error {
	session, err := b.logIn()
	if err != nil {
		return err
	}

	logger.Info("online")

	if *game != "" {
		logger.Warne(session.UpdateStatus(0, *game))
	}

	signal, closed := <-quit
	if closed {
		logger.Debugf("quit channel closed, shutting down")
	}
	if signal != nil {
		logger.Debugf("quit channel received signal %v, shutting down", signal)
	}

	logger.Info("shutting down")
	b.logOut(session)

	return nil
}

func (b *bot) logIn() (*discordgo.Session, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	session, err := discordgo.New(token)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	if session == nil {
		return nil, Error.New("log in failed")
	}

	if err = b.addHandlers(session); err != nil {
		return nil, Error.Wrap(err)
	}

	if err = session.Open(); err != nil {
		return nil, Error.Wrap(err)
	}

	return session, nil
}

func getToken() (token string, err error) {
	if token = os.Getenv("DISCORD_TOKEN"); token != "" {
		return token, nil
	}

	if *discordToken != "" {
		return *discordToken, nil
	}

	return "", Error.New("no discord token specified")
}

func (b *bot) addHandlers(session *discordgo.Session) error {
	b.handler_callbacks = append(b.handler_callbacks,
		session.AddHandler(b.messageHandler),
		session.AddHandler(b.presenceHandler))

	return nil
}

func (b *bot) logOut(session *discordgo.Session) {
	logger.Errore(session.Close())
	logger.Errore(b.removeHandlers())
	logger.Info("offline")
}

func (b *bot) removeHandlers() error {
	for _, cb := range b.handler_callbacks {
		if cb != nil {
			cb()
		}
	}

	return nil
}

func (b *bot) messageHandler(ds *discordgo.Session,
	m *discordgo.MessageCreate) {

	s := newCachingSession(ds, b.session_cache)

	if m.Author.ID == b.myUserId(s) {
		logger.Debugf("-> %s: %s", m.Author.Username, m.Content)
		return
	}

	private := isPrivateMessage(s, m)
	if !private && !strings.HasPrefix(m.Content, *commandPrefix) {
		return
	}
	logger.Debugf("<-%t %s: %s", private, m.Author.Username, m.Content)

	if err := b.handleCommand(s, m); err != nil {
		logger.Warne(err)
		return
	}
}

func (b *bot) myUserId(s Session) string {
	if b.user_id != "" {
		return b.user_id
	}

	user, err := s.User("@me")
	if err != nil {
		logger.Warnf("unable to find my user id: %v", err)
		return ""
	}
	b.user_id = user.ID

	return user.ID
}

func (b *bot) presenceHandler(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	logger.Debugf("presenceHandler isn't implemented yet")
	return
}

func isPrivateMessage(s Session, m *discordgo.MessageCreate) bool {
	ch, err := s.Channel(m.ChannelID)
	if err != nil {
		return false
	}
	return ch.IsPrivate
}

func (b *bot) ReceiveRouter(router httpapi.Router) {
	router.HandleFunc("/", b.handleHTTP)
	router.HandleFunc("/oauth/redirect", b.oauthRedirect)
}
