package ranker

import "testing"

func TestRankTasksOrdersByScoreDesc(t *testing.T) {
	tasks := []Task{
		{Name: "low", Impact: 4, Urgency: 4, Effort: 5, Risk: 4},
		{Name: "high", Impact: 9, Urgency: 9, Effort: 3, Risk: 2},
	}

	ranked := RankTasks(tasks)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(ranked))
	}

	if ranked[0].Name != "high" {
		t.Fatalf("expected high first, got %s", ranked[0].Name)
	}
}

func TestBuildSprintPlanRespectsBudget(t *testing.T) {
	tasks := []Task{
		{Name: "a", Impact: 8, Urgency: 8, Effort: 4, Risk: 2},
		{Name: "b", Impact: 7, Urgency: 7, Effort: 4, Risk: 3},
		{Name: "c", Impact: 9, Urgency: 6, Effort: 6, Risk: 2},
	}

	plan, used := BuildSprintPlan(tasks, 8)
	if used > 8.0 {
		t.Fatalf("used effort exceeded budget: %f", used)
	}

	if len(plan) == 0 {
		t.Fatalf("expected non-empty plan")
	}
}

