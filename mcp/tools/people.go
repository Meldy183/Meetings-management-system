package tools

import (
	"context"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
	"meetings-mcp/client"
)

func registerPeopleTools(server *mcp.Server, c *client.Client) error {
	// list_people
	type ListPeopleArgs struct {
		Query string `json:"query" jsonschema:"description=Optional search string. Split on whitespace — each word is matched as a substring of the full name. Empty string returns all people."`
	}
	if err := server.RegisterTool("list_people", "List all people or search by name using partial word matching",
		func(ctx context.Context, args ListPeopleArgs) (*mcp.ToolResponse, error) {
			people, err := c.ListPeople(ctx, args.Query)
			if err != nil {
				return nil, err
			}
			return jsonResponse(people)
		}); err != nil {
		return fmt.Errorf("list_people: %w", err)
	}

	// get_person
	type GetPersonArgs struct {
		ID int `json:"id" jsonschema:"required,description=Person ID"`
	}
	if err := server.RegisterTool("get_person", "Get a single person by ID",
		func(ctx context.Context, args GetPersonArgs) (*mcp.ToolResponse, error) {
			person, err := c.GetPerson(ctx, args.ID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(person)
		}); err != nil {
		return fmt.Errorf("get_person: %w", err)
	}

	// create_person
	type CreatePersonArgs struct {
		LastName   string  `json:"last_name" jsonschema:"required,description=Last name"`
		FirstName  string  `json:"first_name" jsonschema:"required,description=First name"`
		MiddleName *string `json:"middle_name" jsonschema:"description=Patronymic (optional)"`
		Info       *string `json:"info" jsonschema:"description=Role or position shown in exported documents (optional)"`
	}
	if err := server.RegisterTool("create_person", "Create a new person in the database. Returns 409 if a person with the same full name already exists.",
		func(ctx context.Context, args CreatePersonArgs) (*mcp.ToolResponse, error) {
			person, err := c.CreatePerson(ctx, client.PersonCreateRequest{
				LastName:   args.LastName,
				FirstName:  args.FirstName,
				MiddleName: args.MiddleName,
				Info:       args.Info,
			})
			if err != nil {
				return nil, err
			}
			return jsonResponse(person)
		}); err != nil {
		return fmt.Errorf("create_person: %w", err)
	}

	// update_person
	type UpdatePersonArgs struct {
		ID         int     `json:"id" jsonschema:"required,description=Person ID to update"`
		LastName   *string `json:"last_name" jsonschema:"description=New last name (omit to keep current)"`
		FirstName  *string `json:"first_name" jsonschema:"description=New first name (omit to keep current)"`
		MiddleName *string `json:"middle_name" jsonschema:"description=New patronymic (omit to keep current)"`
		Info       *string `json:"info" jsonschema:"description=New role or position (omit to keep current)"`
	}
	if err := server.RegisterTool("update_person", "Partially update a person — only the fields you provide are changed",
		func(ctx context.Context, args UpdatePersonArgs) (*mcp.ToolResponse, error) {
			person, err := c.UpdatePerson(ctx, args.ID, client.PersonUpdateRequest{
				LastName:   args.LastName,
				FirstName:  args.FirstName,
				MiddleName: args.MiddleName,
				Info:       args.Info,
			})
			if err != nil {
				return nil, err
			}
			return jsonResponse(person)
		}); err != nil {
		return fmt.Errorf("update_person: %w", err)
	}

	// sort_people
	type SortPeopleArgs struct {
		IDs []int `json:"ids" jsonschema:"required,description=Person IDs to sort alphabetically by last_name then first_name then middle_name"`
	}
	if err := server.RegisterTool("sort_people", "Return the given person IDs sorted alphabetically by last_name, first_name, middle_name",
		func(ctx context.Context, args SortPeopleArgs) (*mcp.ToolResponse, error) {
			sorted, err := c.SortPeople(ctx, args.IDs)
			if err != nil {
				return nil, err
			}
			return jsonResponse(client.SortPeopleResponse{IDs: sorted})
		}); err != nil {
		return fmt.Errorf("sort_people: %w", err)
	}

	return nil
}
