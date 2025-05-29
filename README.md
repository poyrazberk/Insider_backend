# Insider Backend – Football League Simulation (Go + Gin)

A lightweight backend that **simulates a 4-team Premier-League-style mini-league**, updates the table week-by-week, and estimates championship probabilities via Monte-Carlo after Week 4.

---

## Table of Contents
1. [Features](#-features)  
2. [Tech Stack & Architecture](#-tech-stack--architecture)  
3. [Setup & Run Locally](#-setup--run-locally)  
4. [Database Schema (SQL)](#-database-schema-sql)  
5. [API Endpoints](#-api-endpoints)  
6. [Postman Collection](#-postman-collection)  

---

## 1.Features
| Category | Description |
|----------|-------------|
| **Simulation** | Plays 6 weeks (double round-robin) based on team strengths. |
| **Interface-based design** | `TeamService`, `MatchService` interfaces + concrete services (`MyTeamService`, `MyMatchService`). |
| **Struct composition** | Services embed `*sql.DB` and depend on interfaces, not concrete types. |
| **Monte-Carlo champion odds** | 15 000 simulations of the remaining schedule; results rounded to one decimal. |
| **Result editing** | `/change-match-result` reverts old stats, applies new score, recalculates table + probabilities. |
| **Reset helpers** | `/reset-teams`, `/reset-matches` for a clean slate. |
| **Postman ready** | Full collection supplied for quick testing. |

---

## 2.Tech Stack & Architecture
| Layer | Tech / Pattern |
|-------|----------------|
| Language | Go 1.22 |
| Web framework | Gin |
| DB | MySQL 8 (can swap via the interfaces) |
| Design | Interface-oriented + struct composition |
| Simulation | Pure in-memory logic → zero DB I/O per Monte-Carlo run |


## 3.Setup & Run Locally  

### 3.1 Prerequisites  
| Tool | Version | Notes |
|------|---------|-------|
| **Go** | 1.22 or later | https://go.dev/dl  
| **MySQL** | 8.x | running locally on **localhost:3306**  
| **Git** | latest | for cloning the repo |

> **Optional:** Docker & Docker Compose if you prefer containerised setup.

---

### 3.2 Clone the repository  

```bash
git clone https://github.com/<poyrazberk>/Insider_backend.git
cd Insider_backend
go run .
```

## 4. Database Schema (SQL)

```sql
-- Drop old tables if they exist
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS teams;

-- Create teams table
CREATE TABLE teams (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    strength INT NOT NULL,
    
    points INT DEFAULT 0,
    goals_for INT DEFAULT 0,
    goals_against INT DEFAULT 0,
    goal_diff INT DEFAULT 0,
    wins INT DEFAULT 0,
    draws INT DEFAULT 0,
    losses INT DEFAULT 0
);

-- Create matches table
CREATE TABLE matches (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name_home VARCHAR(50) NOT NULL,
    name_away VARCHAR(50) NOT NULL,
    home_team_id INT NOT NULL,
    away_team_id INT NOT NULL,
    home_goals INT,
    away_goals INT,
    week INT,
    played BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

-- Insert sample teams
INSERT INTO teams (name, strength) VALUES
('Manchester United', 68),
('Liverpool', 98),
('Leicester City', 55),
('Manchester City', 84);

-- Insert matches
-- Week 1
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Manchester United', 'Liverpool', 1, 2, 1, false),
('Leicester City', 'Manchester City', 3, 4, 1, false);

-- Week 2
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Manchester United', 'Leicester City', 1, 3, 2, false),
('Liverpool', 'Manchester City', 2, 4, 2, false);

-- Week 3
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Manchester United', 'Manchester City', 1, 4, 3, false),
('Liverpool', 'Leicester City', 2, 3, 3, false);

-- Week 4
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Liverpool', 'Manchester United', 2, 1, 4, false),
('Manchester City', 'Leicester City', 4, 3, 4, false);

-- Week 5
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Leicester City', 'Manchester United', 3, 1, 5, false),
('Manchester City', 'Liverpool', 4, 2, 5, false);

-- Week 6
INSERT INTO matches (name_home, name_away, home_team_id, away_team_id, week, played) VALUES
('Manchester City', 'Manchester United', 4, 1, 6, false),
('Leicester City', 'Liverpool', 3, 2, 6, false);
```
## 5.API Endpoints 

### GET /teams
 Lists all teams and their current statistics (win/lose/draw counts, points, ids, and names)

### GET /matches
 Lists all matches including their results if played

### POST /play-week
 Plays the next unplayed week and returns updated standings and, if available, championship probabilities

### POST /play-all
 Plays all remaining weeks and returns results week-by-week


### POST /reset-teams
 Resets all team statistics (points, goals, wins, etc.)

### POST /reset-matches
 Resets all matches (clears goals and sets played to false)

### PUT /update-match
 Manually updates a specific match’s score
 Request Body:
{
  "match_id": 3,
  "home_goals": 2,
  "away_goals": 1
}


## 6.Postman Collection
🔗 [Click here to open in Postman](https://www.postman.com/supply-cosmologist-86813505/workspace/insider-backend-workspace/collection/36875182-53d01ea1-30f7-4869-bf5e-b6aeaa4b8821?action=share&creator=36875182)

```
This Postman Collection contains all the essential endpoints required to simulate and manage the football league backend.

### Included Endpoints:

| Method | Endpoint                | Description |
|--------|-------------------------|-------------|
| GET    | /teams                  | Fetch current team standings |
| GET    | /matches                | Get all match fixtures |
| POST   | /play-week              | Simulate the next unplayed week |
| POST   | /play-all               | Simulate the rest of the season |
| POST   | /reset-teams            | Reset all team statistics |
| POST   | /reset-matches          | Reset all match scores and flags |
| POST   | /change-match-result    | Manually override a match result |

```




