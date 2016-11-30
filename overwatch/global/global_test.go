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

package global

import (
	"testing"

	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/overwatch/mockoverwatch"
	"github.com/ewollesen/zenbot/zentest"
)

type globalTest struct {
	*zentest.ZenTest
	gow *GlobalOverwatch
}

func newGlobalTest(t *testing.T) (*globalTest, *GlobalOverwatch) {
	gow := New(mockoverwatch.New())
	return &globalTest{
		ZenTest: zentest.New(t),
		gow:     gow,
	}, gow
}

func TestSkillRank(t *testing.T) {
	test, _ := newGlobalTest(t)

	test.AssertSR("testuser1#1111", 2000)
	test.AssertSR("foundus#1111", 4999)
	test.AssertSR("foundeu#2222", 4998)
}

func TestSkillRankNotFound(t *testing.T) {
	test, _ := newGlobalTest(t)

	test.AssertNotFound("notfound#1234")
}

func (t *globalTest) AssertSR(btag string, expected int) {
	sr, err := t.gow.SkillRank(overwatch.PlatformPC, btag)
	t.AssertEqual(sr, expected)
	t.AssertNil(err)
}

func (t *globalTest) AssertNotFound(btag string) {
	sr, err := t.gow.SkillRank(overwatch.PlatformPC, btag)
	t.AssertErrorContainedBy(err, overwatch.BattleTagNotFound)
	t.AssertEqual(sr, -1)
}
