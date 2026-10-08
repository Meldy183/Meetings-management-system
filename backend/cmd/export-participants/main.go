// Command export-participants exports the entire registry to XLSX using read-only SQL.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"meetings-editor/internal/registryexport"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Export failed:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("export-participants", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("out", "participants-"+time.Now().Format("20060102-150405")+".xlsx", "new XLSX file to create (never overwritten)")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: DATABASE_URL=<PostgreSQL DSN> export-participants [-out participants.xlsx]")
		fmt.Fprintln(stderr, "Exports the entire registry. Only read-only database queries are executed.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if !strings.EqualFold(filepath.Ext(*output), ".xlsx") {
		return errors.New("output file must have the .xlsx extension")
	}
	if _, err := os.Lstat(*output); err == nil {
		return errors.New("output already exists: choose another -out path")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check output path: %w", err)
	}
	if dsn := os.Getenv("DATABASE_URL"); strings.TrimSpace(dsn) == "" {
		return errors.New("DATABASE_URL is required; no default database is selected")
	}

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	people, err := registryexport.Read(readCtx, os.Getenv("DATABASE_URL"))
	cancel()
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := registryexport.WriteXLSX(people, *output); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Exported %d participants to %s\n", len(people), *output)
	return err
}
