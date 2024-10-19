package main

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "sync"
    "time"
    "math/rand"
)

var (
    ThreadCount int
    DurationSeconds int
)

func main() {
    // Set up logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    // Parse command-line arguments
    if len(os.Args) != 5 {
        fmt.Println("Usage: ./tt-concurrent-load-generator <TRAIN_TICKET_UI_IPADDR> <BASE_DATE> <NUM_THREADS> <DURATION_SECONDS>")
        os.Exit(1)
    }

    ipAddr := os.Args[1]
    baseDate := os.Args[2]
    var err error
    ThreadCount, err = strconv.Atoi(os.Args[3])
    if err != nil {
        log.Fatalf("Invalid thread count: %v", err)
    }
    DurationSeconds, err = strconv.Atoi(os.Args[4])
    if err != nil {
        log.Fatalf("Invalid duration: %v", err)
    }

    url := fmt.Sprintf("http://%s:8080", ipAddr)
    log.Printf("Connecting to: %s", url)

    BaseDate, err = time.Parse("2006-01-02", baseDate)
    if err != nil {
        log.Fatalf("Invalid date format: %v", err)
    }

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
    stopChan := make(chan struct{})

    for i := 0; i < ThreadCount; i++ {
        wg.Add(1)
        go worker(i, url, scenarios, &wg, stopChan)
    }

    // Run for the specified duration
    time.Sleep(time.Duration(DurationSeconds) * time.Second)
    close(stopChan)

    wg.Wait()

    log.Println("Load test completed")
}

func worker(id int, url string, scenarios []struct{
    name     string
    function func(*Query)
}, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()
    
    q := NewQuery(url)
    log.Printf("Worker %d: Attempting to login", id)
    err := q.Login("fdse_microservice", "111111")
    if err != nil {
        log.Printf("Worker %d: Login failed: %v", id, err)
        return
    }
    log.Printf("Worker %d: Login successful", id)
    
    scenarioCount := 0
    for {
        select {
        case <-stopChan:
            log.Printf("Worker %d: Stopping after executing %d scenarios", id, scenarioCount)
            return
        default:
            UpdateBaseDate() // Update BaseDate to a new random date before each scenario
            
            randomIndex := rand.Intn(len(scenarios))
            scenario := scenarios[randomIndex]
            
            log.Printf("Worker %d: Starting scenario %d: %s", id, scenarioCount+1, scenario.name)
            scenario.function(q)
            log.Printf("Worker %d: Completed scenario %d: %s", id, scenarioCount+1, scenario.name)
            
            scenarioCount++
        }
    }
}