package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const baseURL = "http://localhost:8080"
const baseIndex = 100000

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	PRID     string `json:"pull_request_id"`
	PRName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

var (
	teams       []string
	usersByTeam map[string][]string
	allUsers    []string
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🚀 Starting Vegeta Load Test...")
	fmt.Println("=" + repeat("=", 50))

	setupTestData()

	// RPS = 5
	rate := vegeta.Rate{Freq: 5, Per: time.Second}
	duration := 30 * time.Second
	targeter := createTargeter()

	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "PR Reviewer Service Load Test") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("Total Requests:     %d\n", metrics.Requests)
	fmt.Printf("Successful:         %d (%.2f%%)\n", metrics.StatusCodes["200"]+metrics.StatusCodes["201"], metrics.Success*100)
	fmt.Printf("Failed:             %d\n", metrics.Requests-uint64(metrics.StatusCodes["200"])-uint64(metrics.StatusCodes["201"]))
	fmt.Printf("RPS:                %.2f\n", metrics.Rate)
	fmt.Printf("Duration:           %s\n", metrics.Duration)
	fmt.Printf("Mean:             %s\n", metrics.Latencies.Mean)
	fmt.Printf("P50:              %s\n", metrics.Latencies.P50)
	fmt.Printf("P95:              %s\n", metrics.Latencies.P95)
	fmt.Printf("P99:              %s\n", metrics.Latencies.P99)
	fmt.Printf("Max:              %s\n", metrics.Latencies.Max)

	reporter := vegeta.NewTextReporter(&metrics)
	f, _ := os.Create("vegeta-report.txt")
	defer f.Close()
	reporter.Report(f)

	fmt.Println("report saved to vegeta-report.txt")
}

func setupTestData() {
	fmt.Println("Creating 20 teams with 200 users...")

	teams = make([]string, 20)
	usersByTeam = make(map[string][]string)
	allUsers = make([]string, 0, 4000)

	client := &http.Client{Timeout: 10 * time.Second}

	teamSizes := []int{
		200, 200, 200, 200, 200, 200, 200, 200, 200, 200,
		200, 200, 200, 200, 200, 200, 200, 200, 200, 200,
	}

	userCounter := baseIndex
	successCount := 0

	for i := 0; i < 20; i++ {
		teamName := fmt.Sprintf("team-%d", baseIndex+i)
		teams[i] = teamName

		teamSize := teamSizes[i]
		members := make([]TeamMember, teamSize)
		teamUsers := make([]string, teamSize)

		for j := 0; j < teamSize; j++ {
			userID := fmt.Sprintf("u-%d", userCounter)
			members[j] = TeamMember{
				UserID:   userID,
				Username: fmt.Sprintf("User%d", userCounter),
				IsActive: true,
			}
			teamUsers[j] = userID
			allUsers = append(allUsers, userID)
			userCounter++
		}

		usersByTeam[teamName] = teamUsers

		team := Team{
			TeamName: teamName,
			Members:  members,
		}

		body, _ := json.Marshal(team)
		req, _ := http.NewRequest("POST", baseURL+"/team/add", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode == 201 || resp.StatusCode == 200 {
				successCount++
			}
			resp.Body.Close()
		}
	}
}

func createTargeter() vegeta.Targeter {
	counter := baseIndex
	prCounter := baseIndex + 1000

	return func(tgt *vegeta.Target) error {

		roll := rand.Intn(100)

		switch {
		case roll < 40:
			randomTeam := teams[rand.Intn(len(teams))]
			teamUsers := usersByTeam[randomTeam]
			authorID := teamUsers[rand.Intn(len(teamUsers))]

			prID := fmt.Sprintf("pr-%d", prCounter)
			pr := PullRequest{
				PRID:     prID,
				PRName:   fmt.Sprintf("Feature %d", prCounter),
				AuthorID: authorID,
			}
			prCounter++

			body, _ := json.Marshal(pr)
			*tgt = vegeta.Target{
				Method: "POST",
				URL:    baseURL + "/pullRequest/create",
				Body:   body,
				Header: http.Header{"Content-Type": []string{"application/json"}},
			}

		case roll < 70:
			randomTeam := teams[rand.Intn(len(teams))]
			*tgt = vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s/team/get?team_name=%s", baseURL, randomTeam),
			}

		case roll < 90:
			randomUser := allUsers[rand.Intn(len(allUsers))]
			*tgt = vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s/users/getReview?user_id=%s", baseURL, randomUser),
			}

		default:
			*tgt = vegeta.Target{
				Method: "GET",
				URL:    baseURL + "/stats",
			}
		}

		counter++
		return nil
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
