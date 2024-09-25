package operations

import (
    "log"

    ".."
)

func QueryOrderAndPay(q *main.Query) {
    pairs, err := q.QueryOrders([]int{0, 1}, false)
    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    pairs2, err := q.QueryOrders([]int{0, 1}, true)
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
    orderID, tripID := pair[0], pair[1]

    err = q.PayOrder(orderID, tripID)
    if err != nil {
        log.Printf("Error paying for order: %v", err)
        return
    }

    log.Printf("Order %s queried and paid", orderID)
}