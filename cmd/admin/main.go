package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"qalbum-server/pkg/dao"
)

type Config struct {
	DBPath string `yaml:"db_path"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	db, err := sqlx.Connect("sqlite3", "data/db/qalbum.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userDAO := dao.NewUserDAO(db)

	switch command {
	case "user":
		handleUserCommand(os.Args[2:], userDAO)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("QAlbum Admin Tool")
	fmt.Println("Usage: qalbum-admin <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  user add --openid <openid>  Add a new user")
	fmt.Println("  user list                    List all users")
}

func handleUserCommand(args []string, userDAO *dao.UserDAO) {
	if len(args) < 1 {
		fmt.Println("Usage: qalbum-admin user <action> [options]")
		fmt.Println("Actions:")
		fmt.Println("  add --openid <openid>  Add a new user")
		fmt.Println("  list                    List all users")
		return
	}

	action := args[0]

	switch action {
	case "add":
		addUser(args[1:], userDAO)
	case "list":
		listUsers(userDAO)
	default:
		fmt.Println("Unknown action:", action)
	}
}

func addUser(args []string, userDAO *dao.UserDAO) {
	var openid string
	flagSet := flag.NewFlagSet("add", flag.ExitOnError)
	flagSet.StringVar(&openid, "openid", "", "User OpenID")
	flagSet.Parse(args)

	if openid == "" {
		fmt.Println("Error: --openid is required")
		return
	}

	user, err := userDAO.Create(openid)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}

	fmt.Printf("User created successfully:\n")
	fmt.Printf("  ID: %d\n", user.ID)
	fmt.Printf("  OpenID: %s\n", user.OpenID)
	fmt.Printf("  CreatedAt: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
}

func listUsers(userDAO *dao.UserDAO) {
	users, err := userDAO.List()
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		return
	}

	fmt.Printf("Total users: %d\n", len(users))
	fmt.Println()
	for _, user := range users {
		fmt.Printf("ID: %d, OpenID: %s, CreatedAt: %s\n",
			user.ID,
			user.OpenID,
			user.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
