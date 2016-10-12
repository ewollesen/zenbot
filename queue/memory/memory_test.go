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

package memory

import (
	"testing"

	"github.com/ewollesen/zenbot/queue"
)

func TestClear(t *testing.T) {
	queue.CommonTestClear(t, New())
}

func TestDequeueN(t *testing.T) {
	queue.CommonTestDequeueN(t, New())
}

func TestEnqueue(t *testing.T) {
	queue.CommonTestEnqueue(t, New())
}

func TestIter(t *testing.T) {
	queue.CommonTestIter(t, New())
}

func TestPosition(t *testing.T) {
	queue.CommonTestPosition(t, New())
}

func TestRemove(t *testing.T) {
	queue.CommonTestRemove(t, New())
	queue.CommonTestRemoveFirst(t, New())
	queue.CommonTestRemoveMiddle(t, New())
	queue.CommonTestRemoveLast(t, New())
}

func TestSize(t *testing.T) {
	queue.CommonTestSize(t, New())
}
