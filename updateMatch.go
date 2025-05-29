package main

import (
    "database/sql"
    "errors"
)

func UpdateMatchResult(db *sql.DB, teamService TeamService, matchService *MyMatchService, matchID, homeGoals, awayGoals int) ([]Team, interface{}, error) {
    // Get the match info
    var homeTeamID, awayTeamID, oldHomeGoals, oldAwayGoals, week int
    var played bool
    err := db.QueryRow(
        "SELECT home_team_id, away_team_id, IFNULL(home_goals,0), IFNULL(away_goals,0), played, week FROM matches WHERE id = ?",
        matchID,
    ).Scan(&homeTeamID, &awayTeamID, &oldHomeGoals, &oldAwayGoals, &played, &week)
    if err != nil {
        return nil, nil, errors.New("match not found")
    }

    // If match was played, revert old stats
    if played {
        revertTeamStats := func(teamID, goalsFor, goalsAgainst int, isHome bool) error {
            var points, wins, draws, losses int
            err := db.QueryRow("SELECT points, wins, draws, losses FROM teams WHERE id = ?", teamID).
                Scan(&points, &wins, &draws, &losses)
            if err != nil {
                return err
            }
            if oldHomeGoals > oldAwayGoals && isHome {
                points -= 3
                wins -= 1
            } else if oldHomeGoals < oldAwayGoals && !isHome {
                points -= 3
                wins -= 1
            } else if oldHomeGoals == oldAwayGoals {
                points -= 1
                draws -= 1
            } else {
                losses -= 1
            }
            _, err = db.Exec(`
                UPDATE teams SET
                    points = ?,
                    goals_for = goals_for - ?,
                    goals_against = goals_against - ?,
                    goal_diff = goal_diff - (? - ?),
                    wins = ?,
                    draws = ?,
                    losses = ?
                WHERE id = ?`,
                points, goalsFor, goalsAgainst, goalsFor, goalsAgainst, wins, draws, losses, teamID)
            return err
        }
        _ = revertTeamStats(homeTeamID, oldHomeGoals, oldAwayGoals, true)
        _ = revertTeamStats(awayTeamID, oldAwayGoals, oldHomeGoals, false)
    }

    // Update match result
    _, err = db.Exec(
        "UPDATE matches SET home_goals = ?, away_goals = ?, played = true WHERE id = ?",
        homeGoals, awayGoals, matchID)
    if err != nil {
        return nil, nil, errors.New("failed to update match")
    }

    // Update teams with new result
    updateTeamStats := func(teamID, goalsFor, goalsAgainst int, isHome bool) error {
        var points, wins, draws, losses int
        err := db.QueryRow("SELECT points, wins, draws, losses FROM teams WHERE id = ?", teamID).
            Scan(&points, &wins, &draws, &losses)
        if err != nil {
            return err
        }
        if homeGoals > awayGoals && isHome {
            points += 3
            wins += 1
        } else if homeGoals < awayGoals && !isHome {
            points += 3
            wins += 1
        } else if homeGoals == awayGoals {
            points += 1
            draws += 1
        } else {
            losses += 1
        }
        _, err = db.Exec(`
            UPDATE teams SET
                points = ?,
                goals_for = goals_for + ?,
                goals_against = goals_against + ?,
                goal_diff = goal_diff + (? - ?),
                wins = ?,
                draws = ?,
                losses = ?
            WHERE id = ?`,
            points, goalsFor, goalsAgainst, goalsFor, goalsAgainst, wins, draws, losses, teamID)
        return err
    }
    _ = updateTeamStats(homeTeamID, homeGoals, awayGoals, true)
    _ = updateTeamStats(awayTeamID, awayGoals, homeGoals, false)

    // Get updated standings
    teams, err := teamService.GetTeams()
    if err != nil {
        return nil, nil, errors.New("failed to get teams")
    }

    // Get updated probabilities for the current week)
    probabilities, _ := matchService.probabilities_Message(teamService.(*MyTeamService), matchService, week)

    return teams, probabilities, nil
}