package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"streamlane-go/internal/ranker"
)

type rankRequest struct {
	Tasks  []ranker.Task `json:"tasks"`
	Budget float64       `json:"budget"`
}

type rankResponse struct {
	Ranked []ranker.RankedTask `json:"ranked"`
	Plan   []ranker.RankedTask `json:"plan"`
	Used   float64             `json:"used"`
}

func main() {
	input := flag.String("input", "", "Path to JSON file with tasks")
	budget := flag.Float64("budget", 12.0, "Effort budget")
	serve := flag.Bool("serve", false, "Run HTTP API server instead of one-shot CLI")
	flag.Parse()

	if *serve {
		runServer(*budget)
		return
	}

	if *input == "" {
		fmt.Println("Provide -input sample/tasks.json or run with -serve")
		os.Exit(1)
	}

	request, err := parseInput(*input, *budget)
	if err != nil {
		fmt.Printf("invalid input: %v\n", err)
		os.Exit(1)
	}

	response := buildResponse(request)
	encoded, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("failed to encode response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}

func parseInput(filePath string, fallbackBudget float64) (rankRequest, error) {
	text, err := os.ReadFile(filePath)
	if err != nil {
		return rankRequest{}, err
	}

	request := rankRequest{}
	if err := json.Unmarshal(text, &request); err != nil {
		return rankRequest{}, err
	}

	if request.Budget <= 0 {
		request.Budget = fallbackBudget
	}
	return request, nil
}

func buildResponse(request rankRequest) rankResponse {
	ranked := ranker.RankTasks(request.Tasks)
	plan, used := ranker.BuildSprintPlan(request.Tasks, request.Budget)
	return rankResponse{
		Ranked: ranked,
		Plan:   plan,
		Used:   used,
	}
}

func runServer(defaultBudget float64) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux.HandleFunc("/rank", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		request := rankRequest{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid json"}`))
			return
		}
		if request.Budget <= 0 {
			request.Budget = defaultBudget
		}

		response := buildResponse(request)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	fmt.Println("streamlane-go listening on :8094")
	if err := http.ListenAndServe(":8094", mux); err != nil {
		fmt.Printf("server failed: %v\n", err)
		os.Exit(1)
	}
}

