# Insider_backend

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
7. [Code Structure](#-code-structure)  
8. [How Probabilities Are Calculated](#-how-probabilities-are-calculated)  
9. [Extras / Future Work](#-extras--future-work)

---

## Features
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

## Tech Stack & Architecture
| Layer | Tech / Pattern |
|-------|----------------|
| Language | Go 1.22 |
| Web framework | Gin |
| DB | MySQL 8 (can swap via the interfaces) |
| Design | Interface-oriented + struct composition |
| Simulation | Pure in-memory logic → zero DB I/O per Monte-Carlo run |


## Setup & Run Locally  

### 1. Prerequisites  
| Tool | Version | Notes |
|------|---------|-------|
| **Go** | 1.22 or later | https://go.dev/dl  
| **MySQL** | 8.x | running locally on **localhost:3306**  
| **Git** | latest | for cloning the repo |

> **Optional:** Docker & Docker Compose if you prefer containerised setup.

---

### 2. Clone the repository  

```bash
git clone https://github.com/<poyrazberk>/Insider_backend.git
cd Insider_backend
go run .
```

## Database Schema (SQL)

DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS teams;

CREATE TABLE teams (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(50) NOT NULL,
  strength      INT NOT NULL,
  points        INT DEFAULT 0,
  goals_for     INT DEFAULT 0,
  goals_against INT DEFAULT 0,
  goal_diff     INT DEFAULT 0,
  wins          INT DEFAULT 0,
  draws         INT DEFAULT 0,
  losses        INT DEFAULT 0
);

CREATE TABLE matches (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  name_home     VARCHAR(50) NOT NULL,
  name_away     VARCHAR(50) NOT NULL,
  home_team_id  INT NOT NULL,
  away_team_id  INT NOT NULL,
  home_goals    INT,
  away_goals    INT,
  week          INT NOT NULL,
  played        BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (home_team_id) REFERENCES teams(id),
  FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

INSERT INTO teams (name, strength) VALUES
('Manchester United', 68),
('Liverpool',         98),
('Leicester City',    55),
('Manchester City',   84);



