package main

import (
    "fmt"
    "sync"
    "time"
)

// ScenarioStats tracks execution statistics for scenarios
type ScenarioStats struct {
    mu sync.Mutex
    // Map of worker ID -> scenario name -> count
    stats map[int]map[string]int
    startTime time.Time
}

// NewScenarioStats creates a new ScenarioStats instance
func NewScenarioStats() *ScenarioStats {
    return &ScenarioStats{
        stats: make(map[int]map[string]int),
        startTime: time.Now(),
    }
}

// IncrementScenario safely increments the count for a specific scenario in a worker
func (s *ScenarioStats) IncrementScenario(workerID int, scenarioName string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.stats[workerID]; !exists {
        s.stats[workerID] = make(map[string]int)
    }
    s.stats[workerID][scenarioName]++
}

// GetStats returns formatted statistics for all workers and scenarios
func (s *ScenarioStats) GetStats() string {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    duration := time.Since(s.startTime).Seconds()
    
    // First, get all unique scenario names
    scenarioNames := make(map[string]bool)
    for _, workerStats := range s.stats {
        for name := range workerStats {
            scenarioNames[name] = true
        }
    }
    
    // Build the statistics report
    var result string
    result = "\nLoad Test Statistics:\n"
    result += fmt.Sprintf("Total Duration: %.2f seconds\n\n", duration)
    
    // Per-worker statistics
    for workerID := range s.stats {
        result += fmt.Sprintf("Worker %d Statistics:\n", workerID)
        totalScenarios := 0
        
        for scenarioName := range scenarioNames {
            count := s.stats[workerID][scenarioName]
            rate := float64(count) / duration
            result += fmt.Sprintf("  %-20s: %5d total, %8.2f/sec\n", 
                scenarioName, count, rate)
            totalScenarios += count
        }
        
        totalRate := float64(totalScenarios) / duration
        result += fmt.Sprintf("  %-20s: %5d total, %8.2f/sec\n\n", 
            "Total", totalScenarios, totalRate)
    }
    
    // Calculate and add global statistics
    result += "Global Statistics:\n"
    globalStats := make(map[string]int)
    totalGlobal := 0
    
    for _, workerStats := range s.stats {
        for name, count := range workerStats {
            globalStats[name] += count
            totalGlobal += count
        }
    }
    
    for scenarioName := range scenarioNames {
        count := globalStats[scenarioName]
        rate := float64(count) / duration
        result += fmt.Sprintf("  %-20s: %5d total, %8.2f/sec\n",
            scenarioName, count, rate)
    }
    
    globalRate := float64(totalGlobal) / duration
    result += fmt.Sprintf("  %-20s: %5d total, %8.2f/sec\n",
        "Total", totalGlobal, globalRate)
    
    return result
}