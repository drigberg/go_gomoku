
package db

import (
	"fmt"
	"log"
	"os"
	"database/sql"
	"go_gomoku/types"
)
var (
	db *sql.DB
)

func Init() {
	fmt.Println("Connecting to db...")

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	db.Exec() // insert sql from init.sql
}

func CreateUser(user *types.User) {
	if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
		fmt.Sprintf("Error incrementing tick: %q", err)
	}
}

func GetUser(userId string) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
			fmt.Sprintf("Error reading users: %q", err)
			return
	}

	defer rows.Close()
	for rows.Next() {
			var user types.User
			if err := rows.Scan(&user); err != nil {
				fmt.Sprintf("Error scanning users: %q", err)
					return
			}
			fmt.Sprintf("Read from db: %s\n", tick.String())
	}
}
