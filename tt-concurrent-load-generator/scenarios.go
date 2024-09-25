package main

import (
    "fmt"
    "log"
)

var highspeedWeights = map[bool]int{true: 60, false: 40}

func QueryAndCancel(q *Query) {
    log.Println("Starting QueryAndCancel operation")
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        log.Println("Querying high-speed orders")
        pairs, err = q.QueryOrders([]int{0, 1}, false)
    } else {
        log.Println("Querying normal orders")
        pairs, err = q.QueryOrders([]int{0, 1}, true)
    }

    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    if len(pairs) == 0 {
        log.Println("No orders found for cancellation")
        return
    }

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    log.Printf("Selected order %s for cancellation", orderID)

    err = q.CancelOrder(orderID, q.UID)
    if err != nil {
        log.Printf("Error cancelling order %s: %v", orderID, err)
        return
    }

    log.Printf("Order %s successfully queried and canceled", orderID)
}

func QueryAndCollect(q *Query) {
    log.Println("Starting QueryAndCollect operation")
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        log.Println("Querying high-speed orders for collection")
        pairs, err = q.QueryOrders([]int{1}, false)
    } else {
        log.Println("Querying normal orders for collection")
        pairs, err = q.QueryOrders([]int{1}, true)
    }

    if err != nil {
        log.Printf("Error querying orders for collection: %v", err)
        return
    }

    if len(pairs) == 0 {
        log.Println("No orders found for collection")
        return
    }

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    log.Printf("Selected order %s for collection", orderID)

    err = q.CollectTicket(orderID)
    if err != nil {
        log.Printf("Error collecting ticket for order %s: %v", orderID, err)
        return
    }

    log.Printf("Order %s successfully queried and collected", orderID)
}

func QueryAndPreserve(q *Query) error {
    var start, end string
    var tripIDs []string
    var tripDate string
    var err error

    highSpeed := RandomBoolean()
    if highSpeed {
        start, end = "Shang Hai", "Su Zhou"
        tripIDs, tripDate, err = q.QueryHighSpeedTicket([2]string{start, end}, BaseDate)
    } else {
        start, end = "Shang Hai", "Nan Jing"
        tripIDs, tripDate, err = q.QueryNormalTicket([2]string{start, end}, BaseDate)
    }

    if err != nil {
        return fmt.Errorf("error querying tickets: %v", err)
    }

    if len(tripIDs) == 0 {
        return fmt.Errorf("no trips available from %s to %s on %s", start, end, BaseDate.Format("2006-01-02"))
    }

    err = q.Preserve(start, end, tripIDs, highSpeed, tripDate)
    if err != nil {
        return fmt.Errorf("error preserving ticket: %v", err)
    }

    log.Printf("Ticket preserved from %s to %s for %s", start, end, tripDate)
    return nil
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

    if err != nil {
        log.Printf("Error querying orders: %v", err)
        return
    }

    if len(pairs) == 0 {
        log.Println("No orders found")
        return
    }

    pair := RandomFromList(pairs).([2]string)
    orderID, tripID := pair[0], pair[1]

    err = q.PayOrder(orderID, tripID)
    if err != nil {
        log.Printf("Error paying for order: %v", err)
        return
    }

    log.Printf("Order %s queried and paid", orderID)
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

    err = q.RebookTicket(pair[0], pair[1], pair[1], BaseDate)
    if err != nil {
        log.Printf("Error rebooking ticket: %v", err)
        return
    }

    log.Printf("%s queried and rebooked", pair[0])
}

func QueryAndExecute(q *Query) {
    log.Println("Starting QueryAndExecute operation")
    var pairs [][2]string
    var err error

    if RandomFromWeighted(highspeedWeights) {
        log.Println("Querying high-speed orders for execution")
        pairs, err = q.QueryOrders([]int{2}, false)
    } else {
        log.Println("Querying normal orders for execution")
        pairs, err = q.QueryOrders([]int{2}, true)
    }

    if err != nil {
        log.Printf("Error querying orders for execution: %v", err)
        return
    }

    if len(pairs) == 0 {
        log.Println("No orders found for execution")
        return
    }

    pair := RandomFromList(pairs).([2]string)
    orderID := pair[0]

    log.Printf("Selected order %s for execution", orderID)

    err = q.EnterStation(orderID)
    if err != nil {
        log.Printf("Error entering station for order %s: %v", orderID, err)
        return
    }

    log.Printf("Order %s successfully queried and executed (entered station)", orderID)
}