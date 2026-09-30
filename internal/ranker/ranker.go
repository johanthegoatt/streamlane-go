package ranker

import (
	"math"
	"sort"
)

type Task struct {
	Name    string  `json:"name"`
	Impact  float64 `json:"impact"`
	Urgency float64 `json:"urgency"`
	Effort  float64 `json:"effort"`
	Risk    float64 `json:"risk"`
}

type RankedTask struct {
	Task
	Score float64 `json:"score"`
}

func Score(task Task) float64 {
	effort := task.Effort
	if effort < 0.25 {
		effort = 0.25
	}

	return Value(task) / effort
}

func RankTasks(tasks []Task) []RankedTask {
	result := make([]RankedTask, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, RankedTask{
			Task:  task,
			Score: Score(task),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Name < result[j].Name
		}
		return result[i].Score > result[j].Score
	})
	return result
}

// Value is the undivided benefit of a task: the numerator of Score. Score
// ranks by value per unit of effort, but a sprint plan should maximise the
// total value that fits in the budget, which is a different question.
func Value(task Task) float64 {
	riskMitigation := 10.0 - task.Risk
	if riskMitigation < 0 {
		riskMitigation = 0
	}
	return (task.Impact * 1.8) + (task.Urgency * 1.4) + (riskMitigation * 0.6)
}

const (
	// effortUnit is the resolution the planner works at. Efforts are rounded
	// up to it, so a plan can never exceed the real budget.
	effortUnit = 0.1
	// maxDPCells bounds the knapsack table (tasks x budget units) so a huge
	// budget cannot allocate unbounded memory; beyond it we fall back to greedy.
	maxDPCells = 4_000_000
)

// BuildSprintPlan picks the set of tasks with the highest total Value whose
// effort fits in budget. Filling greedily by score-per-effort can be
// arbitrarily far from optimal for 0/1 selection (one cheap task can block
// two valuable ones), so this solves the 0/1 knapsack exactly with dynamic
// programming. The plan is returned in rank order.
func BuildSprintPlan(tasks []Task, budget float64) ([]RankedTask, float64) {
	ranked := RankTasks(tasks)
	capacity := int(math.Floor(budget/effortUnit + 1e-9))
	if capacity <= 0 || len(ranked) == 0 {
		return []RankedTask{}, 0
	}
	if (len(ranked)+1)*(capacity+1) > maxDPCells {
		return greedyPlan(ranked, budget)
	}

	weights := make([]int, len(ranked))
	for i, task := range ranked {
		weights[i] = int(math.Ceil(math.Max(task.Effort, 0)/effortUnit - 1e-9))
	}

	// best[i][c]: highest value using the first i ranked tasks within c units.
	best := make([][]float64, len(ranked)+1)
	best[0] = make([]float64, capacity+1)
	for i, task := range ranked {
		prev, row := best[i], make([]float64, capacity+1)
		value := Value(task.Task)
		for c := 0; c <= capacity; c++ {
			row[c] = prev[c]
			if w := weights[i]; w <= c && prev[c-w]+value > row[c] {
				row[c] = prev[c-w] + value
			}
		}
		best[i+1] = row
	}

	chosen := make([]bool, len(ranked))
	for i, c := len(ranked), capacity; i > 0; i-- {
		if best[i][c] != best[i-1][c] {
			chosen[i-1] = true
			c -= weights[i-1]
		}
	}

	plan := make([]RankedTask, 0, len(ranked))
	used := 0.0
	for i, task := range ranked {
		if chosen[i] {
			plan = append(plan, task)
			used += task.Effort
		}
	}
	return plan, used
}

func greedyPlan(ranked []RankedTask, budget float64) ([]RankedTask, float64) {
	plan := make([]RankedTask, 0, len(ranked))
	used := 0.0
	for _, task := range ranked {
		if used+task.Effort > budget {
			continue
		}
		plan = append(plan, task)
		used += task.Effort
	}
	return plan, used
}
