package span

import (
	"sort"
	"math"
	"iter"
)

type Pos interface {
	Cmp(Pos) int
}

type Span interface {
	Left() Pos
	Right() Pos
}

type Pos2 interface {
	Cmp(Pos2) int
}

func Min(ps ...Pos) Pos {
	if len(ps) < 1 {
		return nil
	}

	out := ps[0]
	for _, p := range ps[1:] {
		if p.Cmp(out) < 0 {
			out = p
		}
	}

	return out
}

func Max(ps ...Pos) Pos {
	if len(ps) < 1 {
		return nil
	}

	out := ps[0]
	for _, p := range ps[1:] {
		if p.Cmp(out) > 0 {
			out = p
		}
	}

	return out
}

func Touching(s1, s2 Span) bool {
	right := Min(s1.Right(), s2.Right())
	left := Max(s1.Left(), s2.Left())
	return left.Cmp(right) <= 0
}

func Overlapping(s1, s2 Span) bool {
	right := Min(s1.Right(), s2.Right())
	left := Max(s1.Left(), s2.Left())
	return left.Cmp(right) < 0
}

type span struct {
	left Pos
	right Pos
}

func (s span) Left() Pos {
	return s.left
}

func (s span) Right() Pos {
	return s.right
}

func Union(s1, s2 Span) Span {
	if !Touching(s1, s2) {
		return nil
	}
	return span{Min(s1.Left(), s2.Left()), Max(s1.Right(), s2.Right())}
}

func Intersect(s1, s2 Span) Span {
	if !Overlapping(s1, s2) {
		return nil
	}
	return span{Max(s1.Left(), s2.Left()), Min(s1.Right(), s2.Right())}
}

func Range(ss ...Span) Span {
	if len(ss) < 1 {
		return nil
	}
	lmin := ss[0].Left()
	rmax := ss[0].Right()
	for _, s := range ss {
		if s.Left().Cmp(lmin) < 0 {
			lmin = s.Left()
		}
		if s.Right().Cmp(rmax) > 0 {
			rmax = s.Right()
		}
	}

	return span{lmin, rmax}
}

func SortSpans(ss []Span) {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Left().Cmp(ss[j].Left()) < 0
	})
}

type bucket struct {
	sorted bool
	full span
	members []Span
}

func (b bucket) Left() Pos {
	return b.full.Left()
}

func (b bucket) Right() Pos {
	return b.full.Right()
}

func sortBuckets(bs []bucket) {
	sort.Slice(bs, func(i, j int) bool {
		return bs[i].full.Left().Cmp(bs[j].full.Left()) < 0
	})
}

func (b *bucket) Add(sp Span) {
	b.sorted = false
	b.members = append(b.members, sp)

	if b.full.left == nil {
		b.full.left = sp.Left()
	} else {
		b.full.left = Min(b.full.Left(), sp.Left())
	}

	if b.full.right == nil {
		b.full.right = sp.Right()
	} else {
		b.full.right = Max(b.full.Right(), sp.Right())
	}
}

func (b *bucket) Sort() {
	b.sorted = true
	SortSpans(b.members)
}

type Set struct {
	sorted bool
	largestBucketSize int
	buckets []bucket
}

func (s *Set) Sort() {
	s.sorted = true
	sortBuckets(s.buckets)
}

func (s *Set) firstTouch(sp Span) (bi, si int) {
	for i, b := range s.buckets {
		if sp.Right().Cmp(b.Left()) < 0 {
			return -1, -1
		}
		if Touching(sp, b) {
			for j, m := range b.members {
				if Touching(sp, m) {
					return i, j
				}
			}
		}
	}
	return len(s.buckets), -1
}

func (s *Set) Touching(sp Span) bool {
	_, j := s.firstTouch(sp)
	return j != -1
}

func (s *Set) Add(sp Span) {
	s.sorted = false

	if len(s.buckets) < 1 {
		s.buckets = append(s.buckets, bucket{
			full: span{sp.Left(), sp.Right()},
		})
	}

	i, j := s.firstTouch(sp)
	var newbucketsize int
	if j >= 0 {
		s.buckets[i].Add(sp)
		newbucketsize = len(s.buckets[i].members)
	} else if i == -1 {
		s.buckets[0].Add(sp)
		newbucketsize = len(s.buckets[0].members)
	} else {
		s.buckets[len(s.buckets)-1].Add(sp)
		newbucketsize = len(s.buckets[len(s.buckets)-1].members)
	}

	if newbucketsize > s.largestBucketSize {
		s.largestBucketSize = newbucketsize
	}

	if s.NeedsResize() {
		s.Resize()
	}
}

func (s *Set) NeedsResize() bool {
	target := len(s.buckets) + 20
	return s.largestBucketSize > target
}

func (s *Set) All() iter.Seq[Span] {
	return func(yield func(Span) bool) {
		if !s.sorted {
			s.Sort()
		}

		for i, _ := range s.buckets {
			b := &s.buckets[i]
			if !b.sorted {
				b.Sort()
			}
			for _, sp := range b.members {
				if !yield(sp) {
					return
				}
			}
		}
	}
}

func (s *Set) Resize() {
	count := 0
	for _, b := range s.buckets {
		count += len(b.members)
	}
	target := int(math.Ceil(math.Sqrt(float64(count))))

	s.largestBucketSize = 0
	newBuckets := make([]bucket, 0, target)

	spans := s.All()

	i := 0
	for sp := range spans {
		if i % target == 0 {
			b := bucket{}
			b.full.left = sp.Left()
			b.full.right = sp.Right()
			newBuckets = append(newBuckets, b)
		}

		b := &newBuckets[len(newBuckets) - 1]
		b.Add(sp)
		if len(b.members) > s.largestBucketSize {
			s.largestBucketSize = len(b.members)
		}

		i++
	}

	s.buckets = newBuckets
	s.sorted = false
}
