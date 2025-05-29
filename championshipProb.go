package main

import (
	"math"
)

// SimulateChampionshipProbabilities simulates the championship probabilities for each team
func SimulateChampionshipProbabilities(teamService TeamService, matchService MatchService, currentWeek int) (map[int]float64, error) {
	const numSimulations = 15000
	counts := make(map[int]int)

	// Get real teams and matches from the database
	initialTeams, err := teamService.GetTeams()
	if err != nil {
		return nil, err
	}
	initialMatches, err := matchService.GetMatches()
	if err != nil {
		return nil, err
	}

	// Monte Carlo simulation to estimate championship probabilities
	for sim := 0; sim < numSimulations; sim++ {
		// Clone initial teams and matches to avoid modifying the original data
		teams := cloneTeams(initialTeams)
		matches := cloneMatches(initialMatches)

		// Play remaining weeks
		for week := currentWeek + 1; week <= 6; week++ {
			playWeekSimulation(week, teams, matches)
		}

		// Find the leader of the championship
		leaderID := findLeader(teams)
		counts[leaderID]++
	}

	// Calculate probabilities based on counts
	probabilities := map[int]float64{
		// Initialize probabilities map
		1: 0.0,
		2: 0.0,
		3: 0.0,
		4: 0.0,
	}
	for teamID, count := range counts {
		realValue := (float64(count) / float64(numSimulations)) * 100.0 // Convert to percentage without decreasing precision
		probabilities[teamID] = math.Round(realValue*1000) / 1000.0 // Round to three decimal place
	}
	return probabilities, nil
}

// Clone teams from db to avoid modifying the original data
func cloneTeams(original []Team) []Team {
	cloned := make([]Team, len(original))
	for i, t := range original {
		cloned[i] = Team{
			ID:      		t.ID,
			Name:    		t.Name,
			Strength: 		t.Strength,
			GoalsFor: 		t.GoalsFor,
			GoalsAgainst: 	t.GoalsAgainst,
			GoalDiff: 		t.GoalDiff,
			Wins:   		t.Wins,
			Draws:  		t.Draws,
			Losses: 		t.Losses,
			Points:  		t.Points,
		}
	}
	return cloned
}

// Clone matches from db to avoid modifying the original data
func cloneMatches(original []Match) []Match {
	cloned := make([]Match, len(original))
	for i, m := range original {
		cloned[i] = Match{
			ID:          m.ID,
			NameHome:    m.NameHome,
			NameAway:    m.NameAway,
			HomeTeamID:  m.HomeTeamID,
			AwayTeamID:  m.AwayTeamID,
			HomeGoals:   m.HomeGoals,
			AwayGoals:   m.AwayGoals,
			Week:        m.Week,
			Played:      m.Played,
		}
	}
	return cloned
}

// This function simulates a week of matches, updates the scores, and returns the standings
func playWeekSimulation(week int, teams []Team, matches []Match) () {
	for i := range matches {
		match := &matches[i]
		if match.Week == week && !match.Played {
			homeTeam := findTeamByID(teams, match.HomeTeamID)
			awayTeam := findTeamByID(teams, match.AwayTeamID)

			if homeTeam != nil && awayTeam != nil {
				homeGoals, awayGoals := simulateMatch(homeTeam.Strength, awayTeam.Strength)

				// Update match results
				match.HomeGoals = &homeGoals
				match.AwayGoals = &awayGoals
				match.Played = true

				// Update teams' stats
				updateTeamStats(homeTeam, homeGoals, awayGoals)
				updateTeamStats(awayTeam, awayGoals, homeGoals)
			}
		}
	}
}

// Find a team by its ID in the list of teams
func findTeamByID(teams []Team, id int) *Team {
	for i := range teams {
		if teams[i].ID == id {
			return &teams[i]
		}
	}
	return nil
} 

// Update the team statistics based on the match results
func updateTeamStats(team *Team, goalsFor int, goalsAgainst int) {
	team.GoalsFor += goalsFor
	team.GoalsAgainst += goalsAgainst
	team.GoalDiff = team.GoalsFor - team.GoalsAgainst

	if goalsFor > goalsAgainst {
		team.Wins++
		team.Points += 3 // 3 points for a win
	} else if goalsFor < goalsAgainst {
		team.Losses++
	} else {
		team.Draws++
		team.Points++ // 1 point for a draw
	}
}

// Find the leader of the championship based on points
func findLeader(teams []Team) int {
    leaderID := -1
    maxPoints := -1
    maxGoalDiff := -1
    maxGoalsFor := -1

    for _, team := range teams {
        if team.Points > maxPoints || // Check for higher points
            (team.Points == maxPoints && team.GoalDiff > maxGoalDiff) || // Check for higher goal difference if points are equal
            (team.Points == maxPoints && team.GoalDiff == maxGoalDiff && team.GoalsFor > maxGoalsFor) { // Check for higher goals for if points and goal difference are equal
            maxPoints = team.Points
            maxGoalDiff = team.GoalDiff
            maxGoalsFor = team.GoalsFor
            leaderID = team.ID
        }
    }
    return leaderID
}