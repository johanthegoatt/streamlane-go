# Streamlane Go

Go-based sprint prioritization service for turning raw task signals into ranked execution plans.

## Features

- Rank tasks by weighted productivity score
- Select a budget-constrained sprint batch
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

