package span

import (
	"testing"
	"fmt"
)

func TestSet(t *testing.T) {
	s := NewSet[int]()
	s.Add(Span[int]{0, 1})
	s.Add(Span[int]{-5, 22})
	s.Add(Span[int]{3, 4})
	fmt.Println(s)
}

func TestSet2(t *testing.T) {
	set := NewOrderedSet[Span[int64], int64]([]Span[int64]{Span[int64]{0,10}, Span[int64]{20, 30}, Span[int64]{15,16}})
	fmt.Println(set.Touching(Span[int64]{3,5}))
	fmt.Println(set.Touching(Span[int64]{8, 12}))
	fmt.Println(set.Touching(Span[int64]{10,11}))
	fmt.Println(set.Touching(Span[int64]{18, 22}))
	fmt.Println(set.Touching(Span[int64]{18, 20}))
}
