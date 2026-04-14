package engine

type GeneratorSolver struct {
	Variables   []Variable
	Constraints map[int][]Intersection
	Dictionary  map[int][]string
	GridState   []rune
	domainCache []int
	OnPlace     func(id int, word string)
	OnBacktrack func(id int, word string)
}

func NewGeneratorSolver(vars []Variable, constraints map[int][]Intersection, dict map[int][]string) *GeneratorSolver {
	gs := &GeneratorSolver{
		Variables:   vars,
		Constraints: constraints,
		Dictionary:  dict,
		GridState:   make([]rune, Size*Size),
		domainCache: make([]int, len(vars)),
	}
	for i := range gs.GridState {
		gs.GridState[i] = ' '
	}
	return gs
}

func (gs *GeneratorSolver) Solve() bool {
	vIdx := gs.GetBestVariable()
	if vIdx == -1 {
		return true
	} // Success
	if vIdx == -2 {
		return false
	} // Dead end

	v := &gs.Variables[vIdx]
	for _, word := range gs.Dictionary[v.Length] {
		if gs.fits(v, word) {
			if gs.OnPlace != nil {
				gs.OnPlace(v.ID, word)
			}
			backup := gs.applyWord(v, word)
			v.Value = word

			if gs.Solve() {
				return true
			}

			v.Value = ""
			gs.undoWord(v, backup)
			if gs.OnBacktrack != nil {
				gs.OnBacktrack(v.ID, "")
			}
		}
	}
	return false
}

// ... inside engine/solvers.go ...
func (gs *GeneratorSolver) GetBestVariable() int {
	bestIdx := -1
	minRemaining := 1000000
	foundAnyEmpty := false

	for i := range gs.Variables {
		if gs.Variables[i].Value != "" {
			continue
		}
		foundAnyEmpty = true
		count := gs.countPossibleWords(&gs.Variables[i])

		// If a variable has NO options, this branch is dead.
		// Return a specific error code or handle it.
		if count == 0 {
			return -2 // Signal that this is a FAILURE, not COMPLETION
		}

		if count < minRemaining {
			minRemaining = count
			bestIdx = i
		}
	}

	if !foundAnyEmpty {
		return -1 // Signal COMPLETION
	}
	return bestIdx
}

func (gs *GeneratorSolver) countPossibleWords(v *Variable) int {
	count := 0
	for _, word := range gs.Dictionary[v.Length] {
		match := true
		for i, idx := range v.Indices {
			if gs.GridState[idx] != ' ' && gs.GridState[idx] != rune(word[i]) {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}
	return count
}

func (gs *GeneratorSolver) fits(v *Variable, word string) bool {
	// Safety check: Ensure the word length matches the variable length
	if len(word) != v.Length {
		return false
	}

	for i, char := range word {
		// Safety check: Ensure we don't exceed GridState or Indices
		if i >= len(v.Indices) {
			break
		}
		gridIdx := v.Indices[i]
		if gs.GridState[gridIdx] != ' ' && gs.GridState[gridIdx] != rune(char) {
			return false
		}
	}

	for _, inter := range gs.Constraints[v.ID] {
		neighbor := &gs.Variables[inter.OtherVarID]
		if neighbor.Value != "" {
			continue
		}

		// The panic likely happened here or in neighborHasOptions
		// checking if the character at our intersection offset fits the neighbor
		if inter.SelfOffset >= len(word) {
			continue // Should not happen if lengths are correct
		}

		if !gs.neighborHasOptions(neighbor, inter.OtherOffset, rune(word[inter.SelfOffset])) {
			return false
		}
	}
	return true
}

func (gs *GeneratorSolver) neighborHasOptions(neighbor *Variable, offset int, char rune) bool {
	words, exists := gs.Dictionary[neighbor.Length]
	if !exists {
		return false
	}

	for _, word := range words {
		// Safety: Ensure the dictionary word is actually the right length
		// and the offset is within bounds
		if len(word) != neighbor.Length || offset >= len(word) {
			continue
		}

		if rune(word[offset]) != char {
			continue
		}

		match := true
		for i, idx := range neighbor.Indices {
			if gs.GridState[idx] != ' ' && gs.GridState[idx] != rune(word[i]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func (gs *GeneratorSolver) applyWord(v *Variable, word string) []rune {
	backup := make([]rune, len(v.Indices))
	for i, idx := range v.Indices {
		backup[i] = gs.GridState[idx]
		gs.GridState[idx] = rune(word[i])
	}
	return backup
}

func (gs *GeneratorSolver) undoWord(v *Variable, backup []rune) {
	for i, idx := range v.Indices {
		gs.GridState[idx] = backup[i]
	}
}

// (Keep your fits, GetBestVariable, and neighborHasOptions methods here...)
