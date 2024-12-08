package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	sdk "github.com/kislerdm/neon-sdk-go"
)

func main() {
	token := os.Getenv("NEON_API_KEY")
	if token == "" {
		log.Fatal("env variable NEON_API_KEY must be set")
	}

	client, err := sdk.NewClient(sdk.Config{Key: token})
	if err != nil {
		log.Printf("could not initialize SDK: %v\n", err)
	}

	resp, err := client.CreateProject(sdk.ProjectCreateRequest{})
	if err != nil {
		log.Printf("could not create the project: %v\n", err)
	}

	connectionURI := resp.ConnectionURIs[0].ConnectionURI

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connectionURI)
	if err != nil {
		log.Printf("could not connect to database: %v\n", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	r, _ := conn.Query(ctx, "select now() at time zone 'utc';")
	defer func() { r.Close() }()
	for r.Next() {
		var now time.Time
		if e := r.Scan(&now); e != nil {
			log.Printf("could not scan row: %v\n", e)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "current UTC timestamp from database: %v\n", now.String())
		}
	}
}
