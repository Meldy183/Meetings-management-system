package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mcp "github.com/metoro-io/mcp-golang"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"

	"meetings-mcp/client"
	"meetings-mcp/tools"
)

func main() {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	addr := os.Getenv("MCP_ADDR")
	if addr == "" {
		addr = ":3000"
	}

	c := client.New(backendURL)

	transport := mcphttp.NewHTTPTransport("/mcp").WithAddr(addr)
	server := mcp.NewServer(transport,
		mcp.WithName("meetings-mcp"),
		mcp.WithVersion("1.0.0"),
		mcp.WithInstructions("MCP server for the Meetings Management System. Use these tools to manage meetings, people, agenda items, and speakers. Workflow: 1) create_meeting, 2) add people via add_person_to_meeting, 3) set_meeting_chairperson, 4) add agenda items via add_agenda_item, 5) export when status is complete."),
	)

	if err := tools.Register(server, c); err != nil {
		log.Fatalf("failed to register tools: %v", err)
	}

	if err := server.Serve(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	log.Printf("MCP server listening on %s/mcp (backend: %s)", addr, backendURL)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down")
}
