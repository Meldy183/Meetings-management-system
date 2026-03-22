package tools

import (
	"encoding/json"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
	"meetings-mcp/client"
)

func Register(server *mcp.Server, c *client.Client) error {
	if err := registerPeopleTools(server, c); err != nil {
		return fmt.Errorf("people tools: %w", err)
	}
	if err := registerMeetingTools(server, c); err != nil {
		return fmt.Errorf("meeting tools: %w", err)
	}
	return nil
}

func jsonResponse(v interface{}) (*mcp.ToolResponse, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	return mcp.NewToolResponse(mcp.NewTextContent(string(b))), nil
}
