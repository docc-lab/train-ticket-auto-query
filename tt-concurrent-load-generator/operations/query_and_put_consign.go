package operations

import (
    "log"
    "time"

    ".."
)

func QueryAndPutConsign(q *main.Query) {
    pairs, err := q.QueryOrdersAllInfo(false)
    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    pairs2, err := q.QueryOrdersAllInfo(true)
    if err != nil {
        log.Printf("Error querying other orders: %v", err)
        return
    }

    allPairs := append(pairs, pairs2...)
    if len(allPairs) == 0 {
        log.Println("No orders found")
        return
    }

    pair := main.RandomFromList(allPairs).(map[string]interface{})

    pair["targetDate"] = time.Now().Format("2006-01-02 15:04:05")

    err = q.PutConsign(pair)
    if err != nil {
        log.Printf("Error putting consign: %v", err)
        return
    }

    log.Printf("Order %s queried and put consign", pair["orderId"])
}