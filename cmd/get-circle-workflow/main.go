package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/tmc/circleci"
)

var (
	flagVerbose    = flag.Bool("v", false, "verbose")
	flagWorkflowID = flag.String("workflow-id", "", "workflow id")
)

func main() {
	flag.Parse()
	if err := run(*flagWorkflowID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(id string) error {
	opts := []circleci.ClientOption{
		circleci.WithSessionToken(os.Getenv("CIRCLECI_SESSION_TOKEN")),
	}
	if *flagVerbose {
		opts = append(opts, circleci.WithDebugLogging())
	}
	c, err := circleci.NewClient(opts...)
	if err != nil {
		return err
	}
	workflow, err := c.GetWorkflow(id)
	if err != nil {
		return errors.Wrap(err, "issue getting workflow")
	}
	return json.NewEncoder(os.Stdout).Encode(workflow)
}
