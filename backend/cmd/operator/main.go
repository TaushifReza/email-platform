// Command operator bootstraps the first platform operator. There is no
// operator self-signup; Cloudflare Access supplies authentication at runtime
// and this command only maintains the operator allowlist.
package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/TaushifReza/email-platform/backend/internal/platform/config"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platformadmin"
)

func main() {
	if err := run(); err != nil {
		log.Println("operator:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 || os.Args[1] != "create" {
		return errors.New("usage: operator create <email>")
	}
	email := platformadmin.NormalizeEmail(os.Args[2])
	if email == "" {
		return errors.New("email is required")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := db.NewWithSystem(ctx, cfg.DatabaseURL, cfg.SystemDatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	id, err := platformadmin.NewID()
	if err != nil {
		return err
	}
	created, err := platformadmin.NewRepository(pool).CreateOperator(ctx, id, email)
	if err != nil {
		return err
	}
	if !created {
		log.Printf("operator %s already exists", email)
		return nil
	}
	log.Printf("created operator %s", email)
	return nil
}
