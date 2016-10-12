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

import "github.com/ewollesen/discordgo"

func isPermitted(s Session, m *discordgo.MessageCreate, perm int) (
	bool, error) {

	if isPrivateMessage(s, m) {
		logger.Debugf("no perms on private channels")
		return false, nil
	}

	perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		return false, err
	}
	return perms&perm > 0, nil
}
