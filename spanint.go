package span

import (
)

type IntSet[T Spanner[int64]] struct {
	winsize int64
	wins map[int64][]int
	spanners []T
	biggestBin int
	total int
}

func (s *IntSet[T]) Spanners() []T {
	return s.spanners
}

func leftWin(winsize, pos int64) int64 {
	return (pos / winsize) * pos
}

func rightWin(winsize, pos int64) int64 {
	if pos % winsize == 0 {
		return pos
	}
	return ((pos / winsize) + 1) * pos
}

func (s *IntSet[T]) Add(t T) {
	s.spanners = append(s.spanners, t)
	idx := len(s.spanners) - 1
	l, r := t.Left(), t.Right()
	lw, rw := leftWin(s.winsize, l), rightWin(s.winsize, r)
	for i := lw; i < rw; i += s.winsize {
		s.wins[i] = append(s.wins[i], idx)
		s.biggestBin = max(s.biggestBin, len(s.wins[i]))
	}
	s.total++
}

func (s *IntSet[T]) AddDynamic(t T) {
	s.Add(t)
	if s.winsize > 1 && s.biggestBin > s.total / 1000 {
		s2 := NewIntSet(s.winsize / 2, s.Spanners()...)
		*s = *s2
	}
}

func (s *IntSet[T]) addHitsIdxs(dst map[int]struct{}, t T) {
	l, r := t.Left(), t.Right()
	lw, rw := leftWin(s.winsize, l), rightWin(s.winsize, r)
	for i := lw; i < rw; i += s.winsize {
		for _, spidx := range s.wins[i] {
			if Touching(t, s.spanners[spidx]) {
				dst[spidx] = struct{}{}
			}
		}
	}
}

func (s *IntSet[T]) findHits(dst []T, t T) []T {
	m := map[int]struct{}{}
	s.addHitsIdxs(m, t)
	for idx, _ := range m {
		dst = append(dst, s.spanners[idx])
	}
	return dst
}

func NewIntSet[T Spanner[int64]](winsize int64, values ...T) *IntSet[T] {
	s := &IntSet[T]{}
	s.winsize = winsize
	for _, t := range values {
		s.Add(t)
	}
	return s
}

func NewIntSetDynamic[T Spanner[int64]](values ...T) *IntSet[T] {
	s := &IntSet[T]{}
	s.winsize = 1 << 32
	for _, t := range values {
		s.AddDynamic(t)
	}
	return s
}
