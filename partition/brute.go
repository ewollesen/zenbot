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
	"math"
	"math/rand"
)

type team struct {
	// represents
	number      int
	ranks_total int
}

func (t *team) vs(other *team) int {
	return int(math.Abs(float64(t.ranks_total - other.ranks_total)))
}

type game struct {
	team_a_idx int
	team_b_idx int
	ranks_diff int
}

// The caller should handle cases where len(all) is <= 2
func balancedBrute(ranks []int) (a, b []int) {
	if len(ranks) > 12 {
		logger.Warnf("This is a brute force algorithm whose " +
			"performance declines sharply as the number of " +
			"inputs increases. USE WITH CAUTION.")
	}

	if len(ranks) > 32 {
		logger.Errorf("Too many ranks. Overflow would occur.")
		return a, b
	}

	games := []*game{}
	teams := []*team{}
	team_size := len(ranks) / 2
	max := twoPow32(len(ranks))

	for i := 0; i < max; i++ {
		// Optimization: replace with a table
		if !validTeam(i, team_size) {
			continue
		}

		sum := 0
		for j := 0; twoPow32(j) < i; j++ {
			if i&twoPow32(j) > 0 {
				sum += ranks[j]
			}
		}

		teams = append(teams, &team{
			number:      i,
			ranks_total: sum,
		})
	}

	num_expected_teams := numExpectedTeams(len(ranks), team_size)
	if len(teams) != num_expected_teams {
		logger.Warnf("expected %d teams, found %d",
			num_expected_teams, len(teams))
	}

	// Find pairs of teams with no overlaps in players
	var team_a, team_b *team
	for i := 0; i < len(teams)-1; i++ {
		team_a = teams[i]
		for j := 1; j < len(teams); j++ {
			team_b = teams[j]

			if team_a.number&team_b.number != 0 ||
				numOnes(team_a.number+team_b.number) != len(ranks) {
				continue
			}

			games = append(games, &game{
				team_a_idx: i,
				team_b_idx: j,
				ranks_diff: team_a.vs(team_b),
			})
		}
	}

	if len(games) == 0 {
		logger.Warnf("unable to find any games")
		return a, b
	}

	var best_game *game
	var best_games []*game
	min_diff := math.MaxInt32
	for _, g := range games {
		if g.ranks_diff < min_diff {
			best_games = []*game{g}
			min_diff = g.ranks_diff
		}
		if g.ranks_diff == min_diff {
			best_games = append(best_games, g)
		}
	}

	best_game = best_games[rand.Intn(len(best_games))]

	logger.Debugf("best match found: diff %d %+v",
		best_game.ranks_diff, best_game)
	logger.Debugf("teams a: %032b (%d)",
		best_game.team_a_idx, teams[best_game.team_a_idx].ranks_total)
	logger.Debugf("      b: %032b (%d)",
		best_game.team_b_idx, teams[best_game.team_b_idx].ranks_total)

	for i := 0; i < len(ranks); i++ {
		team_number := twoPow32(i)
		team_a := teams[best_game.team_a_idx]
		team_b := teams[best_game.team_b_idx]
		if team_a.number&team_number == team_number {
			a = append(a, ranks[i])
			logger.Debugf("team_a: %032b matches (%d) %032b",
				team_a.number, i, team_number)
		}
		if team_b.number&team_number == team_number {
			b = append(b, ranks[i])
			logger.Debugf("team_b: %032b matches (%d) %032b",
				team_b.number, i, team_number)
		}
	}

	return a, b
}

func validTeam(i, team_size int) bool { return numOnes(i) == team_size }

func numOnes(i int) int {
	sum := 0

	for j := 0; twoPow32(j) <= i; j++ {
		if twoPow32(j)&i > 0 {
			sum++
		}
	}
	return sum
}

func twoPow32(i int) int {
	return int(math.Pow(2, float64(i)))
}

func numExpectedTeams(num_players, team_size int) int {
	return fact(num_players) / (fact(team_size) * fact(num_players-team_size))
}

func fact(n int) int {
	if n <= 1 {
		return 1
	}
	return n * fact(n-1)
}
