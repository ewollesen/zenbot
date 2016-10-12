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

package zentest

import (
	"regexp"
	"runtime"
	"testing"

	"github.com/spacemonkeygo/errors"
)

type ZenTest struct {
	*testing.T
}

func New(t *testing.T) *ZenTest {
	return &ZenTest{
		T: t,
	}
}

func (t *ZenTest) Assert(pred bool) {
	if !pred {
		t.fail(true, "expected %+v to be true", pred)
	}
}

func (t *ZenTest) AssertEqual(actual, expected interface{}) {
	if actual != expected {
		t.fail(true, "got %+v, expected %+v", actual, expected)
	}
}

func (t *ZenTest) AssertNil(value interface{}) {
	if value != nil {
		t.fail(true, "expected %+v to be %+v", value, nil)
	}
}

func (t *ZenTest) AssertErrorContainedBy(err error,
	error_class *errors.ErrorClass) {

	if err == nil || !error_class.Contains(err) {
		t.fail(true, "expected %+v to be contained by %+v", err, error_class)
	}
}

func (t *ZenTest) AssertContains(haystack []string, needle string) {
	for _, straw := range haystack {
		if straw == needle {
			return
		}
	}

	t.fail(true, "expected %+v to be in %+v", needle, haystack)
}

func (t *ZenTest) AssertContainsRe(haystack []string, needle string) {
	re := regexp.MustCompile(needle)
	for _, straw := range haystack {
		if re.MatchString(straw) {
			return
		}
	}

	t.fail(true, "expected %+v to match in %+v", re, haystack)
}

func (t *ZenTest) fail(fatal bool, template string, args ...interface{}) {
	stack_buffer := make([]byte, 4096)
	stack_len := runtime.Stack(stack_buffer, false)
	t.Logf(template, args...)
	t.Logf("%s", string(stack_buffer[:stack_len]))
	if fatal {
		t.FailNow()
	} else {
		t.Fail()
	}
}
