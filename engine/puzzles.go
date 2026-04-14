package engine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
)

type Cell struct {
	Char    rune `json:"char"`
	IsBlock bool `json:"isBlock"` // Svelte uses cell.isBlock
	Index   int  `json:"index"`
}

type Puzzle struct {
	sync.RWMutex
	Cells []Cell `json:"cells"`
}

func NewPuzzle() *Puzzle {
	p := &Puzzle{Cells: make([]Cell, Size*Size)}
	for i := 0; i < Size*Size; i++ {
		p.Cells[i] = Cell{Index: i, Char: ' ', IsBlock: false}
	}
	return p
}

// GenerateAndExport creates the physical layout (The Generator)
func (p *Puzzle) GenerateAndExport(seed int64) ([]Variable, map[int][]Intersection) {
	p.Lock()
	defer p.Unlock()
	r := rand.New(rand.NewSource(seed))

	for {
		// Reset and place random symmetric blocks
		for i := range p.Cells {
			p.Cells[i].IsBlock = false
		}
		for i := 0; i < 20; i++ {
			idx := r.Intn((Size * Size) / 2)
			p.Cells[idx].IsBlock = true
			p.Cells[(Size*Size-1)-idx].IsBlock = true
		}

		if p.validate_layout() {
			idCount := 0
			allVars := append(p.GetAcrossVariables(&idCount), p.GetDownVariables(&idCount)...)
			return allVars, MapIntersections(allVars)
		}
		seed++
		r.Seed(seed)
	}
}

type Dictionary struct {
	// Maps word length -> slice of words
	Words map[int][]string
}

// NewDictionary reads the JSON file and prepares the word domains
func NewDictionary(path string) (*Dictionary, error) {
	fmt.Printf("[SYSTEM] Loading dictionary from: %s\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dictionary file: %w", err)
	}

	var raw map[string][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse dictionary JSON: %w", err)
	}

	formatted := make(map[int][]string)
	totalWords := 0

	for lenStr, list := range raw {
		length, err := strconv.Atoi(lenStr)
		if err != nil {
			continue
		}
		formatted[length] = list
		totalWords += len(list)
	}

	fmt.Printf("[SYSTEM] Dictionary loaded successfully. Total words: %d\n", totalWords)
	return &Dictionary{Words: formatted}, nil
}

// ... inside engine/puzzle.go ...

func (p *Puzzle) validate_layout() bool {
	for i := range p.Cells {
		if p.Cells[i].IsBlock {
			continue
		}
		r, c := i/Size, i%Size
		hLen, vLen := 1, 1
		for tc := c - 1; tc >= 0 && !p.Cells[GetIndex(r, tc)].IsBlock; tc-- {
			hLen++
		}
		for tc := c + 1; tc < Size && !p.Cells[GetIndex(r, tc)].IsBlock; tc++ {
			hLen++
		}
		for tr := r - 1; tr >= 0 && !p.Cells[GetIndex(tr, c)].IsBlock; tr-- {
			vLen++
		}
		for tr := r + 1; tr < Size && !p.Cells[GetIndex(tr, c)].IsBlock; tr++ {
			vLen++
		}
		if hLen < 3 && vLen < 3 {
			return false
		}
	}
	return true
}

func (p *Puzzle) GetAcrossVariables(idStart *int) []Variable {
	var vars []Variable
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			idx := GetIndex(r, c)
			if p.Cells[idx].IsBlock {
				continue
			}
			if c == 0 || p.Cells[GetIndex(r, c-1)].IsBlock {
				var indices []int
				currC := c
				for currC < Size && !p.Cells[GetIndex(r, currC)].IsBlock {
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

func (p *Puzzle) GetDownVariables(idStart *int) []Variable {
	var vars []Variable
	for c := 0; c < Size; c++ {
		for r := 0; r < Size; r++ {
			idx := GetIndex(r, c)
			if p.Cells[idx].IsBlock {
				continue
			}
			if r == 0 || p.Cells[GetIndex(r-1, c)].IsBlock {
				var indices []int
				currR := r
				for currR < Size && !p.Cells[GetIndex(currR, c)].IsBlock {
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

// (Keep your validate_layout and GetAcross/Down methods here...)
