package main

import (
    "log"
    "time"
)

var highspeedWeights = map[bool]int{true: 60, false: 40}

func QueryAndCancel(q *Query) {
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        pairs, err = q.QueryOrders([]int{0, 1}, false)
    } else {
        pairs, err = q.QueryOrders([]int{0, 1}, true)
    }

    if err != nil || len(pairs) == 0 {
        log.Println("No orders found or error occurred")
        return
    }

    pair := RandomFromList(pairs)
    err = q.CancelOrder(pair[0], q.UID)
    if err != nil {
        log.Printf("Error cancelling order: %v", err)
        return
    }

    log.Printf("%s queried and canceled", pair[0])
}

func QueryAndCollect(q *Query) {
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        pairs, err = q.QueryOrders([]int{1}, false)
    } else {
        pairs, err = q.QueryOrders([]int{1}, true)
    }

    if err != nil || len(pairs) == 0 {
        log.Println("No orders found or error occurred")
        return
    }

    pair := RandomFromList(pairs)
    err = q.CollectTicket(pair[0])
    if err != nil {
        log.Printf("Error collecting ticket: %v", err)
        return
    }

    log.Printf("%s queried and collected", pair[0])
}

func QueryAndPreserve(q *Query) {
    var start, end string
    var tripIDs []string
    var err error

    highSpeed := RandomFromWeighted(highspeedWeights)
    if highSpeed {
        start, end = "Shang Hai", "Su Zhou"
        tripIDs, err = q.QueryHighSpeedTicket([2]string{start, end}, time.Now())
    } else {
        start, end = "Shang Hai", "Nan Jing"
        tripIDs, err = q.QueryNormalTicket([2]string{start, end}, time.Now())
    }

    if err != nil {
        log.Printf("Error querying tickets: %v", err)
        return
    }

    err = q.Preserve(start, end, tripIDs, highSpeed)
    if err != nil {
        log.Printf("Error preserving ticket: %v", err)
        return
    }

    log.Printf("Ticket preserved from %s to %s", start, end)
}

func QueryAndConsign(q *Query) {
    var list []map[string]interface{}
    var err error

    if RandomFromWeighted(highspeedWeights) {
        list, err = q.QueryOrdersAllInfo(false)
    } else {
        list, err = q.QueryOrdersAllInfo(true)
    }

    if err != nil || len(list) == 0 {
        log.Println("No orders found or error occurred")
        return
    }

    res := RandomFromList(list)
    err = q.PutConsign(res)
    if err != nil {
        log.Printf("Error putting consign: %v", err)
        return
    }

    log.Printf("%s queried and put consign", res["orderId"])
}

func QueryAndPay(q *Query) {
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        pairs, err = q.QueryOrders([]int{0, 1}, false)
    } else {
        pairs, err = q.QueryOrders([]int{0, 1}, true)
    }

    if err != nil || len(pairs) == 0 {
        log.Println("No orders found or error occurred")
        return
    }

    pair := RandomFromList(pairs)
    err = q.PayOrder(pair[0], pair[1])
    if err != nil {
        log.Printf("Error paying order: %v", err)
        return
    }

    log.Printf("%s queried and paid", pair[0])
}

func QueryAndRebook(q *Query) {
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        pairs, err = q.QueryOrders([]int{0, 1}, false)
    } else {
        pairs, err = q.QueryOrders([]int{0, 1}, true)
    }

    if err != nil || len(pairs) == 0 {
        log.Println("No orders found or error occurred")
        return
    }

    pair := RandomFromList(pairs)
    err = q.CancelOrder(pair[0], q.UID)
    if err != nil {
        log.Printf("Error cancelling order: %v", err)
        return
    }

    err = q.RebookTicket(pair[0], pair[1], pair[1])
    if err != nil {
        log.Printf("Error rebooking ticket: %v", err)
        return
    }

    log.Printf("%s queried and rebooked", pair[0])
}