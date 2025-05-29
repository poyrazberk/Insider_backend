package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// PlayAllHandler handles the request to play all matches in the season
func PlayAllHandler(matchService MatchService, teamService TeamService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var results []WeeklyResult
        weekProbabilities := make(map[int]interface{})

		// play all matches in the season
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
            probs, _ := matchService.probabilities_Message(teamService.(*MyTeamService), matchService.(*MyMatchService), week)
            weekProbabilities[week] = probs
        }

		// Respond with the results of all matches played
        c.JSON(http.StatusOK, gin.H{
            "message": "All matches played successfully",
            "weeks":   results,
            "championship_probabilities": weekProbabilities,
        })
    }
}