package engine

import (
	"math/rand"
)

// Generator handles the physical creation of the crossword layout.
// It doesn't solve the words; it just carves out the white spaces.
type Generator struct {
	Puzzle *Puzzle
}

func NewGenerator(p *Puzzle) *Generator {
	return &Generator{Puzzle: p}
}

// GenerateLayout attempts to create a valid symmetric grid.
func (g *Generator) GenerateLayout(seed int64) ([]Variable, map[int][]Intersection) {
	g.Puzzle.Lock()
	defer g.Puzzle.Unlock()

	r := rand.New(rand.NewSource(seed))
	currentSeed := seed

	for {
		g.resetGrid()
		g.applySymmetricBlocks(r)

		if g.validateLayout() {
			// Extract word slots (Variables) and how they cross (Intersections)
			idCount := 0
			across := g.getAcrossVariables(&idCount)
			down := g.getDownVariables(&idCount)
			allVars := append(across, down...)

			return allVars, MapIntersections(allVars)
		}

		// If validation fails, try a new seed
		currentSeed++
		r.Seed(currentSeed)
	}
}

func (g *Generator) resetGrid() {
	for i := range g.Puzzle.Cells {
		g.Puzzle.Cells[i].IsBlock = false
		g.Puzzle.Cells[i].Char = ' '
	}
}

func (g *Generator) applySymmetricBlocks(r *rand.Rand) {
	total := len(g.Puzzle.Cells)
	half := total / 2
	targetPairs := 20 // Adjust this for crossword "density"

	for i := 0; i < targetPairs; i++ {
		idx := r.Intn(half)
		mirrorIdx := (total - 1) - idx

		g.Puzzle.Cells[idx].IsBlock = true
		g.Puzzle.Cells[mirrorIdx].IsBlock = true
	}
}

func (g *Generator) validateLayout() bool {
	for i, cell := range g.Puzzle.Cells {
		if cell.IsBlock {
			continue
		}

		row, col := i/Size, i%Size

		// Check horizontal continuity
		hLen := 1
		for tc := col - 1; tc >= 0 && !g.Puzzle.Cells[GetIndex(row, tc)].IsBlock; tc-- {
			hLen++
		}
		for tc := col + 1; tc < Size && !g.Puzzle.Cells[GetIndex(row, tc)].IsBlock; tc++ {
			hLen++
		}

		// Check vertical continuity
		vLen := 1
		for tr := row - 1; tr >= 0 && !g.Puzzle.Cells[GetIndex(tr, col)].IsBlock; tr-- {
			vLen++
		}
		for tr := row + 1; tr < Size && !g.Puzzle.Cells[GetIndex(tr, col)].IsBlock; tr++ {
			vLen++
		}

		// Standard crossword rule: no isolated letters or 2-letter words
		if hLen < 3 && vLen < 3 {
			return false
		}
	}
	return true
}

// Internal scanners used by the generator to identify Variable slots
func (g *Generator) getAcrossVariables(idStart *int) []Variable {
	var vars []Variable
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			idx := GetIndex(r, c)
			if g.Puzzle.Cells[idx].IsBlock {
				continue
			}

			if c == 0 || g.Puzzle.Cells[GetIndex(r, c-1)].IsBlock {
				var indices []int
				currC := c
				for currC < Size && !g.Puzzle.Cells[GetIndex(r, currC)].IsBlock {
					indices = append(indices, GetIndex(r, currC))
					currC++
				}
				if len(indices) >= 3 {
					vars = append(vars, Variable{ID: *idStart, Indices: indices, Length: len(indices), Direction: "ACROSS"})
					*idStart++
				}
				c = currC
			}
		}
	}
	return vars
}

func (g *Generator) getDownVariables(idStart *int) []Variable {
	var vars []Variable
	for c := 0; c < Size; c++ {
		for r := 0; r < Size; r++ {
			idx := GetIndex(r, c)
			if g.Puzzle.Cells[idx].IsBlock {
				continue
			}

			if r == 0 || g.Puzzle.Cells[GetIndex(r-1, c)].IsBlock {
				var indices []int
				currR := r
				for currR < Size && !g.Puzzle.Cells[GetIndex(currR, c)].IsBlock {
					indices = append(indices, GetIndex(currR, c))
					currR++
				}
				if len(indices) >= 3 {
					vars = append(vars, Variable{ID: *idStart, Indices: indices, Length: len(indices), Direction: "DOWN"})
					*idStart++
				}
				r = currR
			}
		}
	}
	return vars
}
