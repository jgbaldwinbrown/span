package span

import (
	"cmp"
	"slices"
)

type OrderedSet[S Spanner[T], T any] struct {
	Cmpf func(a, b T) int
	Center T
	Left *OrderedSet[S, T]
	Right *OrderedSet[S, T]
	StartSorted []S
	EndSorted []S
}

func (s *OrderedSet[S, T]) Touching(t Spanner[T]) bool {
	if s == nil {
		return false
	}
	if SpanTouchingPointFunc(t, s.Center, s.Cmpf) {
		for _, span := range s.StartSorted {
			if TouchingFunc(s.Cmpf, span, t) {
				return true
			}
		}
		if s.Left.Touching(t) {
			return true
		}
		return s.Right.Touching(t)
	}
	if s.Cmpf(t.Right(), s.Center) <= 0 {
		for _, span := range s.StartSorted {
			if s.Cmpf(span.Left(), t.Right()) >= 0 {
				break
			}
			if TouchingFunc(s.Cmpf, span, t) {
				return true
			}
		}
		return s.Left.Touching(t)
	}
	for i := len(s.EndSorted)-1; i >= 0; i-- {
		span := s.EndSorted[i]
		if s.Cmpf(span.Right(), t.Left()) <= 0 {
			break
		}
		if TouchingFunc(s.Cmpf, span, t) {
			return true
		}
	}
	return s.Right.Touching(t)
}

func (s *OrderedSet[S, T]) AppendTouched(dst *[]S, t Spanner[T]) {
	if s == nil {
		return
	}
	if SpanTouchingPointFunc(t, s.Center, s.Cmpf) {
		for _, span := range s.StartSorted {
			if TouchingFunc(s.Cmpf, span, t) {
				*dst = append(*dst, span)
			}
		}
		s.Left.AppendTouched(dst, t)
		s.Right.AppendTouched(dst, t)
		return
	}
	if s.Cmpf(t.Right(), s.Center) <= 0 {
		for _, span := range s.StartSorted {
			if s.Cmpf(span.Left(), t.Right()) >= 0 {
				break
			}
			if TouchingFunc(s.Cmpf, span, t) {
				*dst = append(*dst, span)
			}
		}
		s.Left.AppendTouched(dst, t)
		return
	}
	for i := len(s.EndSorted)-1; i >= 0; i-- {
		span := s.EndSorted[i]
		if s.Cmpf(span.Right(), t.Left()) <= 0 {
			break
		}
		if TouchingFunc(s.Cmpf, span, t) {
			*dst = append(*dst, span)
		}
	}
	s.Right.AppendTouched(dst, t)
}

func (s *OrderedSet[S, T]) Touched(t Spanner[T]) []S {
	var out []S
	s.AppendTouched(&out, t)
	return out
}

func SpanTouchingPointFunc[S Spanner[T], T any](s S, t T, cmpf func(x, y T) int) bool {
	return cmpf(t, s.Left()) >= 0 && cmpf(t, s.Right()) < 0
}

func SpanTouchingPoint[S Spanner[T], T cmp.Ordered](s S, t T) bool {
	return SpanTouchingPointFunc(s, t, cmp.Compare)
}

func (s *OrderedSet[S, T]) TouchingPoint(t T) bool {
	if s == nil {
		return false
	}
	if s.Cmpf(t, s.Center) == 0 {
		for _, span := range s.StartSorted {
			if SpanTouchingPointFunc(span, t, s.Cmpf) {
				return true
			}
		}
		return false
	}
	if s.Cmpf(t, s.Center) < 0 {
		for _, span := range s.StartSorted {
			if SpanTouchingPointFunc(span, t, s.Cmpf) {
				return true
			}
			break
		}
		return s.Left.TouchingPoint(t)
	}
	for i := len(s.EndSorted) - 1; i >= 0; i-- {
		span := s.EndSorted[i]
		if SpanTouchingPointFunc(span, t, s.Cmpf) {
			return true
		}
		break
	}
	return s.Right.TouchingPoint(t)
}

func CmpLeftFunc[S Spanner[T], T any](cmpf func(x, y T) int) func(x, y S) int {
	return func(a, b S) int {
		if cmpf(a.Left(), b.Left()) < 0 {
			return -1
		}
		if cmpf(b.Left(), a.Left()) < 0 {
			return 1
		}
		return 0
	}
}

func CmpLeft[S Spanner[T], T cmp.Ordered]() func(x, y S) int {
	return CmpLeftFunc[S, T](cmp.Compare)
}

func CmpRightFunc[S Spanner[T], T any](cmpf func(x, y T) int) func(x, y S) int {
	return func(a, b S) int {
		if cmpf(a.Right(), b.Right()) < 0 {
			return -1
		}
		if cmpf(b.Right(), a.Right()) < 0 {
			return 1
		}
		return 0
	}
}

func CmpRight[S Spanner[T], T cmp.Ordered]() func(x, y S) int {
	return CmpRightFunc[S, T](cmp.Compare)
}

func NewOrderedSetFunc[S Spanner[T], T any](values []S, cmpf func(x, y T) int) *OrderedSet[S, T] {
	if len(values) < 1 {
		return nil
	}
	if !slices.IsSortedFunc(values, CmpLeftFunc[S, T](cmpf)) {
		slices.SortFunc(values, CmpLeftFunc[S, T](cmpf))
	}
	toLeft := []S{}
	toRight := []S{}
	set := &OrderedSet[S, T]{
		Center: values[len(values)/2].Left(),
		Cmpf: cmpf,
	}
	for _, val := range values {
		if set.Cmpf(set.Center, val.Left()) >= 0 && set.Cmpf(set.Center, val.Right()) < 0 {
			set.StartSorted = append(set.StartSorted, val)
		} else if set.Cmpf(set.Center, val.Right()) >= 0 {
			toLeft = append(toLeft, val)
		} else {
			toRight = append(toRight, val)
		}
	}
	set.EndSorted = slices.Clone(set.StartSorted)
	slices.SortFunc(set.EndSorted, CmpRightFunc[S, T](cmpf))
	set.Left = NewOrderedSetFunc(toLeft, cmpf)
	set.Right = NewOrderedSetFunc(toRight, cmpf)
	return set
}

func NewOrderedSet[S Spanner[T], T cmp.Ordered](values []S) *OrderedSet[S, T] {
	return NewOrderedSetFunc(values, cmp.Compare)
}
