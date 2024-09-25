package main

import (
    "log"
    "os"

    "./scenarios"
)

func main() {
    // Set up logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    url := "http://192.168.188.42:8080"

    // Initialize the Query struct
    q := NewQuery(url)

    // Login to train-ticket and store the cookies
    err := q.Login("username", "password") // You need to provide actual username and password
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    // Execute scenario on current user
    scenarios.QueryAndPreserve(q)

    // Commented out scenarios - uncomment to use
    // scenarios.QueryAndPay(q)
    // scenarios.QueryAndCancel(q)
    // scenarios.QueryAndCollect(q)
    // scenarios.QueryAndExecute(q)
    // scenarios.QueryAndConsign(q)
    // scenarios.QueryAndRebook(q)

    // Commented out direct query executions - uncomment to use
    // _, err = q.QueryHighSpeedTicket([2]string{"Shang Hai", "Su Zhou"}, time.Now())
    // _, err = q.QueryNormalTicket([2]string{"Shang Hai", "Nan Jing"}, time.Now())
    // _, err = q.QueryAssurances()
    // _, err = q.QueryFood([2]string{"Shang Hai", "Su Zhou"}, "D1345")
    // _, err = q.QueryContacts()
    // _, err = q.QueryOrders([]int{0, 1}, false)
    // _, err = q.QueryOrdersAllInfo(false)
    // err = q.QueryRoute("some-route-id")
    // err = q.PutConsign(map[string]interface{}{"orderId": "some-order-id"})
    // err = q.PayOrder("some-order-id", "some-trip-id")
    // err = q.CancelOrder("some-order-id", q.UID)
    // err = q.CollectTicket("some-order-id")

    if err != nil {
        log.Printf("An error occurred: %v", err)
    }
}