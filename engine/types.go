package engine

// Shared constants
const Size = 15

// Variable represents a word slot in the grid
type Variable struct {
	ID        int    `json:"id"`
	Indices   []int  `json:"indices"` // Ensure this is lowercase in the tag
	Length    int    `json:"length"`
	Direction string `json:"direction"`
	Value     string `json:"value"`
}

// Intersection defines where two Variables meet
type Intersection struct {
	Index       int `json:"index"`
	SelfOffset  int `json:"selfOffset"`
	OtherVarID  int `json:"otherVarId"`
	OtherOffset int `json:"otherOffset"`
}

// GetIndex converts 2D coordinates to a 1D array index
func GetIndex(row, col int) int {
	if row < 0 || row >= Size || col < 0 || col >= Size {
		return -1
	}
	return row*Size + col
}

// MapIntersections identifies where Across and Down words collide.
// This is used by the Generator to build the constraint map for the Solver.
func MapIntersections(vars []Variable) map[int][]Intersection {
	constraints := make(map[int][]Intersection)
	gridMap := make(map[int]int) // index -> variableID (for ACROSS)

	// Phase 1: Map all Across indices
	for _, v := range vars {
		if v.Direction == "ACROSS" {
			for _, idx := range v.Indices {
				gridMap[idx] = v.ID
			}
		}
	}

	// Phase 2: Find Down variables that hit those indices
	for _, v := range vars {
		if v.Direction == "DOWN" {
			for downOff, idx := range v.Indices {
				if acrossID, exists := gridMap[idx]; exists {
					// Find the across variable and its offset
					var acrossVar *Variable
					for i := range vars {
						if vars[i].ID == acrossID {
							acrossVar = &vars[i]
							break
						}
					}

					acrossOff := -1
					for i, aIdx := range acrossVar.Indices {
						if aIdx == idx {
							acrossOff = i
							break
						}
					}

					// Add mutually to both directions
					constraints[v.ID] = append(constraints[v.ID], Intersection{
						Index: idx, SelfOffset: downOff, OtherVarID: acrossID, OtherOffset: acrossOff,
					})
					constraints[acrossID] = append(constraints[acrossID], Intersection{
						Index: idx, SelfOffset: acrossOff, OtherVarID: v.ID, OtherOffset: downOff,
					})
				}
			}
		}
	}
	return constraints
}
