package operations

import (
    "log"

    ".."
)

func QueryAndCollectTicket(q *queries.Query) {
    pairs, err := q.QueryOrders([]int{1}, false)
    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    pairs2, err := q.QueryOrders([]int{1}, true)
    if err != nil {
        log.Printf("Error querying other orders: %v", err)
        return
    }

    allPairs := append(pairs, pairs2...)
    if len(allPairs) == 0 {
        log.Println("No orders found")
        return
    }

    pair := queries.RandomFromList(allPairs).([2]string)
    orderID := pair[0]

    err = q.CollectTicket(orderID)
    if err != nil {
        log.Printf("Error collecting ticket: %v", err)
        return
    }

    log.Printf("Order %s queried and collected", orderID)
}