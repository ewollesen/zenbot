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

package partition

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/ewollesen/zenbot/zentest"
)

func TestBrute(t *testing.T) {
	test := newBalanceTest(t)

	test.Assert(true)
}

func TestGoof(t *testing.T) {
	test := newBalanceTest(t)

	Partition([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	test.Assert(true)
}

func TestNumOnes(t *testing.T) {
	test := newBalanceTest(t)

	test.AssertEqual(numOnes(0), 0)
	test.AssertEqual(numOnes(1), 1)
	test.AssertEqual(numOnes(2), 1)
	test.AssertEqual(numOnes(3), 2)
	test.AssertEqual(numOnes(4), 1)
	test.AssertEqual(numOnes(5), 2)
	test.AssertEqual(numOnes(6), 2)
	test.AssertEqual(numOnes(7), 3)
	test.AssertEqual(numOnes(8), 1)
	test.AssertEqual(numOnes(83), 4)
	test.AssertEqual(numOnes(127), 7)
	test.AssertEqual(numOnes(210), 4)
	test.AssertEqual(numOnes(255), 8)
}

func TestEmpty(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{})
	test.AssertEqualContents(a, []int{})
	test.AssertEqualContents(b, []int{})
}

func TestOne(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{1})
	test.AssertEqualContents(a, []int{1})
	test.AssertEqualContents(b, []int{})
}

func TestTwo(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{2, 1})
	test.AssertEqualContents(a, []int{2})
	test.AssertEqualContents(b, []int{1})

	c, d := Partition([]int{1, 2})
	test.AssertEqualContents(c, []int{1})
	test.AssertEqualContents(d, []int{2})
}

func TestThree(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{2, 1, 3})
	test.AssertEqualContents(b, []int{1, 2})
	test.AssertEqualContents(a, []int{3})
}

func TestExample(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{8, 7, 6, 5, 4})
	test.AssertEqualContents(a, []int{4, 5, 7})
	test.AssertEqualContents(b, []int{6, 8})
}

func TestExample2(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{3, 6, 13, 20, 30, 40, 73})
	test.AssertEqualContents(a, []int{3, 20, 30, 40})
	test.AssertEqualContents(b, []int{6, 13, 73})
}

func TestTwelve(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	test.AssertEqualContents(a, []int{2, 4, 5, 6, 10, 12})
	test.AssertEqualContents(b, []int{1, 3, 7, 8, 9, 11})
}

func TestMultiSet(t *testing.T) {
	test := newBalanceTest(t)

	a, b := Partition([]int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6})
	test.AssertEqualContents(a, []int{1, 2, 3, 4, 5, 6})
	test.AssertEqualContents(b, []int{1, 2, 3, 4, 5, 6})

	c, d := Partition([]int{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4})
	test.AssertEqualContents(c, []int{1, 2, 2, 3, 3, 4})
	test.AssertEqualContents(d, []int{1, 1, 2, 3, 4, 4})
}

func TestSixteen(t *testing.T) {
	test := newBalanceTest(t)
	all := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	a, b := Partition(all)
	test.AssertEqualContents(a, []int{1, 4, 6, 7, 10, 12, 13, 15})
	test.AssertEqualContents(b, []int{2, 3, 5, 8, 9, 11, 14, 16})
}

func TestSkillRanks(t *testing.T) {
	test := newBalanceTest(t)

	// These are 12 people's real skill ranks, grabbed from an existing
	// queue, minus one that wasn't properly case-sensitive (I subbed my
	// own) I have not hand-verified this result, but am using it as a sort
	// of regression check.
	a, b := Partition([]int{1836, 1892, 1901, 2176, 2558, 2915,
		2935, 2968, 3420, 3458, 3723, 3963})
	test.AssertEqualContents(a, []int{2915, 2558, 3963, 3420, 2176, 1836})
	test.AssertEqualContents(b, []int{1901, 3458, 3723, 1892, 2968, 2935})
}

type balanceTest struct {
	*zentest.ZenTest
}

func newBalanceTest(t *testing.T) *balanceTest {
	rand.Seed(42)
	return &balanceTest{
		zentest.New(t),
	}
}

func (t *balanceTest) AssertEqualContents(got, expected []int) {
	t.Logf("equal contents of: %v and %v\n", got, expected)
	t.AssertEqual(len(got), len(expected))

	sort.Ints(got)
	sort.Ints(expected)

	for i, _ := range got {
		t.AssertEqual(got[i], expected[i])
	}
}

func (t *balanceTest) AssertEqualContentsStrings(got, expected []string) {
	t.AssertEqual(len(got), len(expected))

	sort.Strings(got)
	sort.Strings(expected)

	for i, _ := range got {
		t.AssertEqual(got[i], expected[i])
	}
}
