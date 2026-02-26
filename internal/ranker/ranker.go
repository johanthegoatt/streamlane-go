package ranker

import "sort"

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

	riskMitigation := 10.0 - task.Risk
	if riskMitigation < 0 {
		riskMitigation = 0
	}

	return ((task.Impact * 1.8) + (task.Urgency * 1.4) + (riskMitigation * 0.6)) / effort
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

func BuildSprintPlan(tasks []Task, budget float64) ([]RankedTask, float64) {
	ranked := RankTasks(tasks)
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

