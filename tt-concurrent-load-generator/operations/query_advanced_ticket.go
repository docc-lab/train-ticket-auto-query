package main

import (
    "fmt"
    "log"
    "time"

    "github.com/docc-lab/train-ticket-auto-query/tt-concurrent-load-generator"
)

func QueryAdvancedTicket(q *queries.Query) {
    placePairs := [][2]string{
        {"Shang Hai", "Su Zhou"},
        {"Su Zhou", "Shang Hai"},
        {"Nan Jing", "Shang Hai"},
    }
    ticketType := "quickest"

    placePair := queries.RandomFromList(placePairs).([2]string)
    date := time.Now().Format("2006-01-02")

    log.Printf("Searching %s route between %s to %s", ticketType, placePair[0], placePair[1])

    tripIDs, err := q.QueryAdvancedTicket(placePair, date, ticketType)
    if err != nil {
        log.Printf("Error querying advanced ticket: %v", err)
        return
    }

    log.Printf("Found %d routes", len(tripIDs))
}