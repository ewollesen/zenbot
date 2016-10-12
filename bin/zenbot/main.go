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

package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	redis "gopkg.in/redis.v4"

	"github.com/ewollesen/zenbot/discord"
	"github.com/ewollesen/zenbot/httpapi"

	"github.com/spacemonkeygo/flagfile"
	"github.com/spacemonkeygo/spacelog"
	spacelog_setup "github.com/spacemonkeygo/spacelog/setup"
)

var (
	logger = spacelog.GetLoggerNamed("zenbot")

	redisAddr = flag.String("redis_addr", "localhost:6379",
		"address of the redis server")
	redisDB = flag.Int("redis_db", 0, "redis database to use")
)

func main() {
	flagfile.Load()
	spacelog_setup.MustSetup("zenbot")

	logger.Infof("starting up")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	var redis_client *redis.Client
	if redis_client = newRedisClient(); redis_client != nil {
		defer func() { logger.Errore(redis_client.Close()) }()
	}

	router := httpapi.New()
	discord_router := router.ForPath("/discord")

	bot := discord.New(redis_client)
	bot.ReceiveRouter(discord_router)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		logger.Errore(bot.Run(quit))
		wg.Done()
	}()
	logger.Errore(router.Serve())

	wg.Wait()
	logger.Infof("shutting down")
}

func newRedisClient() (redis_client *redis.Client) {
	redis_client = redis.NewClient(&redis.Options{
		Addr: *redisAddr, DB: *redisDB})
	if err := redis_client.Ping().Err(); err != nil {
		redis_client = nil
	}

	return redis_client
}
