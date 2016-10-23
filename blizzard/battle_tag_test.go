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

import (
	"testing"

	"github.com/ewollesen/zenbot/zentest"
)

func TestFindBattleTags(t *testing.T) {
	test := zentest.New(t)

	btags := FindBattleTags("this is an example#1234 battle tag")
	test.AssertContains(btags, "example#1234")
	test.AssertEqual(len(btags), 1)

	btags = FindBattleTags("example#1234 another#5678")
	test.AssertContains(btags, "example#1234")
	test.AssertContains(btags, "another#5678")

	btags = FindBattleTags("example#1234, another#5678")
	test.AssertContains(btags, "example#1234")
	test.AssertContains(btags, "another#5678")

	btags = FindBattleTags("example#1234 another#5678")
	test.AssertContains(btags, "example#1234")
	test.AssertContains(btags, "another#5678")
}

func TestFirstBattleTag(t *testing.T) {
	test := zentest.New(t)

	btag := FirstBattleTag("this is an example#1234 battle tag")
	test.AssertEqual(btag, "example#1234")

	btag = FirstBattleTag("there are no battle tags here")
	test.AssertEqual(btag, "")
}
