## 📄 Requirements Document: Poker Round Web App (Updated)

### 1. Goal

A simple web tool for tracking poker games within a local group, organized by seasons. Each player hosts one game per season. After each game, the host, winner, and second place are recorded.

---

### 2. Features

#### 2.1 Season Selection

- Dropdown at the top to select an existing season.

- New seasons are **added directly via the database**, no UI needed.

#### 2.2 Player Overview

- Shows all players in the selected season, split into two sections:

  - **Visited players** (i.e., games already hosted)

  - **Players still to visit**

- Display should be **copy/paste-friendly**, especially for messaging use.

- Sorting:

  - Visited → by date or name

  - Still to visit → alphabetically

#### 2.3 Add Game Entry

- Clicking on a “still to visit” player opens a form:

  - **Date** (default: today)

  - **Winner** (dropdown of all players)

  - **Second place** (dropdown of all players)

- On submit, the game is saved, and the player is moved to “visited”.

---

### 3. Tech Stack

#### Backend

- Language: **Go**

- Database: **MySQL** (or compatible)

- Hosted on **Fly.io**

#### Deployment

- Automated using a **GitHub Actions pipeline**

- Pipeline builds and deploys to Fly.io on push to `main`

#### Frontend

- Server-rendered HTML (e.g., `html/template` in Go)

- **Tailwind CSS** for styling

- Optional: Vanilla JavaScript for modal/form interaction
