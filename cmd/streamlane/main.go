package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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

const (
	listenAddr   = ":8094"
	maxBodyBytes = 1 << 20 // 1 MiB is far above any realistic backlog payload
)

// newMux builds the API routes. Kept separate from the listener so handlers
// can be exercised with httptest.
func newMux(defaultBudget float64) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/rank", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeError(w, http.StatusMethodNotAllowed, "use POST")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		request := rankRequest{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body exceeds 1 MiB")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if request.Budget <= 0 {
			request.Budget = defaultBudget
		}

		writeJSON(w, http.StatusOK, buildResponse(request))
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// runServer uses an explicit http.Server because the package-level
// ListenAndServe has no timeouts, which leaves the API open to slow-header
// (Slowloris) connections that hold goroutines forever (gosec G112).
func runServer(defaultBudget float64) {
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           newMux(defaultBudget),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("streamlane-go listening on %s\n", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("server failed: %v\n", err)
		os.Exit(1)
	}
}
