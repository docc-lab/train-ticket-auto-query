package main

import (
    "log"
    "time"

    "github.com/docc-lab/train-ticket-auto-query/tt-concurrent-load-generator"
)

func QueryTravelLeft(q *main.Query) {
    date := time.Now().Format("2006-01-02")
    var start, end string
    var err error
    var tripIDs []string

    highSpeed := false
    if highSpeed {
        start, end = "Shang Hai", "Su Zhou"
        tripIDs, err = q.QueryHighSpeedTicket([2]string{start, end}, date)
    } else {
        start, end = "Shang Hai", "Nan Jing"
        tripIDs, err = q.QueryNormalTicket([2]string{start, end}, date)
    }

    if err != nil {
        log.Printf("Error querying travel left: %v", err)
        return
    }

    log.Printf("Found %d trips from %s to %s on %s", len(tripIDs), start, end, date)
}