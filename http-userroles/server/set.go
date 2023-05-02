package main

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](keys ...T) Set[T] {
	s := make(Set[T], len(keys))
	for _, k := range keys {
		s.Add(k)
	}
	return s
}

func (s Set[T]) Add(k T) {
	s[k] = struct{}{}
}

func (s Set[T]) Remove(k T) {
	delete(s, k)
}

func (s Set[T]) Contains(k T) bool {
	_, ok := s[k]
	return ok
}

func (s Set[T]) Difference(other Set[T]) Set[T] {
	diff := NewSet[T]()
	for k := range s {
		if !other.Contains(k) {
			diff.Add(k)
		}
	}
	return diff
}

func (s Set[T]) Equals(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}

func (s Set[T]) Elements() []T {
	elems := make([]T, 0, len(s))
	for k := range s {
		elems = append(elems, k)
	}
	return elems
}
