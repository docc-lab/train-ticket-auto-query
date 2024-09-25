package main

import (
    "log"
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

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    err = q.CancelOrder(orderID, q.UID)
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

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    err = q.CancelOrder(orderID, q.UID)
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

    highSpeed := RandomBoolean()
    if highSpeed {
        start, end = "Shang Hai", "Su Zhou"
        tripIDs, err = q.QueryHighSpeedTicket([2]string{start, end}, BaseDate)
    } else {
        start, end = "Shang Hai", "Nan Jing"
        tripIDs, err = q.QueryNormalTicket([2]string{start, end}, BaseDate)
    }

    if err != nil {
        log.Printf("Error querying tickets: %v", err)
        return
    }

    if len(tripIDs) == 0 {
        log.Printf("No trips available from %s to %s on %s", start, end, BaseDate.Format("2006-01-02"))
        return
    }

    err = q.Preserve(start, end, tripIDs, highSpeed)
    if err != nil {
        log.Printf("Error preserving ticket: %v", err)
        return
    }

    log.Printf("Ticket preserved from %s to %s for %s", start, end, BaseDate.Format("2006-01-02"))
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

    res := RandomFromList(list).([2]string)
    orderID := res[0]

    err = q.CancelOrder(orderID, q.UID)
    if err != nil {
        log.Printf("Error putting consign: %v", err)
        return
    }

    log.Printf("%s queried and put consign", res[0])
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

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    err = q.CancelOrder(orderID, q.UID)
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

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    err = q.CancelOrder(orderID, q.UID)
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
