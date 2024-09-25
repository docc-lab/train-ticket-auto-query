package operations

import (
    "log"
    "time"

    ".."
)

func QueryAndRebook(q *main.Query) {
    pairs, err := q.QueryOrders([]int{1}, false)
    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    if len(pairs) == 0 {
        log.Println("No orders found")
        return
    }

    newTripID := "D1345"
    newDate := time.Now().Format("2006-01-02")
    newSeatType := "3"

    for _, pair := range pairs {
        err := q.RebookTicket(pair[0], pair[1], newTripID, newDate, newSeatType)
        if err != nil {
            log.Printf("Error rebooking ticket: %v", err)
        } else {
            log.Printf("Order %s rebooked successfully", pair[0])
        }
    }
}