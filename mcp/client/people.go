package client

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

type PersonUpdateRequest struct {
	LastName   *string `json:"last_name,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
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
	var person Person
	err := c.do(ctx, "GET", fmt.Sprintf("/people/%d", id), nil, &person)
	return &person, err
}

func (c *Client) CreatePerson(ctx context.Context, req PersonCreateRequest) (*Person, error) {
	var person Person
	err := c.do(ctx, "POST", "/people", req, &person)
	return &person, err
}

func (c *Client) UpdatePerson(ctx context.Context, id int, req PersonUpdateRequest) (*Person, error) {
	var person Person
	err := c.do(ctx, "PATCH", fmt.Sprintf("/people/%d", id), req, &person)
	return &person, err
}

type SortPeopleRequest struct {
	IDs []int `json:"ids"`
}

type SortPeopleResponse struct {
	IDs []int `json:"ids"`
}

func (c *Client) SortPeople(ctx context.Context, ids []int) ([]int, error) {
	var resp SortPeopleResponse
	err := c.do(ctx, "POST", "/people/sort", SortPeopleRequest{IDs: ids}, &resp)
	return resp.IDs, err
}
