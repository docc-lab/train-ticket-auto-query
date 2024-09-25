package operations

import (
    "log"
    "time"

    ".."
)

func QueryTravelLeftParallel(q *main.Query) {
    date := time.Now().Format("2006-01-02")
    start, end := "Su Zhou", "Shang Hai"
    highSpeedPlacePair := [2]string{start, end}

    tripIDs, err := q.QueryHighSpeedTicketParallel(highSpeedPlacePair, date)
    if err != nil {
        log.Printf("Error querying travel left parallel: %v", err)
        return
    }

    log.Printf("Found %d trips from %s to %s on %s using parallel query", len(tripIDs), start, end, date)
}