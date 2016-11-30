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

package lootbox

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ewollesen/zenbot/overwatch"
	"github.com/ewollesen/zenbot/overwatch/blizzard"
	"github.com/ewollesen/zenbot/zentest"
)

func TestSkillRank(t *testing.T) {
	test, _ := newLootBoxTest(t)
	defer test.Close()

	test.AssertSR(overwatch.RegionUS, "testuser1#1111", 2000)
	test.AssertSR(overwatch.RegionUS, "foundus#1111", 4999)
	test.AssertSR(overwatch.RegionEU, "foundeu#2222", 4998)

	test.AssertUnranked(overwatch.RegionUS, "unranked#3333")

	test.AssertNotFound(overwatch.RegionUS, "foundeu#2222")
	test.AssertNotFound(overwatch.RegionEU, "foundus#1111")
}

type lootBoxTest struct {
	*zentest.ZenTest
	gow    *lootBox
	server *httptest.Server
}

func foundResponse(sr int) string {
	return fmt.Sprintf(`
{
  "data": {
    "competitive": {
      "rank": "%d"
    }
  }
}`, sr)
}

func foundResponseUnranked() string {
	return fmt.Sprintf(`
{
  "data": {
    "competitive": {
      "rank": null
    }
  }
}`)
}

func newLootBoxTest(t *testing.T) (*lootBoxTest, *lootBox) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("url: %s", req.URL.Path)
		switch {
		case strings.Contains(req.URL.Path, `us/testuser1-1111/`):
			fmt.Fprintf(w, foundResponse(2000))
		case strings.Contains(req.URL.Path, `us/foundus-1111/`):
			fmt.Fprintf(w, foundResponse(4999))
		case strings.Contains(req.URL.Path, `eu/foundeu-2222/`):
			fmt.Fprintf(w, foundResponse(4998))
		case strings.Contains(req.URL.Path, `/unranked-3333/`):
			t.Logf("returning unranked")
			fmt.Fprintf(w, foundResponseUnranked())
		default:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"error":"Found no user with the BattleTag: blah","statusCode":404}`)
		}
	}))
	gow := New(blizzard.New(), server.URL)

	return &lootBoxTest{
		ZenTest: zentest.New(t),
		gow:     gow,
		server:  server,
	}, gow
}

func (t *lootBoxTest) Close() {
	t.server.Close()
}

func (t *lootBoxTest) AssertSR(region, btag string, expected int) {
	sr, err := t.gow.SkillRank(overwatch.PlatformPC, region, btag)
	t.AssertEqual(sr, expected)
	t.AssertNil(err)
}

func (t *lootBoxTest) AssertNotFound(region, btag string) {
	sr, err := t.gow.SkillRank(overwatch.PlatformPC, region, btag)
	t.AssertErrorContainedBy(err, overwatch.BattleTagNotFound)
	t.AssertEqual(sr, overwatch.SkillRankError)
}

func (t *lootBoxTest) AssertUnranked(region, btag string) {
	sr, err := t.gow.SkillRank(overwatch.PlatformPC, region, btag)
	t.AssertErrorContainedBy(err, overwatch.BattleTagUnranked)
	t.AssertEqual(sr, overwatch.SkillRankError)
}
