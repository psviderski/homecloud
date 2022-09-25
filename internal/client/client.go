package client

import (
	"fmt"
)

type Client struct {
	Store *Store
}

func NewClient(storePath string) (*Client, error) {
	s, err := LoadOrCreate(storePath)
	if err != nil {
		return nil, fmt.Errorf("cannot load store: %w", err)
	}
	return &Client{
		Store: s,
	}, nil
}
