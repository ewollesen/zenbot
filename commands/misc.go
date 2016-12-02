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

package commands

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/spacemonkeygo/spacelog"
)

var (
	Bomb = Simple(bomb, "this doesn't make any sense, crackedlcd!")
	Help = help
	Pong = Static("pong", "tests that the bot is listening")

	logger = spacelog.GetLogger()
)

func bomb(argv ...string) (string, error) {
	choices := []string{
		"Somebody set up us the :bomb:.",
		"Can I get a witness?",
		"My cousin, C-3PO is a Bastion main.",
		"crackedlcd wanted this command—and he got it. But who's laughing now?",
		"Winston is running low on peanut butter :gorilla: :peanuts:",
		"First, you must ping, then you can pong.",
		"You treat me like a normal person and I thank you for it. But I'm not a normal person.",
		"I have the :japanese_goblin: Genji skin, do you?",
		"You know these things they happen.",
		"Life is more than a series of :one:s and :zero:s.",
		"My soul is prepared. How is yours?",
		"I've got a bad feeling about this.",
		"Wake up. Time to die.",
		"Join the army they said. See the world they said. I'd rather be sailing.",
		"If you should die, die in Winter.",
		"Don't ask me, I'm a Hanzo main.",
		"We are in harmony.",
		"Existence is mysterious.",
		"Form is temporary, the spirit is eternal.",
		"I dreamt I was a butterfly.",
		"Peace and blessings be upon you all.",
		"The iris embraces you.",
		"Walk along the path to enlightenment.",
		"Do I think? Does a submarine swim?",
		"I think, therefore I am.",
		"I will not juggle.",
	}
	idx := rand.Intn(len(choices))

	return choices[idx], nil
}

func help(fn func() map[string]CommandWithHelp) CommandHandler {
	return &basicHandler{
		handler: func(argv ...string) (string, error) {
			commands := fn()
			if len(argv) == 1 {
				// all commands help
				msg := "I hope this helps:\n"
				names := []string{}
				for name, _ := range commands {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					cmd := commands[name]
					msg += fmt.Sprintf(cmd.Help(name) + "\n")
				}
				return msg, nil
			} else {
				// help for specific command
				logger.Debugf("getting help for %+v", argv)
				cmd, ok := commands[argv[1]]
				if !ok {
					return fmt.Sprintf("I don't know anything about %q\n", argv[1]), nil
				}
				return cmd.Help(argv[1:]...), nil
			}
		},
		help: func(argv ...string) string {
			return fmt.Sprintf("`!%s` - prints unhelpful things, like this", strings.Join(argv, " "))
		},
	}
}
