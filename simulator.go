package main

import (
	"math/rand"
)

// Simulate teams' goals by looking at the expected goal values returned from simulateMatch function
func simulateGoals(expected float64) int {
    prob := rand.Float64()

    switch {
    case prob < 0.4: //Highest probability seperated for expected results case
        return int(expected)
    case prob < 0.65:
        return int(expected) + 1
    case prob < 0.8:
        return int(expected) + 2
	case prob < 0.9:
		return int(expected) + 3	
    default: //To ensure that unexpected results can also occur
        return int(expected) + rand.Intn(5)
    }
}

// Simulate a match between two teams by looking at their strengths
func simulateMatch(homeStrength, awayStrength int) (int, int) {
	total := float64(homeStrength + awayStrength)
	//Premier League statistics show that home teams average 1.6 goals, while away teams average 1.2
	//That's why I use coefficients 2.0 and 1.8 for calculations below to give the advantage to the Home Team
	expectedHome := float64(homeStrength) / total * 2.0 //Home Team has the advantage 
    expectedAway := float64(awayStrength) / total * 1.8 
    return simulateGoals(expectedHome), simulateGoals(expectedAway)
}