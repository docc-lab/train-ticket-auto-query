package operations

import (
    "log"

    ".."
)

func QueryAndEnterStation(q *main.Query) {
    pairs, err := q.QueryOrders([]int{2}, false)
    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    pairs2, err := q.QueryOrders([]int{2}, true)
    if err != nil {
        log.Printf("Error querying other orders: %v", err)
        return
    }

    allPairs := append(pairs, pairs2...)
    if len(allPairs) == 0 {
        log.Println("No orders found")
        return
    }

    pair := main.RandomFromList(allPairs).([][2]string)[0]
    orderID := pair[0]

    err = q.EnterStation(orderID)
    if err != nil {
        log.Printf("Error entering station: %v", err)
        return
    }

    log.Printf("Order %s queried and entered station", orderID)
}