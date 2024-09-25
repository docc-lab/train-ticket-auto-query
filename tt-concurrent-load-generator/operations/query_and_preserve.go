package operations

import (
    "log"
    "time"

    ".."
)

func QueryAndPreserve(q *main.Query) {
    var start, end string
    var tripIDs []string
    var err error

    date := time.Now().Format("2006-01-02")

    highSpeed := main.RandomBoolean()
    if highSpeed {
        start, end = "Shang Hai", "Su Zhou"
        tripIDs, err = q.QueryHighSpeedTicket([2]string{start, end}, date)
    } else {
        start, end = "Shang Hai", "Nan Jing"
        tripIDs, err = q.QueryNormalTicket([2]string{start, end}, date)
    }

    if err != nil {
        log.Printf("Error querying tickets: %v", err)
        return
    }

    if len(tripIDs) == 0 {
        log.Println("No trips available")
        return
    }

    err = q.Preserve(start, end, tripIDs, highSpeed)
    if err != nil {
        log.Printf("Error preserving ticket: %v", err)
        return
    }

    log.Printf("Ticket preserved from %s to %s", start, end)
}