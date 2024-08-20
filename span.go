package span

import (
	"slices"
	"math"
	"iter"
)

type Cmpf[T any] func(T, T) int

type Spanner[T any] interface {
	Left() T
	Right() T
}

type Span[T any] struct {
	Start T
	End T
}

func (s Span[T]) Left() T {
	return s.Start
}

func (s Span[T]) Right() T {
	return s.End
}

func Min[T any](cmpf Cmpf[T], ps ...T) T {
	if len(ps) < 1 {
		var t T
		return t
	}

	out := ps[0]
	for _, p := range ps[1:] {
		if cmpf(p, out) < 0 {
			out = p
		}
	}

	return out
}

func Max[T any](cmpf Cmpf[T], ps ...T) T {
	if len(ps) < 1 {
		var t T
		return t
	}

	out := ps[0]
	for _, p := range ps[1:] {
		if cmpf(p, out) > 0 {
			out = p
		}
	}

	return out
}

func Touching[T any](cmpf Cmpf[T], s1, s2 Spanner[T]) bool {
	right := Min(cmpf, s1.Right(), s2.Right())
	left := Max(cmpf, s1.Left(), s2.Left())
	return cmpf(left, right) <= 0
}

func Overlapping[T any](cmpf Cmpf[T], s1, s2 Spanner[T]) bool {
	right := Min(cmpf, s1.Right(), s2.Right())
	left := Max(cmpf, s1.Left(), s2.Left())
	return cmpf(left, right) < 0
}

func Union[T any](cmpf Cmpf[T], s1, s2 Spanner[T]) (Span[T], bool) {
	if !Touching(cmpf, s1, s2) {
		return Span[T]{}, false
	}
	return Span[T]{Min(cmpf, s1.Left(), s2.Left()), Max(cmpf, s1.Right(), s2.Right())}, true
}

func Intersect[T any](cmpf Cmpf[T], s1, s2 Spanner[T]) (Span[T], bool) {
	if !Overlapping(cmpf, s1, s2) {
		return Span[T]{}, false
	}
	return Span[T]{Max(cmpf, s1.Left(), s2.Left()), Min(cmpf, s1.Right(), s2.Right())}, true
}

func Range[S Spanner[T], T any](cmpf Cmpf[T], ss ...S) Span[T] {
	if len(ss) < 1 {
		return Span[T]{}
	}
	lmin := ss[0].Left()
	rmax := ss[0].Right()
	for _, s := range ss {
		if cmpf(s.Left(), lmin) < 0 {
			lmin = s.Left()
		}
		if cmpf(s.Right(), rmax) > 0 {
			rmax = s.Right()
		}
	}

	return Span[T]{lmin, rmax}
}

func SortSpans[T any](cmpf Cmpf[T], ss []Spanner[T]) {
	slices.SortFunc(ss, func(a, b Spanner[T]) int {
		return cmpf(a.Left(), b.Left())
	})
}

type bucket[T any] struct {
	sorted bool
	full Span[T]
	members []Spanner[T]
}

func (b bucket[T]) Left() T {
	return b.full.Left()
}

func (b bucket[T]) Right() T {
	return b.full.Right()
}

func sortBuckets[T any](cmpf Cmpf[T], bs []bucket[T]) {
	slices.SortFunc(bs, func(a, b bucket[T]) int {
		return cmpf(a.full.Left(), a.full.Right())
	})
}

func (b *bucket[T]) Add(cmpf Cmpf[T], sp Spanner[T]) {
	b.sorted = false
	b.members = append(b.members, sp)

	b.full.Start = Min(cmpf, b.full.Left(), sp.Left())
	b.full.End = Max(cmpf, b.full.Right(), sp.Right())
}

func (b *bucket[T]) Sort(cmpf Cmpf[T]) {
	b.sorted = true
	SortSpans(cmpf, b.members)
}

type Set[T any] struct {
	sorted bool
	largestBucketSize int
	buckets []bucket[T]
}

func (s *Set[T]) Sort(cmpf Cmpf[T]) {
	s.sorted = true
	sortBuckets(cmpf, s.buckets)
}

func (s *Set[T]) firstTouch(cmpf Cmpf[T], sp Spanner[T]) (bi, si int) {
	for i, b := range s.buckets {
		if cmpf(sp.Right(), b.Left()) < 0 {
			return -1, -1
		}
		if Touching(cmpf, sp, b) {
			for j, m := range b.members {
				if Touching(cmpf, sp, m) {
					return i, j
				}
			}
		}
	}
	return len(s.buckets), -1
}

func (s *Set[T]) Touching(cmpf Cmpf[T], sp Spanner[T]) bool {
	_, j := s.firstTouch(cmpf, sp)
	return j != -1
}

func (s *Set[T]) Add(cmpf Cmpf[T], sp Spanner[T]) {
	s.sorted = false

	if len(s.buckets) < 1 {
		s.buckets = append(s.buckets, bucket[T]{
			full: Span[T]{sp.Left(), sp.Right()},
		})
	}

	i, j := s.firstTouch(cmpf, sp)
	var newbucketsize int
	if j >= 0 {
		s.buckets[i].Add(cmpf, sp)
		newbucketsize = len(s.buckets[i].members)
	} else if i == -1 {
		s.buckets[0].Add(cmpf, sp)
		newbucketsize = len(s.buckets[0].members)
	} else {
		s.buckets[len(s.buckets)-1].Add(cmpf, sp)
		newbucketsize = len(s.buckets[len(s.buckets)-1].members)
	}

	if newbucketsize > s.largestBucketSize {
		s.largestBucketSize = newbucketsize
	}

	if s.NeedsResize() {
		s.Resize(cmpf)
	}
}

func (s *Set[T]) NeedsResize() bool {
	target := len(s.buckets) + 20
	return s.largestBucketSize > target
}

func (s *Set[T]) All(cmpf Cmpf[T]) iter.Seq[Spanner[T]] {
	return func(yield func(Spanner[T]) bool) {
		if !s.sorted {
			s.Sort(cmpf)
		}

		for i, _ := range s.buckets {
			b := &s.buckets[i]
			if !b.sorted {
				b.Sort(cmpf)
			}
			for _, sp := range b.members {
				if !yield(sp) {
					return
				}
			}
		}
	}
}

func (s *Set[T]) Resize(cmpf Cmpf[T]) {
	count := 0
	for _, b := range s.buckets {
		count += len(b.members)
	}
	target := int(math.Ceil(math.Sqrt(float64(count))))

	s.largestBucketSize = 0
	newBuckets := make([]bucket[T], 0, target)

	spans := s.All(cmpf)

	i := 0
	for sp := range spans {
		if i % target == 0 {
			b := bucket[T]{}
			b.full.Start = sp.Left()
			b.full.End = sp.Right()
			newBuckets = append(newBuckets, b)
		}

		b := &newBuckets[len(newBuckets) - 1]
		b.Add(cmpf, sp)
		if len(b.members) > s.largestBucketSize {
			s.largestBucketSize = len(b.members)
		}

		i++
	}

	s.buckets = newBuckets
	s.sorted = false
}
