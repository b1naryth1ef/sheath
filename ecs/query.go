package ecs

type Query[T any] struct {
	u    *Universe
	refs []*T
}

func Exec[T any](u *Universe) *Query[T] {
	refs := make([]*T, 0)

	for _, data := range u.entities {
		var item T
		if data.Exec(&item) {
			refs = append(refs, &item)
		}
	}

	return &Query[T]{u: u, refs: refs}
}

func (q *Query[T]) Iter() func(yield func(*T) bool) {
	return func(yield func(*T) bool) {
		for _, item := range q.refs {
			if !yield(item) {
				return
			}
		}
	}
}

func (q *Query[T]) Len() int {
	return len(q.refs)
}
