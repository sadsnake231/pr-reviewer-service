package domain

import "time"

type User struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	TeamName string
	Members  []User
}

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type PRStats struct {
	TotalPRs         int64
	OpenPRs          int64
	MergedPRs        int64
	AuthoredPRsCount map[string]int64
	ReviewersCount   map[string]int64
}
