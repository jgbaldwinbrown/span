package span

import (
	"testing"
	"fmt"
)

type intPos int

func (p intPos) Cmp(p2 Pos) int {
	return int(p) - int(p2.(intPos))
}

type intSpan struct {
	left int
	right int
}

func (s intSpan) Left() Pos {
	return intPos(s.left)
}

func (s intSpan) Right() Pos {
	return intPos(s.right)
}

func TestSet(t *testing.T) {
	var s Set
	s.Add(intSpan{0, 1})
	s.Add(intSpan{-5, 22})
	s.Add(intSpan{3, 4})
	fmt.Println(s)
}
