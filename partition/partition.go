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
	"time"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/spacelog"
)

var (
	Error  = errors.NewClass("balance2")
	logger = spacelog.GetLogger()
)

func init() {
	rand.Seed(time.Now().Unix())
}

func Partition(ranks []int) (a, b []int) {
	if len(ranks) == 0 {
		return a, b
	}

	if len(ranks) == 1 {
		return ranks, b
	}

	if len(ranks) == 2 {
		return ranks[:1], ranks[1:]
	}

	if len(ranks)%2 != 0 {
		return kk(ranks)
	}

	return balancedBrute(ranks)
}
