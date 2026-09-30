# Streamlane Go

Go-based sprint prioritization service for turning raw task signals into ranked execution plans.

## Features

- Rank tasks by weighted productivity score
- Select the budget-constrained sprint batch with the highest total value (exact 0/1 knapsack, 0.1 effort resolution)
- Expose ranking via lightweight HTTP API

## Run

```powershell
cd streamlane-go
go run ./cmd/streamlane -input sample/tasks.json -budget 10
```

## Test

```powershell
cd streamlane-go
go test ./...
```

## API

- `GET /health`
- `POST /rank` with JSON body `{ "tasks": [...], "budget": 10 }`


## How the plan is chosen

Tasks are ranked by value per unit of effort, where value is
`1.8*impact + 1.4*urgency + 0.6*(10 - risk)`. Filling the sprint greedily in
that order is the classic ratio heuristic, and for all-or-nothing tasks it can
be arbitrarily far from optimal: one cheap task taken first can strand the
budget so that two valuable tasks no longer fit.

The planner therefore solves the 0/1 knapsack exactly with dynamic programming.
Efforts are rounded up to 0.1, so a plan never exceeds the real budget. The
table is capped at 4M cells; past that it falls back to the greedy fill.

The API runs behind an `http.Server` with read, write, header and idle
timeouts, and caps request bodies at 1 MiB.
