package atcoderproblems

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"time"

	"atcoder_shojin/backend/internal/domain/training"
)

const defaultBaseURL = "https://kenkoooo.com/atcoder"

var abcContestID = regexp.MustCompile(`^abc[0-9]+$`)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

type contestJSON struct {
	ID               string `json:"id"`
	StartEpochSecond int64  `json:"start_epoch_second"`
	DurationSecond   int64  `json:"duration_second"`
	Title            string `json:"title"`
}
type problemJSON struct {
	ID        string `json:"id"`
	ContestID string `json:"contest_id"`
	Index     string `json:"problem_index"`
	Name      string `json:"name"`
}
type contestProblemJSON struct {
	ContestID string `json:"contest_id"`
	ProblemID string `json:"problem_id"`
	Index     string `json:"problem_index"`
}
type modelJSON struct {
	Difficulty *int `json:"difficulty"`
}
type submissionJSON struct {
	ID          int64  `json:"id"`
	EpochSecond int64  `json:"epoch_second"`
	ProblemID   string `json:"problem_id"`
	Result      string `json:"result"`
}

func (client *Client) FetchCatalog(ctx context.Context) ([]training.Contest, []training.Problem, error) {
	var rawContests []contestJSON
	if err := client.getJSON(ctx, "/resources/contests.json", &rawContests); err != nil {
		return nil, nil, err
	}
	var rawProblems []problemJSON
	if err := client.getJSON(ctx, "/resources/problems.json", &rawProblems); err != nil {
		return nil, nil, err
	}
	var pairs []contestProblemJSON
	if err := client.getJSON(ctx, "/resources/contest-problem.json", &pairs); err != nil {
		return nil, nil, err
	}
	var models map[string]modelJSON
	if err := client.getJSON(ctx, "/resources/problem-models.json", &models); err != nil {
		return nil, nil, err
	}
	counts := map[string]int{}
	for _, pair := range pairs {
		counts[pair.ContestID]++
	}
	contests := make([]training.Contest, 0, len(rawContests))
	abcContests := make(map[string]struct{})
	for _, contest := range rawContests {
		if !abcContestID.MatchString(contest.ID) {
			continue
		}
		abcContests[contest.ID] = struct{}{}
		contests = append(contests, training.Contest{ID: contest.ID, Title: contest.Title, StartTime: time.Unix(contest.StartEpochSecond, 0).UTC(), DurationSecond: contest.DurationSecond, ProblemCount: counts[contest.ID]})
	}
	metadata := make(map[string]problemJSON, len(rawProblems))
	for _, problem := range rawProblems {
		metadata[problem.ID] = problem
	}
	problems := make([]training.Problem, 0, len(pairs))
	for _, pair := range pairs {
		if _, exists := abcContests[pair.ContestID]; !exists || (pair.Index != "D" && pair.Index != "E" && pair.Index != "F") {
			continue
		}
		problem := metadata[pair.ProblemID]
		title := problem.Name
		if title == "" {
			title = pair.ProblemID
		}
		model := models[pair.ProblemID]
		problems = append(problems, training.Problem{ID: pair.ProblemID, ContestID: pair.ContestID, Index: pair.Index, Title: title, Difficulty: model.Difficulty})
	}
	return contests, problems, nil
}

func (client *Client) FetchSubmissions(ctx context.Context, userID string, fromSecond int64) ([]training.Submission, error) {
	all := make([]training.Submission, 0)
	for page := 0; page < 10000; page++ {
		path := fmt.Sprintf("/atcoder-api/v3/user/submissions?user=%s&from_second=%d", url.QueryEscape(userID), fromSecond)
		var raw []submissionJSON
		if err := client.getJSON(ctx, path, &raw); err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			break
		}
		sort.Slice(raw, func(i, j int) bool {
			if raw[i].EpochSecond == raw[j].EpochSecond {
				return raw[i].ID < raw[j].ID
			}
			return raw[i].EpochSecond < raw[j].EpochSecond
		})
		for _, item := range raw {
			all = append(all, training.Submission{ID: item.ID, EpochSecond: item.EpochSecond, ProblemID: item.ProblemID, Result: item.Result})
		}
		if len(raw) < 500 {
			break
		}
		next := raw[len(raw)-1].EpochSecond + 1
		if next <= fromSecond {
			return nil, fmt.Errorf("submission cursor did not advance")
		}
		fromSecond = next
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return all, nil
}

func (client *Client) getJSON(ctx context.Context, path string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "atcoder-shojin-v1")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("AtCoder Problems request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("AtCoder Problems returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode AtCoder Problems response: %w", err)
	}
	return nil
}
