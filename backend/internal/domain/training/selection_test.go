package training

import (
	"testing"
)

type noShuffle struct{}

func (noShuffle) Shuffle(_ int, _ func(i, j int)) {}
func intPointer(value int) *int                   { return &value }

func TestSelectProblemSetUsesDistinctContestsAndConfiguredOrder(t *testing.T) {
	config := DefaultConfig()
	candidates := []Problem{
		{ID: "d-a", ContestID: "a", Index: "D", Difficulty: intPointer(1000)},
		{ID: "e-a", ContestID: "a", Index: "E", Difficulty: intPointer(1200)},
		{ID: "e-b", ContestID: "b", Index: "E", Difficulty: intPointer(1350)},
		{ID: "e-c", ContestID: "c", Index: "E", Difficulty: intPointer(1500)},
		{ID: "e-d", ContestID: "d", Index: "E", Difficulty: intPointer(1600)},
		{ID: "f-e", ContestID: "e", Index: "F", Difficulty: intPointer(1700)},
	}
	selected, level, err := SelectProblemSet(config, candidates, noShuffle{})
	if err != nil {
		t.Fatal(err)
	}
	if level != 0 || len(selected) != 5 {
		t.Fatalf("level=%d len=%d", level, len(selected))
	}
	used := map[string]bool{}
	wantSlots := []string{"D1", "E1", "E2", "E3", "F1"}
	for index, item := range selected {
		if item.Slot.Name != wantSlots[index] {
			t.Fatalf("slot %d = %s", index, item.Slot.Name)
		}
		if used[item.Problem.ContestID] {
			t.Fatalf("duplicate contest %s", item.Problem.ContestID)
		}
		used[item.Problem.ContestID] = true
	}
}

func TestSelectProblemSetFallsBackAndAllowsMissingDifficulty(t *testing.T) {
	config := DefaultConfig()
	candidates := []Problem{
		{ID: "d", ContestID: "a", Index: "D"}, {ID: "e1", ContestID: "b", Index: "E"},
		{ID: "e2", ContestID: "c", Index: "E"}, {ID: "e3", ContestID: "d", Index: "E"},
		{ID: "f", ContestID: "e", Index: "F"},
	}
	_, level, err := SelectProblemSet(config, candidates, noShuffle{})
	if err != nil || level != 3 {
		t.Fatalf("level=%d err=%v", level, err)
	}
}

func TestSelectProblemSetReportsShortage(t *testing.T) {
	_, _, err := SelectProblemSet(DefaultConfig(), nil, noShuffle{})
	if err != ErrProblemSetUnavailable {
		t.Fatalf("err=%v", err)
	}
}
