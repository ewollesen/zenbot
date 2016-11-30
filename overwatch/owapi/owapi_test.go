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

package owapi

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
	test, _ := newOwApiTest(t)
	defer test.Close()

	test.AssertSR("testuser1#1111", 2000)
	test.AssertSR("foundus#1111", 4999)
	test.AssertSR("foundeu#2222", 4998)

	test.AssertUnranked("unranked#3333")

	test.AssertNotFound("notfound#1234")
}

type owApiTest struct {
	*zentest.ZenTest
	gow    *owApi
	server *httptest.Server
}

func foundResponse(region string, sr int) string {
	return fmt.Sprintf(`
{
  "%s": {
    "stats": {
      "competitive": {
        "overall_stats": {
          "comprank": %d
        }
      }
    }
  }
}`, region, sr)
}

func foundResponseUnranked(region string) string {
	return fmt.Sprintf(`
{
  "%s": {
    "stats": {
      "competitive": {
        "overall_stats": {
          "comprank": null
        }
      }
    }
  }
}`, region)
}

func newOwApiTest(t *testing.T) (*owApiTest, *owApi) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.Contains(req.URL.Path, `/u/testuser1-1111`):
			fmt.Fprintf(w, foundResponse(overwatch.RegionUS, 2000))
		case strings.Contains(req.URL.Path, `/u/foundus-1111`):
			fmt.Fprintf(w, foundResponse(overwatch.RegionUS, 4999))
		case strings.Contains(req.URL.Path, `/u/foundeu-2222`):
			fmt.Fprintf(w, foundResponse(overwatch.RegionUS, 4998))
		case strings.Contains(req.URL.Path, `/u/unranked-3333`):
			fmt.Fprintf(w, foundResponseUnranked(overwatch.RegionUS))
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"msg":"profile not found","error":404}`)
		}
	}))
	gow := New(blizzard.New(), server.URL)

	return &owApiTest{
		ZenTest: zentest.New(t),
		gow:     gow,
		server:  server,
	}, gow
}

func (t *owApiTest) Close() {
	t.server.Close()
}

func (t *owApiTest) AssertSR(btag string, expected int) {
	sr, _, err := t.gow.SkillRank(overwatch.PlatformPC, btag)
	t.AssertEqual(sr, expected)
	t.AssertNil(err)
}

func (t *owApiTest) AssertNotFound(btag string) {
	sr, _, err := t.gow.SkillRank(overwatch.PlatformPC, btag)
	t.AssertErrorContainedBy(err, overwatch.BattleTagNotFound)
	t.AssertEqual(sr, -1)
}

func (t *owApiTest) AssertUnranked(btag string) {
	sr, _, err := t.gow.SkillRank(overwatch.PlatformPC, btag)
	t.AssertErrorContainedBy(err, overwatch.BattleTagUnranked)
	t.AssertEqual(sr, -1)
}
