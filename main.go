package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

// --- Structs for domain models ---

// Team represents a football team with its attributes
type Team struct {
    ID           int    `json:"id"`
    Name         string `json:"name"`
    Strength     int    `json:"strength"`
    Points       int    `json:"points"`
    GoalsFor     int    `json:"goals_for"`
    GoalsAgainst int    `json:"goals_against"`
    GoalDiff     int    `json:"goal_diff"`
    Wins         int    `json:"wins"`
    Draws        int    `json:"draws"`
    Losses       int    `json:"losses"`
}

// Match struct with its attributes
type Match struct {
    ID         int    `json:"id"`
    NameHome   string `json:"name_home"`
    NameAway   string `json:"name_away"`
    HomeTeamID int    `json:"home_team_id"`
    AwayTeamID int    `json:"away_team_id"`
    HomeGoals  *int   `json:"home_goals"`
    AwayGoals  *int   `json:"away_goals"`
	Week 	   int    `json:"week"`
    Played     bool   `json:"played"`
}

// WeeklyResult struct used to return weekly results in the /play-all endpoint
type WeeklyResult struct {
    Week      int    `json:"week"`
    Standings []Team `json:"standings"`
}

// --- Interfaces ---

// TeamService interface defines methods for managing teams
type TeamService interface {
    GetTeams() ([]Team, error)
    ResetTeams() error
}

// MatchService interface defines methods for managing matches
type MatchService interface {
    GetMatches() ([]Match, error)
    PlayWeek() (int, []Team, error)
    ResetMatches() error
	probabilities_Message(teamService *MyTeamService, matchService *MyMatchService, week int) (interface{}, error)
}

// --- Structs implementing interfaces ---

// myTeamService implements TeamService interface
type MyTeamService struct {
    db *sql.DB
}

// myMatchService implements MatchService interface
type MyMatchService struct {
    db          *sql.DB
    teamService TeamService
}

// --- TeamService methods ---

// GetTeams retrieves all teams from the database
func (s *MyTeamService) GetTeams() ([]Team, error) {
    rows, err := s.db.Query(`SELECT id, name, strength, points, goals_for, goals_against, goal_diff, wins, draws, losses 
						    FROM teams`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var teams []Team
    for rows.Next() {
        var t Team
        if err := rows.Scan(&t.ID, &t.Name, &t.Strength, &t.Points, &t.GoalsFor, &t.GoalsAgainst, &t.GoalDiff, &t.Wins, &t.Draws, &t.Losses); err != nil {
            return nil, err
        }
        teams = append(teams, t)
    }
    return teams, nil
}

// ResetTeams resets all teams to their initial state
func (s *MyTeamService) ResetTeams() error {
    _, err := s.db.Exec(`UPDATE teams 
						SET points = 0,
							goals_for = 0,
							goals_against = 0,
							goal_diff = 0,
							wins = 0,
							draws = 0,
							losses = 0
    `)
    return err
}


// --- MatchService methods ---

// GetMatches retrieves all matches from the database
func (s *MyMatchService) GetMatches() ([]Match, error) {
    rows, err := s.db.Query(`SELECT id, name_home, name_away, home_team_id, away_team_id, home_goals, away_goals, week, played 
							FROM matches`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var matches []Match 
    for rows.Next() {
        var m Match
        var homeGoals, awayGoals sql.NullInt64
        if err := rows.Scan(&m.ID, &m.NameHome, &m.NameAway, &m.HomeTeamID, &m.AwayTeamID, &homeGoals, &awayGoals, &m.Week, &m.Played); err != nil {
            return nil, err
        }
        if homeGoals.Valid {
            val := int(homeGoals.Int64)
            m.HomeGoals = &val
        }
        if awayGoals.Valid {
            val := int(awayGoals.Int64)
            m.AwayGoals = &val
        }
        matches = append(matches, m)
    }
    return matches, nil
}

// This function simulates a week of matches, updates the scores, and returns the standings
func (s *MyMatchService) PlayWeek() (int, []Team, error) {
    var lastPlayedWeek sql.NullInt64
    err := s.db.QueryRow("SELECT MAX(week) FROM matches WHERE played = true").Scan(&lastPlayedWeek)
    if err != nil {
        return 0, nil, err
    }

	// Determine the ne,xt week to play
    nextWeek := 1
    if lastPlayedWeek.Valid {
        if lastPlayedWeek.Int64 < 6 {
            nextWeek = int(lastPlayedWeek.Int64) + 1
        } else {
            return 0, nil, fmt.Errorf("Season has ended")
        }
    }


    rows, err := s.db.Query("SELECT id, home_team_id, away_team_id FROM matches WHERE week = ? AND played = false", nextWeek)
    if err != nil {
        return 0, nil, err
    }
    defer rows.Close()

	// Fr each match simulate the result and update the database
    for rows.Next() {
        var id, homeID, awayID int
        if err := rows.Scan(&id, &homeID, &awayID); err != nil {
            return 0, nil, err
        }
        var homeStrength, awayStrength int
        err := s.db.QueryRow("SELECT strength FROM teams WHERE id = ?", homeID).Scan(&homeStrength)
        if err != nil {
            return 0, nil, err
        }
        err = s.db.QueryRow("SELECT strength FROM teams WHERE id = ?", awayID).Scan(&awayStrength)
        if err != nil {
            return 0, nil, err
        }

		// Simulate the match result
        home_goals, away_goals := simulateMatch(homeStrength, awayStrength)

        // Update the match result in the database
        _, err = s.db.Exec("UPDATE matches SET home_goals = ?, away_goals = ?, played = true WHERE id = ?", home_goals, away_goals, id)
        if err != nil {
            return 0, nil, err
        }
        _, err = s.db.Exec("UPDATE teams SET goals_for = goals_for + ?, goals_against = goals_against + ?, goal_diff = goal_diff + ? WHERE id = ?", home_goals, away_goals, home_goals-away_goals, homeID)
        if err != nil {
            return 0, nil, err
        }
        _, err = s.db.Exec("UPDATE teams SET goals_for = goals_for + ?, goals_against = goals_against + ?, goal_diff = goal_diff + ? WHERE id = ?", away_goals, home_goals, away_goals-home_goals, awayID)
        if err != nil {
            return 0, nil, err
        }

		// UPDATE the teams' points and stats based on the match result

		// case1: Home team wins
        if home_goals > away_goals { 
            _, err = s.db.Exec("UPDATE teams SET points = points + 3, wins = wins + 1 WHERE id = ?", homeID)
            if err != nil {
                return 0, nil, err
            }
            _, err = s.db.Exec("UPDATE teams SET losses = losses + 1 WHERE id = ?", awayID)
            if err != nil {
                return 0, nil, err
            }
        }
		// case2: Away team wins
        if home_goals < away_goals { 
            _, err = s.db.Exec("UPDATE teams SET points = points + 3, wins = wins + 1 WHERE id = ?", awayID)
            if err != nil {
                return 0, nil, err
            }
            _, err = s.db.Exec("UPDATE teams SET losses = losses + 1 WHERE id = ?", homeID)
            if err != nil {
                return 0, nil, err
            }
        }
		// case3: Draw, both teams get 1 point
        if home_goals == away_goals { 
            _, err = s.db.Exec("UPDATE teams SET points = points + 1, draws = draws + 1 WHERE id = ?", awayID)
            if err != nil {
                return 0, nil, err
            }
            _, err = s.db.Exec("UPDATE teams SET points = points + 1, draws = draws + 1 WHERE id = ?", homeID)
            if err != nil {
                return 0, nil, err
            }
        }
    }

    // Get updated standings
    rows, err = s.db.Query(`SELECT id, name, strength, points, goals_for, goals_against, goal_diff, wins, draws, losses 
						   FROM teams
						   ORDER BY points DESC, goal_diff DESC, goals_for DESC`)
    if err != nil {
        return 0, nil, err
    }
    defer rows.Close()

	// Collect the standings after the week has been played
    var teams []Team
    for rows.Next() {
        var t Team
        if err := rows.Scan(&t.ID, &t.Name, &t.Strength, &t.Points, &t.GoalsFor, &t.GoalsAgainst, &t.GoalDiff, &t.Wins, &t.Draws, &t.Losses); err != nil {
            return 0, nil, err
        }
        teams = append(teams, t)
    }

	// Return the next week number and the updated standings
    return nextWeek, teams, nil
}


// Reset all matches to their initial state
func (s *MyMatchService) ResetMatches() error {
    _, err := s.db.Exec(`
        UPDATE matches 
        SET home_goals = NULL,
            away_goals = NULL,
            played = FALSE
    `)
    return err
}

// probabilities_Message prepares a message with championship probabilities based on the current week
func (s *MyMatchService) probabilities_Message(teamService *MyTeamService, matchService *MyMatchService, week int) (interface{}, error) {
	if week <= 3 {
		return "Not enough weeks played to calculate championship probabilities", nil
	}
	probabilities, err := SimulateChampionshipProbabilities(teamService, matchService, week)
	if err != nil {
		return "Could not calculate probabilities: " + err.Error(), nil
	}
	return probabilities, nil
}


// --- Main and Handlers ---

// main function initializes the database connection and sets up the HTTP server
func main() {
	// Initialize the database connection
    db, err := sql.Open("mysql", "root:berkemre123@tcp(127.0.0.1:3306)/leaguedb")
    if err != nil {
        log.Fatal("DB bağlantısı başarısız:", err)
    }
    err = db.Ping()
    if err != nil {
        log.Fatal("DB erişimi başarısız:", err)
    }

	// Initialize services
    teamService := &MyTeamService{db: db}
    matchService := &MyMatchService{db: db, teamService: teamService}

	// Initialize Gin router
    r := gin.Default()

	// Endpoint to get all teams
    r.GET("/teams", func(c *gin.Context) {
        teams, err := teamService.GetTeams()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, teams)
    })

	// Endpoint to get all matches
    r.GET("/matches", func(c *gin.Context) {
        matches, err := matchService.GetMatches()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, matches)
    })

	// Endpoint to play a single week of matches
    r.POST("/play-week", func(c *gin.Context) {
        week, teams, err := matchService.PlayWeek()
        if err != nil {
            c.JSON(http.StatusOK, gin.H{"error": err.Error()})
            return
        }

		probabilities, err := matchService.probabilities_Message(teamService, matchService, week)

        c.JSON(http.StatusOK, gin.H{
            "message":   fmt.Sprintf("Week %d played successfully", week),
            "standings": teams,
			"championship_probabilities": probabilities,
        })
    })

	// Endpoint to play all weeks until the season ends
    r.POST("/play-all", func(c *gin.Context) {
		var results []WeeklyResult
		weekProbabilities := make(map[int]interface{})

		for {
			week, standings, err := matchService.PlayWeek()
			if err != nil {
				if err.Error() == "Season has ended" {
					break
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Collect standings for this week
			results = append(results, WeeklyResult{
				Week:      week,
				Standings: standings,
			})

			// Collect probabilities for this week (after week 3)
			probs, _ := matchService.probabilities_Message(teamService, matchService, week)
			weekProbabilities[week] = probs
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "All matches played successfully",
			"weeks":   results,
			"championship_probabilities": weekProbabilities,
		})
	})

	// Endpoint to change match result. Then update standings and championship probabilities for that week accordingly.
	r.POST("/change-match-result", func(c *gin.Context) {
		type ChangeMatchRequest struct {
			MatchID    int `json:"match_id"`
			HomeGoals  int `json:"home_goals"`
			AwayGoals  int `json:"away_goals"`
		}
		var req ChangeMatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		teams, probabilities, err := UpdateMatchResult(matchService.db, teamService, matchService, req.MatchID, req.HomeGoals, req.AwayGoals)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Match result updated successfully",
			"standings": teams,
			"championship_probabilities": probabilities,
		})
	})


	// Endpoint to reset all teams
    r.POST("/reset-teams", func(c *gin.Context) {
        if err := teamService.ResetTeams(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset teams: " + err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Teams reset successfully"})
    })

	// Endpoint to reset all matches
    r.POST("/reset-matches", func(c *gin.Context) {
        if err := matchService.ResetMatches(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset matches: " + err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Matches reset successfully"})
    })

    r.Run(":8080")
}
