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

package blizzard

import "regexp"

var (
	btagRe     = regexp.MustCompile("^\\pL[\\pL\\pN]{2,11}#\\d{1,7}$")
	btagTextRe = regexp.MustCompile("\\b\\pL[\\pL\\pN]{2,11}#\\d{1,7}\\b")
)

func FindBattleTags(text string) (btags []string) {
	for _, match := range btagTextRe.FindAllStringSubmatch(text, -1) {
		btags = append(btags, match[0])
	}
	return btags
}

func FirstBattleTag(text string) (btag string) {
	btags := FindBattleTags(text)
	if len(btags) == 0 {
		return ""
	}
	return btags[0]
}

func WellFormedBattleTag(btag string) bool {
	return btagRe.MatchString(btag)
}
