package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/term"
	"whisper/backend/internal/config"
	"whisper/backend/internal/database"
	"whisper/backend/internal/repository"
	"whisper/backend/internal/service"
)

var validUsername = regexp.MustCompile(`^[A-Za-z0-9_]{3,50}$`)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Admin username (3-50 letters, numbers, or _): ")
	username, err := reader.ReadString('\n')
	if err != nil && len(username) == 0 {
		fatal("could not read username")
	}
	username = strings.TrimSpace(username)
	if !validUsername.MatchString(username) {
		fatal("username must be 3-50 letters, numbers, or underscores")
	}

	password, err := readPassword("Password (12-72 bytes): ")
	if err != nil {
		fatal("could not read password")
	}
	if len(password) < 12 || len(password) > 72 {
		fatal("password must be 12-72 bytes")
	}
	confirmation, err := readPassword("Confirm password: ")
	if err != nil || string(password) != string(confirmation) {
		fatal("passwords did not match")
	}

	cfg, err := config.Load()
	if err != nil {
		fatal(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fatal("could not connect to PostgreSQL: " + err.Error())
	}
	defer db.Close()
	if err := database.Migrate(ctx, db); err != nil {
		fatal("could not apply database migrations: " + err.Error())
	}
	passwordHash, err := service.HashPassword(string(password))
	if err != nil {
		fatal("could not hash password")
	}
	if err := repository.NewAdminRepository(db).Create(ctx, username, passwordHash); err != nil {
		fatal("could not create admin (the username may already exist): " + err.Error())
	}
	fmt.Printf("Admin %q created.\n", username)
}

func readPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return password, err
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}
