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

import "sort"

// The caller should handle cases where len(all) is <= 2
func kk(all []int) (a, b []int) {
	logger.Warnf("Using the KK algorithm, which is not guaranteed " +
		"to generate teams with equal numbers of players.")

	sort.Ints(all)

	if len(all) == 2 {
		return []int{all[1]}, []int{all[0]}
	}

	nodes := []node{}
	for _, value := range all {
		nodes = append(nodes, node{label: value, value: value})
	}

	loops := 0
	var largest, next_largest *node
	for len(nodes) > 1 {
		By(label).Sort(nodes)
		logger.Debugf("loop: %v\n", nodes)
		largest, nodes = rpop(nodes)
		next_largest, nodes = rpop(nodes)
		largest.label -= next_largest.label
		largest.children = append(largest.children, *next_largest)
		nodes = append(nodes, *largest)
		loops++
		if loops > 2*len(all) {
			logger.Warnf("aborting due to loop count")
			break
		}
	}

	colorings := 0
	var colorTree func(node, bool)
	colorTree = func(root node, color bool) {
		colorings++
		if colorings > 2*len(all) {
			logger.Warnf("aborting due to coloring count")
			return
		}
		if color {
			a = append(a, root.value)
		} else {
			b = append(b, root.value)
		}
		for _, child := range root.children {
			colorTree(child, !color)
		}
	}
	colorTree(nodes[0], true)

	logger.Debugf("done: %+v", nodes)

	return a, b
}

type node struct {
	label    int
	value    int
	children []node
}

func rpop(list []node) (popped *node, left []node) {
	last_idx := len(list) - 1
	popped = &list[last_idx]
	if last_idx == 0 {
		left = []node{}
	} else {
		left = list[:last_idx]
	}
	return popped, left
}

type By func(left, right *node) bool

func (by By) Sort(nodes []node) {
	ts := &nodeSorter{
		nodes: nodes,
		by:    by,
	}
	sort.Sort(ts)
}

type nodeSorter struct {
	nodes []node
	by    func(left, right *node) bool
}

func (s *nodeSorter) Len() int {
	return len(s.nodes)
}

func (s *nodeSorter) Less(i, j int) bool {
	return s.by(&s.nodes[i], &s.nodes[j])
}

func (s *nodeSorter) Swap(i, j int) {
	s.nodes[i], s.nodes[j] = s.nodes[j], s.nodes[i]
}

func label(left, right *node) bool {
	if left.label < right.label {
		return true
	}
	return false
}
