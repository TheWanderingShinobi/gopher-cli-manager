package main

import (
	"github.com/TheWanderingShinobi/gopher-cli-manager/internal/database"
	"github.com/TheWanderingShinobi/gopher-cli-manager/internal/tui"
)

func main() {
	db, err := database.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tui.StartTea(db)
}
