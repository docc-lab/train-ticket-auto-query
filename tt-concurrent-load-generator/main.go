package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ThreadCount     int
	DurationSeconds int
	stats           *ScenarioStats
	activeScenarios []byte
	sem             *semaphore.Weighted
)

func main() {
	// Add warm-up flag
	isWarmup := flag.Bool("warmup", false, "Run in warm-up mode")
	flag.Parse()

	args := flag.Args()

	// Check arguments based on mode
	if *isWarmup {
		if len(args) != 3 {
			fmt.Println("Warm-up mode usage: ./tt-concurrent-load-generator -warmup <TRAIN_TICKET_UI_IPADDR> <BASE_DATE> <NUM_THREADS>")
			os.Exit(1)
		}
	} else {
		if len(args) < 4 || len(args) > 5 {
			fmt.Println("Load test mode usage: ./tt-concurrent-load-generator <TRAIN_TICKET_UI_IPADDR> <BASE_DATE> <NUM_THREADS> <DURATION_SECONDS> [<SCENARIO_FLAGS>]")
			os.Exit(1)
		}
	}

	ipAddr := args[0]
	baseDate := args[1]
	var err error
	ThreadCount, err = strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("Invalid thread count: %v", err)
	}

	if len(args) == 5 {
		if len(args[4]) != 8 {
			log.Fatalf("Invalid bitmap length for scenarios!")
		}

		activeScenarios = make([]byte, 8)
		for i, e := range strings.Split(args[4], "") {
			if e == "0" {
				activeScenarios[i] = 0
			} else if e == "1" {
				activeScenarios[i] = 1
			} else {
				log.Fatalf("Invalid bitmap value for scenarios!")
			}
		}
	} else {
		activeScenarios = []byte{1, 1, 1, 1, 1, 1, 1, 1}
	}

	if !*isWarmup {
		DurationSeconds, err = strconv.Atoi(args[3])
		if err != nil {
			log.Fatalf("Invalid duration: %v", err)
		}
	}

	url := fmt.Sprintf("http://%s:8080", ipAddr)
	log.Printf("Connecting to: %s", url)

	BaseDate, err = time.Parse("2006-01-02", baseDate)
	if err != nil {
		log.Fatalf("Invalid date format: %v", err)
	}

	if *isWarmup {
		runWarmup(url)
	} else {
		runLoadTest(url)
	}
}

func runWarmup(url string) {
	log.Println("Starting warm-up session...")

	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	counter := NewWarmupCounter()
	startTime := time.Now()

	// Initialize order cache manager
	InitOCM()

	go dataFetchWorker(url, &wg, stopChan)

	time.Sleep(time.Second)

	for i := 0; i < ThreadCount; i++ {
		wg.Add(1)
		go WarmupWorker(i, url, &wg, counter)
	}

	wg.Wait()
	duration := time.Since(startTime)

	close(stopChan)

	log.Printf("Warm-up completed in %v. Created orders:", duration)
	log.Printf("- Unpaid orders: %d (target: 2000)", counter.unpaidCount)
	log.Printf("- Paid orders: %d (target: 1000)", counter.paidCount)
	log.Printf("- Collected orders: %d (target: 1000)", counter.collectedCount)
	log.Printf("- Consigned orders: %d (target: 1000)", counter.consignedCount)
	log.Printf("Total orders created: %d", counter.getTotalCount())
}

func runLoadTest(url string) {
	// Original load test logic
	allScenarios := []struct {
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
		{"QueryOnlyHighSpeed", QueryOnlyHighSpeed},
	}

	scenarios := make([]struct {
		name     string
		function func(*Query)
	}, 0)

	for i := 0; i < 8; i += 1 {
		if activeScenarios[i] == 1 {
			scenarios = append(scenarios, allScenarios[i])
		}
	}

	// Initialize statistics tracking
	stats = NewScenarioStats()

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	sem = semaphore.NewWeighted(int64(ThreadCount))

	// Initialize order cache manager
	InitOCM()

	go dataFetchWorker(url, &wg, stopChan)

	time.Sleep(time.Second)

	for i := 0; i < ThreadCount; i++ {
		wg.Add(1)
		go worker(i, url, scenarios, &wg, stopChan)
	}

	// Run for the specified duration
	time.Sleep(time.Duration(DurationSeconds) * time.Second)
	close(stopChan)

	wg.Wait()

	// Print statistics
	log.Println(stats.GetStats())
	log.Println("Load test completed")
}

func worker(id int, url string, scenarios []struct {
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
		//_ = sem.Acquire(context.Background(), 1)

		select {
		case <-stopChan:
			log.Printf("Worker %d: Stopping after executing %d scenarios", id, scenarioCount)

			//sem.Release(1)

			return
		default:
			UpdateBaseDate() // Update BaseDate to a new random date before each scenario

			randomIndex := rand.Intn(len(scenarios))
			scenario := scenarios[randomIndex]

			log.Printf("Worker %d: Starting scenario %d: %s", id, scenarioCount+1, scenario.name)
			scenario.function(q)
			stats.IncrementScenario(id, scenario.name)
			log.Printf("Worker %d: Completed scenario %d: %s", id, scenarioCount+1, scenario.name)

			scenarioCount++
		}

		//sem.Release(1)
	}
}

func dataFetchWorker(url string, wg *sync.WaitGroup, stopChan <-chan struct{}) {
	defer wg.Done()

	acquired := 0

	q := NewQuery(url)
	log.Printf("Order query worker: Attempting to login")
	err := q.Login("fdse_microservice", "111111")
	if err != nil {
		log.Printf("Order query worker: Login failed: %v", err)
		return
	}
	log.Printf("Order query worker: Login successful")

	for {
		select {
		case <-stopChan:
			log.Printf("Order query worker stopping!")
			return
		default:
			if acquired == ThreadCount {
				log.Printf("ACQUIRED ALL RESOURCES FOR CACHE UPDATE")
				UpdateOrderCache(q)
				OCManager.QuerySem.Release(int64(ThreadCount))
				acquired = 0
				log.Printf("Order query worker: Attempting to login")
				err := q.Login("fdse_microservice", "111111")
				if err != nil {
					log.Printf("Order query worker: Login failed: %v", err)
					return
				}
				log.Printf("Order query worker: Login successful")
				time.Sleep(time.Second * time.Duration(rand.Intn(10)+20))
			} else {
				_ = OCManager.QuerySem.Acquire(context.Background(), 1)
				acquired += 1
				log.Printf("Total cache update resources acquired: [ %v ]", acquired)
			}
		}
	}
}
