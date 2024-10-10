package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    "math/rand"
)

var (
    ThreadCount int
    ScenariosPerThread int
)

func main() {
    // Set up logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    // Parse command-line arguments
    if len(os.Args) != 5 {
        fmt.Println("Usage: ./tt-concurrent-load-generator <TRAIN_TICKET_UI_IPADDR> <BASE_DATE> <NUM_THREAD> <NUM_SCENARIOS_PER_THREAD>")
        os.Exit(1)
    }

    ipAddr := os.Args[1]
    baseDate := os.Args[2]
    ThreadCount, _ = strconv.Atoi(os.Args[3])
    ScenariosPerThread, _ = strconv.Atoi(os.Args[4])

    url := fmt.Sprintf("http://%s:8080", ipAddr)
    log.Printf("Connecting to: %s", url)

    var err error
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

    for i := 0; i < ThreadCount; i++ {
        wg.Add(1)
        go worker(i, url, scenarios, &wg)
    }

    wg.Wait()

    log.Println("All workers completed their scenarios")
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