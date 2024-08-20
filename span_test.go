package span

import (
	"testing"
	"fmt"
	"cmp"
)

func TestSet(t *testing.T) {
	c := cmp.Compare[int]
	var s Set[int]
	s.Add(c, Span[int]{0, 1})
	s.Add(c, Span[int]{-5, 22})
	s.Add(c, Span[int]{3, 4})
	fmt.Println(s)
}
