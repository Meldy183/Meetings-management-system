package main

import (
	"context"
	"fmt"
)

type Person struct {
	ID         int     `json:"id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Info       *string `json:"info,omitempty"`
}

type PersonCreateRequest struct {
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Info       *string `json:"info,omitempty"`
}


func (c *Client) ListPeople(ctx context.Context, q string) ([]Person, error) {
	var people []Person
	path := "/people" + buildQuery(map[string]string{"q": q})
	err := c.do(ctx, "GET", path, nil, &people)
	return people, err
}

func (c *Client) GetPerson(ctx context.Context, id int) (*Person, error) {
	var p Person
	err := c.do(ctx, "GET", fmt.Sprintf("/people/%d", id), nil, &p)
	return &p, err
}

func (c *Client) CreatePerson(ctx context.Context, req PersonCreateRequest) (*Person, error) {
	var p Person
	err := c.do(ctx, "POST", "/people", req, &p)
	return &p, err
}

func (c *Client) UpdatePerson(ctx context.Context, id int, req PersonCreateRequest) (*Person, error) {
	var p Person
	err := c.do(ctx, "PATCH", fmt.Sprintf("/people/%d", id), req, &p)
	return &p, err
}
