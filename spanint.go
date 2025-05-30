package span

import (
	"cmp"
	"slices"
)

type OrderedSet[S Spanner[T], T cmp.Ordered] struct {
	Center T
	Left *OrderedSet[S, T]
	Right *OrderedSet[S, T]
	StartSorted []S
	EndSorted []S
}

func (s *OrderedSet[S, T]) Touching(t S) bool {
	if s == nil {
		return false
	}
	if SpanTouchingPoint(t, s.Center) {
		for _, span := range s.StartSorted {
			if Touching(span, t) {
				return true
			}
		}
		if s.Left.Touching(t) {
			return true
		}
		return s.Right.Touching(t)
	}
	if t.Right() <= s.Center {
		for _, span := range s.StartSorted {
			if span.Left() >= t.Right() {
				break
			}
			if Touching(span, t) {
				return true
			}
		}
		return s.Left.Touching(t)
	}
	for i := len(s.EndSorted)-1; i >= 0; i-- {
		span := s.EndSorted[i]
		if span.Right() <= t.Left() {
			break
		}
		if Touching(span, t) {
			return true
		}
	}
	return s.Right.Touching(t)
}

func SpanTouchingPoint[S Spanner[T], T cmp.Ordered](s S, t T) bool {
	return t >= s.Left() && t < s.Right()
}

func (s *OrderedSet[S, T]) TouchingPoint(t T) bool {
	if s == nil {
		return false
	}
	if t == s.Center {
		for _, span := range s.StartSorted {
			if SpanTouchingPoint(span, t) {
				return true
			}
		}
		return false
	}
	if t < s.Center {
		for _, span := range s.StartSorted {
			if SpanTouchingPoint(span, t) {
				return true
			}
			break
		}
		return s.Left.TouchingPoint(t)
	}
	for i := len(s.EndSorted) - 1; i >= 0; i-- {
		span := s.EndSorted[i]
		if SpanTouchingPoint(span, t) {
			return true
		}
		break
	}
	return s.Right.TouchingPoint(t)
}

func CmpLeft[S Spanner[T], T cmp.Ordered](a, b S) int {
	if a.Left() < b.Left() {
		return -1
	}
	if b.Left() < a.Left() {
		return 1
	}
	return 0
}

func CmpRight[S Spanner[T], T cmp.Ordered](a, b S) int {
	if a.Right() < b.Right() {
		return -1
	}
	if b.Right() < a.Right() {
		return 1
	}
	return 0
}

func NewOrderedSet[S Spanner[T], T cmp.Ordered](values []S) *OrderedSet[S, T] {
	if len(values) < 1 {
		return nil
	}
	if !slices.IsSortedFunc(values, CmpLeft) {
		slices.SortFunc(values, CmpLeft)
	}
	toLeft := []S{}
	toRight := []S{}
	set := &OrderedSet[S, T]{
		Center: values[len(values)/2].Left(),
	}
	for _, val := range values {
		if set.Center >= val.Left() && set.Center < val.Right() {
			set.StartSorted = append(set.StartSorted, val)
		} else if set.Center >= val.Right() {
			toLeft = append(toLeft, val)
		} else {
			toRight = append(toRight, val)
		}
	}
	set.EndSorted = slices.Clone(set.StartSorted)
	slices.SortFunc(set.EndSorted, CmpRight)
	set.Left = NewOrderedSet(toLeft)
	set.Right = NewOrderedSet(toRight)
	return set
}

func NewOrderedSetDynamic[S Spanner[T], T cmp.Ordered](values []S) *OrderedSet[S, T] {
	return NewOrderedSet(values)
}
