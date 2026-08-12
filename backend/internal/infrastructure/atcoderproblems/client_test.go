package atcoderproblems

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCatalogJoinsDifficultyAndProblemCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/resources/contests.json":
			response.Write([]byte(`[{"id":"abc100","start_epoch_second":1000,"duration_second":6000,"title":"ABC 100"},{"id":"joisc2008","start_epoch_second":900,"duration_second":6000,"title":"JOI"}]`))
		case "/resources/problems.json":
			response.Write([]byte(`[{"id":"abc100_a","contest_id":"abc100","problem_index":"A","name":"A title"},{"id":"abc100_d","contest_id":"abc100","problem_index":"D","name":"D title"},{"id":"joisc2008_committee","contest_id":"joisc2008","problem_index":"committee","name":"Committee"}]`))
		case "/resources/contest-problem.json":
			response.Write([]byte(`[{"contest_id":"abc100","problem_id":"abc100_a","problem_index":"A"},{"contest_id":"abc100","problem_id":"abc100_d","problem_index":"D"},{"contest_id":"abc100","problem_id":"abc100_e","problem_index":"E"},{"contest_id":"joisc2008","problem_id":"joisc2008_committee","problem_index":"committee"}]`))
		case "/resources/problem-models.json":
			response.Write([]byte(`{"abc100_d":{"difficulty":1200}}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()
	contests, problems, err := NewClient(server.URL, server.Client()).FetchCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(contests) != 1 || contests[0].ProblemCount != 3 || len(problems) != 2 || problems[0].Difficulty == nil || *problems[0].Difficulty != 1200 || problems[1].Title != "abc100_e" {
		t.Fatalf("contests=%#v problems=%#v", contests, problems)
	}
}

func TestFetchSubmissionsMapsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("user") != "fken_prime_57" {
			t.Error("missing user")
		}
		response.Write([]byte(`[{"id":1,"epoch_second":123,"problem_id":"abc100_d","result":"AC"}]`))
	}))
	defer server.Close()
	items, err := NewClient(server.URL, server.Client()).FetchSubmissions(context.Background(), "fken_prime_57", 100)
	if err != nil || len(items) != 1 || items[0].Result != "AC" {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
