package common

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
)

type AnnotateEnumStack struct {
	Slice []*annotateast.AnnotateEnumState
}

func (s *AnnotateEnumStack) Push(val *annotateast.AnnotateEnumState) {
	s.Slice = append(s.Slice, val)
}

func (s *AnnotateEnumStack) Pop() (val *annotateast.AnnotateEnumState) {
	if len(s.Slice) == 0 {
		return nil
	}

	index := len(s.Slice) - 1
	val = s.Slice[index]
	s.Slice = s.Slice[:index]
	return val
}

func (s *AnnotateEnumStack) Len() int {
	return len(s.Slice)
}

func (s *AnnotateEnumStack) Peek() (val *annotateast.AnnotateEnumState) {
	if len(s.Slice) == 0 {
		return nil
	}

	return s.Slice[len(s.Slice)-1]
}
