package main

import (
    "log"
    "os"
)

func main() {
    // Set up logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    url := "http://192.168.188.42:8080"
    log.Printf("Connecting to: %s", url)

    q := NewQuery(url)
    log.Println("Attempting to login...")
    err := q.Login("username", "password")
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    log.Println("Login successful")

    // Execute scenario on current user
    QueryAndPreserve(q)

    // Commented out scenarios - uncomment to use
    // QueryAndPay(q)
    // QueryAndCancel(q)
    // QueryAndCollect(q)
    // QueryAndExecute(q)
    // QueryAndConsign(q)
    // QueryAndRebook(q)

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