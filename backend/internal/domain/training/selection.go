package training

import "sort"

type SelectedProblem struct {
	Slot SlotConfig
	Problem Problem
}

func SelectProblemSet(config Config, candidates []Problem, random RandomSource) ([]SelectedProblem, int, error) {
	for level := 0; level <= 3; level++ {
		pools := make(map[string][]Problem, len(config.Slots))
		for _, slot := range config.Slots {
			for _, problem := range candidates {
				if problem.Index != slot.ProblemIndex || !difficultyMatches(problem.Difficulty, slot, level) { continue }
				pools[slot.Name] = append(pools[slot.Name], problem)
			}
			random.Shuffle(len(pools[slot.Name]), func(i, j int) { pools[slot.Name][i], pools[slot.Name][j] = pools[slot.Name][j], pools[slot.Name][i] })
		}
		order := append([]SlotConfig(nil), config.Slots...)
		sort.SliceStable(order, func(i, j int) bool { return len(pools[order[i].Name]) < len(pools[order[j].Name]) })
		chosen := make(map[string]Problem, len(order))
		if chooseSlots(order, pools, 0, map[string]bool{}, chosen) {
			result := make([]SelectedProblem, 0, len(config.Slots))
			for _, slot := range config.Slots { result = append(result, SelectedProblem{Slot: slot, Problem: chosen[slot.Name]}) }
			return result, level, nil
		}
	}
	return nil, 0, ErrProblemSetUnavailable
}

func chooseSlots(order []SlotConfig, pools map[string][]Problem, index int, used map[string]bool, chosen map[string]Problem) bool {
	if index == len(order) { return true }
	slot := order[index]
	for _, problem := range pools[slot.Name] {
		if used[problem.ContestID] { continue }
		used[problem.ContestID] = true
		chosen[slot.Name] = problem
		if chooseSlots(order, pools, index+1, used, chosen) { return true }
		delete(chosen, slot.Name)
		delete(used, problem.ContestID)
	}
	return false
}

func difficultyMatches(value *int, slot SlotConfig, level int) bool {
	if level == 3 { return true }
	if value == nil { return false }
	expansion := level * 100
	return *value >= slot.DifficultyMin-expansion && *value <= slot.DifficultyMax+expansion
}

