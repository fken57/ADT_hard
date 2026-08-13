package training

import (
	"testing"
)

type noShuffle struct{}

func (noShuffle) Intn(_ int) int                  { return 0 }
func (noShuffle) Shuffle(_ int, _ func(i, j int)) {}
func intPointer(value int) *int                   { return &value }

func TestSelectProblemSetUsesDistinctContestsAndConfiguredOrder(t *testing.T) {
	config := DefaultConfig()
	candidates := []Problem{
		{ID: "d-a", ContestID: "a", Index: "D", Difficulty: intPointer(1000)},
		{ID: "e-a", ContestID: "a", Index: "E", Difficulty: intPointer(1200)},
		{ID: "e-b", ContestID: "b", Index: "E", Difficulty: intPointer(1150)},
		{ID: "e-c", ContestID: "c", Index: "E", Difficulty: intPointer(1300)},
		{ID: "e-d", ContestID: "d", Index: "E", Difficulty: intPointer(1500)},
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
	wantSlots := []string{"Warmup", "Stable", "Main", "Stretch", "Challenge"}
	for index, item := range selected {
		if item.Slot.Name != wantSlots[index] {
			t.Fatalf("slot %d = %s", index, item.Slot.Name)
		}
		if used[item.Problem.ContestID] {
			t.Fatalf("duplicate contest %s", item.Problem.ContestID)
		}
		used[item.Problem.ContestID] = true
		if index > 0 && *selected[index-1].Problem.Difficulty > *item.Problem.Difficulty {
			t.Fatalf("difficulty decreased from %d to %d", *selected[index-1].Problem.Difficulty, *item.Problem.Difficulty)
		}
	}
}

func TestDifficultiesAscendingRejectsAnInversion(t *testing.T) {
	slots := []SlotConfig{{Name: "Stable"}, {Name: "Main"}}
	chosen := map[string]Problem{
		"Stable": {Difficulty: intPointer(1200)},
		"Main":   {Difficulty: intPointer(1180)},
	}
	if difficultiesAscending(slots, chosen) {
		t.Fatal("a decreasing difficulty sequence must be rejected")
	}
}

func TestSelectProblemSetRejectsMissingDifficulty(t *testing.T) {
	config := DefaultConfig()
	candidates := []Problem{
		{ID: "d", ContestID: "a", Index: "D"}, {ID: "e1", ContestID: "b", Index: "E"},
		{ID: "e2", ContestID: "c", Index: "E"}, {ID: "e3", ContestID: "d", Index: "E"},
		{ID: "f", ContestID: "e", Index: "F"},
	}
	_, _, err := SelectProblemSet(config, candidates, noShuffle{})
	if err != ErrProblemSetUnavailable {
		t.Fatalf("err=%v", err)
	}
}

type rolledRandom struct{ value int }

func (random rolledRandom) Intn(_ int) int           { return random.value }
func (rolledRandom) Shuffle(_ int, _ func(i, j int)) {}

func TestSelectDifficultyProfileUsesConfiguredWeights(t *testing.T) {
	profiles := DefaultConfig().DifficultyProfiles
	for _, test := range []struct {
		roll int
		want string
	}{{0, "STANDARD"}, {69, "STANDARD"}, {70, "LIGHT"}, {84, "LIGHT"}, {85, "HEAVY"}, {99, "HEAVY"}} {
		if got := SelectDifficultyProfile(profiles, rolledRandom{value: test.roll}); got.Name != test.want {
			t.Fatalf("roll %d selected %s, want %s", test.roll, got.Name, test.want)
		}
	}
}

func TestSelectProblemSetReportsShortage(t *testing.T) {
	_, _, err := SelectProblemSet(DefaultConfig(), nil, noShuffle{})
	if err != ErrProblemSetUnavailable {
		t.Fatalf("err=%v", err)
	}
}
