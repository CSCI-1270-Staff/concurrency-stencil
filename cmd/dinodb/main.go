package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dinodb/pkg/config"
	"dinodb/pkg/join"
	"dinodb/pkg/list"
	"dinodb/pkg/pager"
	"dinodb/pkg/repl"

	"dinodb/pkg/database"

	"github.com/google/uuid"
)

// Default port 8335 (BEES).
const DEFAULT_PORT int = 8335

const LOG_FILE_NAME = "data/dinodb.log"

// [HASH/BTREE]
// // Listens for SIGINT or SIGTERM and calls table.CloseDB().
func setupCloseHandler(database *database.Database) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("closehandler invoked")
		database.Close()
		os.Exit(0)
	}()
}

// Start the database.
func main() {
	// Set up flags.
	var promptFlag = flag.Bool("c", true, "use prompt?")
	var projectFlag = flag.String("project", "", "choose project: [go,pager,hash,btree,join] (required)")

	// [HASH/BTREE]
	var dbFlag = flag.String("db", "data/", "DB folder")

	flag.Parse()

	// [HASH/BTREE]
	// Open the db.
	db, err := database.Open(*dbFlag)
	if err != nil {
		panic(err)
	}

	// [HASH/BTREE]
	// Setup close conditions.
	defer db.Close()
	setupCloseHandler(db)

	// Set up REPL resources.
	prompt := config.GetPrompt(*promptFlag)
	repls := make([]*repl.REPL, 0)

	server := false

	// Get the right REPLs.
	switch *projectFlag {
	case "go":
		l := list.NewList()
		repls = append(repls, list.ListRepl(l))

	// [PAGER]
	case "pager":
		pRepl, err := pager.PagerRepl()
		if err != nil {
			fmt.Println(err)
			return
		}
		repls = append(repls, pRepl)

	// [HASH/BTREE]
	case "hash", "btree":
		server = false
		repls = append(repls, database.DatabaseRepl(db))

	// [JOIN]
	case "join":
		server = false
		repls = append(repls, database.DatabaseRepl(db))
		repls = append(repls, join.JoinRepl(db))

	default:
		fmt.Println("must specify -project [go,pager,hash,btree,join]")
		return
	}

	// Combine the REPLs.
	r, err := repl.CombineRepls(repls)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start server if server (concurrency or recovery), else run REPL here.
	if server {
		// 	[CONCURRENCY]
		//ignore for now
		r.Run(uuid.New(), prompt, nil, nil)
	} else {
		r.Run(uuid.New(), prompt, nil, nil)
	}
}
