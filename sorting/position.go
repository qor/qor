package sorting

type positionInterface interface {
	GetPosition() int
	SetPosition(int)
}

type Position struct {
	Position int
}

func (position Position) GetPosition() int {
	return position.Position
}

func (position Position) SetPosition(pos int) {
	position.Position = pos
}

func MoveUp(value positionInterface, pos int) {
	// update other's position: (current position + 1 .. current position + pos) - 1
	// update self's position: current position + pos
}

func MoveDown(value positionInterface, pos int) {
	// update other's position: (current position + 1 .. current position + pos) + 1
	// update self's position: current position - pos
}

func MoveTo(value positionInterface, pos int) {
	// if pos < current position
	//   update other's position: (pos .. current position) + 1
	//   update self's position: pos
	// else if pos > current position
	//   update other's position: (current position .. pos) - 1
	//   update self's position: pos
}
