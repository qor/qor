package sorting

type Position struct {
	Position int
}

func (Position) PositionMoveUp(pos int) {
	// update other's position: (current position + 1 .. current position + pos) - 1
	// update self's position: current position + pos
}

func (Position) PositionMoveDown(pos int) {
	// update other's position: (current position + 1 .. current position + pos) + 1
	// update self's position: current position - pos
}

func (Position) PositionMoveTo(pos int) {
	// if pos < current position
	//   update other's position: (pos .. current position) + 1
	//   update self's position: pos
	// else if pos > current position
	//   update other's position: (current position .. pos) - 1
	//   update self's position: pos
}
