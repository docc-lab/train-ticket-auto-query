package main

import (
    "flag"
    "log"
    "os"
    "sync"
    "time"
    "math/rand"
)

const (
    ThreadCount = 4
    ScenariosPerThread = 10
)

func main() {
    // Set up logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    dateStr := flag.String("date", BaseDate.Format("2006-01-02"), "Initial base date for querying trips (format: YYYY-MM-DD)")
    flag.Parse()

    var err error
    BaseDate, err = time.Parse("2006-01-02", *dateStr)
    if err != nil {
        log.Fatalf("Invalid date format: %v", err)
    }

    url := "http://192.168.188.42:8080"
    log.Printf("Connecting to: %s", url)

    scenarios := []struct {
        name     string
        function func(*Query)
    }{
        {"QueryAndPreserve", QueryAndPreserve},
        {"QueryAndPay", QueryAndPay},
        {"QueryAndCancel", QueryAndCancel},
        {"QueryAndCollect", QueryAndCollect},
        {"QueryAndExecute", QueryAndExecute},
        {"QueryAndConsign", QueryAndConsign},
        {"QueryAndRebook", QueryAndRebook},
    }

    var wg sync.WaitGroup

    for i := 0; i < ThreadCount; i ++ {
        wg.Add(1)
        go worker(i, url, scenarios, &wg)
    }

    for _, scenario := range scenarios {
        UpdateBaseDate() // Update BaseDate to a new random date before each scenario
        log.Printf("Using BaseDate %s for scenario: %s", BaseDate.Format("2006-01-02"), scenario.name)

        q := NewQuery(url)
        log.Printf("Attempting to login for scenario: %s", scenario.name)
        err = q.Login("fdse_microservice", "111111")
        if err != nil {
            log.Printf("Login failed for scenario %s: %v", scenario.name, err)
            continue
        }
        log.Printf("Login successful for scenario: %s", scenario.name)

        log.Printf("Starting scenario: %s", scenario.name)
        scenario.function(q)
        log.Printf("Completed scenario: %s", scenario.name)

        time.Sleep(2 * time.Second) // Add a small delay between scenarios
    }



    if err != nil {
        log.Printf("An error occurred: %v", err)
    }
}

func worker(id int, url string, scenarios []struct{
    name     string
    function func(*Query)
}, wg *sync.WaitGroup) {
    defer wg.Done()
    
    q := NewQuery(url)
    log.Printf("Worker %d: Attempting to login", id)
    err := q.Login("fdse_microservice", "111111")
    if err != nil {
        log.Printf("Worker %d: Login failed: %v", id, err)
        return
    }
    log.Printf("Worker %d: Login successful", id)
    
    for i := 0; i < ScenariosPerThread; i++ {
        UpdateBaseDate() // Update BaseDate to a new random date before each scenario
        
        if len(scenarios) == 0 {
            log.Printf("Worker %d: No scenarios available", id)
            return
        }
        
        randomIndex := rand.Intn(len(scenarios))
        scenario := scenarios[randomIndex]
        
        if scenario.name == "" || scenario.function == nil {
            log.Printf("Worker %d: Invalid scenario at index %d", id, randomIndex)
            continue
        }
        
        log.Printf("Worker %d: Starting scenario %d/%d: %s", id, i+1, ScenariosPerThread, scenario.name)
        scenario.function(q)
        log.Printf("Worker %d: Completed scenario %d/%d: %s", id, i+1, ScenariosPerThread, scenario.name)
    }
    log.Printf("Worker %d: Completed all %d scenarios", id, ScenariosPerThread)
}

    // Direct query executions
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