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

func TestBuildSprintPlanBeatsGreedyWhenCheapTaskBlocks(t *testing.T) {
	// "quick" has the best score per effort, so a greedy fill takes it first
	// and then only one "big" fits. Two bigs are worth far more.
	tasks := []Task{
		{Name: "quick", Impact: 3, Urgency: 3, Effort: 1, Risk: 5},
		{Name: "big-a", Impact: 10, Urgency: 10, Effort: 5, Risk: 0},
		{Name: "big-b", Impact: 10, Urgency: 10, Effort: 5, Risk: 0},
	}

	plan, used := BuildSprintPlan(tasks, 10)
	if used != 10 || len(plan) != 2 {
		t.Fatalf("expected both big tasks using 10, got %d tasks using %v", len(plan), used)
	}
	for _, task := range plan {
		if task.Name == "quick" {
			t.Fatalf("plan should not include quick: %+v", plan)
		}
	}

	_, greedyUsed := greedyPlan(RankTasks(tasks), 10)
	if greedyUsed != 6 {
		t.Fatalf("expected greedy to strand budget at 6, got %v", greedyUsed)
	}
}

func TestBuildSprintPlanNeverExceedsBudgetWithFractionalEffort(t *testing.T) {
	tasks := []Task{
		{Name: "a", Impact: 9, Urgency: 9, Effort: 3.34, Risk: 1},
		{Name: "b", Impact: 9, Urgency: 9, Effort: 3.33, Risk: 1},
		{Name: "c", Impact: 9, Urgency: 9, Effort: 3.34, Risk: 1},
	}

	_, used := BuildSprintPlan(tasks, 10)
	if used > 10 {
		t.Fatalf("used %v exceeds budget 10", used)
	}
}

func TestBuildSprintPlanZeroBudget(t *testing.T) {
	plan, used := BuildSprintPlan([]Task{{Name: "a", Effort: 1}}, 0)
	if len(plan) != 0 || used != 0 {
		t.Fatalf("expected empty plan, got %+v", plan)
	}
}
