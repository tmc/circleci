package circleci_test

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/circleci"
)

func ExampleNewClient() {
	c, err := circleci.NewClient()
	//c, err := circleci.NewClient(circleci.WithDebugLogging())
	if err != nil {
		fmt.Println(err)
	}
	foo, err := c.GetWorkflow(context.Background(), "")
	fmt.Printf("%T")
	// output:
	// Foo
}

func ExampleNewClient_authenticated() {
	token := os.Getenv("CIRCLECI_TOKEN")
	client, err := circleci.NewClient(circleci.WithToken(token), circleci.WithDebugLogging())

	_, _ = client, err
}

func ExampleNewClient_session_authenticated() {
	token := os.Getenv("CIRCLECI_SESSION_TOKEN")
	client, err := circleci.NewClient(circleci.WithSessionToken(token), circleci.WithDebugLogging())

	_, _ = client, err
}

func ExampleWithDebugLogging() {
	client, err := circleci.NewClient(circleci.WithDebugLogging())

	_, _ = client, err
}
