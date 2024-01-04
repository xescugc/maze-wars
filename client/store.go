package client

import "github.com/xescugc/maze-wars/store"

type Store struct {
	*store.Store

	Users *UserStore
}

func NewStore(ss *store.Store, us *UserStore) *Store {
	return &Store{
		Store: ss,
		Users: us,
	}
}
