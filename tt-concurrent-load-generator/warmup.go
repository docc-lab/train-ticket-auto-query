package main

import (
    "log"
    "fmt"
    "time"
    "sync"
    "sync/atomic"
)

type WarmupCounter struct {
    unpaidCount    int32
    paidCount      int32
    collectedCount int32
    consignedCount int32
    mu            sync.Mutex
}

func NewWarmupCounter() *WarmupCounter {
    return &WarmupCounter{}
}

func (w *WarmupCounter) canCreateUnpaid() bool {
    return atomic.LoadInt32(&w.unpaidCount) < 1000
}

func (w *WarmupCounter) canCreatePaid() bool {
    return atomic.LoadInt32(&w.paidCount) < 500
}

func (w *WarmupCounter) canCreateCollected() bool {
    return atomic.LoadInt32(&w.collectedCount) < 500
}

func (w *WarmupCounter) canCreateConsigned() bool {
    return atomic.LoadInt32(&w.consignedCount) < 500
}

func (w *WarmupCounter) incrementUnpaid() {
    atomic.AddInt32(&w.unpaidCount, 1)
}

func (w *WarmupCounter) incrementPaid() {
    atomic.AddInt32(&w.paidCount, 1)
}

func (w *WarmupCounter) incrementCollected() {
    atomic.AddInt32(&w.collectedCount, 1)
}

func (w *WarmupCounter) incrementConsigned() {
    atomic.AddInt32(&w.consignedCount, 1)
}

func (w *WarmupCounter) getTotalCount() int32 {
    return atomic.LoadInt32(&w.unpaidCount) + 
           atomic.LoadInt32(&w.paidCount) + 
           atomic.LoadInt32(&w.collectedCount) + 
           atomic.LoadInt32(&w.consignedCount)
}

func WarmupWorker(id int, url string, wg *sync.WaitGroup, counter *WarmupCounter) {
    defer wg.Done()
    retryCount := 0
    maxRetries := 3
    
    q := NewQuery(url)
    for retryCount < maxRetries {
        // err := q.Login("test1", "111111")
        err := q.Login("fdse_microservice", "111111")
        if err != nil {
            log.Printf("Worker %d: Login attempt %d failed: %v", id, retryCount+1, err)
            retryCount++
            time.Sleep(time.Second * time.Duration(retryCount))
            continue
        }
        log.Printf("Worker %d: Login successful", id)
        break
    }
    
    if retryCount == maxRetries {
        log.Printf("Worker %d: Failed to login after %d attempts", id, maxRetries)
        return
    }

    for {
        total := counter.getTotalCount()
        if total >= 5000 {
            log.Printf("Worker %d: Target count reached, exiting", id)
            return
        }

        UpdateBaseDate() // Update BaseDate to a new random date

        var err error
        switch {
        case counter.canCreateUnpaid():
            err = createUnpaidOrder(q)
            if err == nil {
                counter.incrementUnpaid()
                log.Printf("Worker %d: Created unpaid order. Total unpaid: %d/1000", 
                    id, atomic.LoadInt32(&counter.unpaidCount))
            }

        case counter.canCreatePaid():
            err = createPaidOrder(q)
            if err == nil {
                counter.incrementPaid()
                log.Printf("Worker %d: Created paid order. Total paid: %d/500", 
                    id, atomic.LoadInt32(&counter.paidCount))
            }

        case counter.canCreateCollected():
            err = createCollectedOrder(q)
            if err == nil {
                counter.incrementCollected()
                log.Printf("Worker %d: Created collected order. Total collected: %d/500", 
                    id, atomic.LoadInt32(&counter.collectedCount))
            }

        case counter.canCreateConsigned():
            err = createConsignedOrder(q)
            if err == nil {
                counter.incrementConsigned()
                log.Printf("Worker %d: Created consigned order. Total consigned: %d/500", 
                    id, atomic.LoadInt32(&counter.consignedCount))
            }

        default:
            log.Printf("Worker %d: All quotas met", id)
            return
        }

        if err != nil {
            log.Printf("Worker %d: Error creating order: %v", id, err)
            // Add a small delay before retrying
            time.Sleep(time.Millisecond * 100)
        }
    }
}

func createUnpaidOrder(q *Query) error {
    start := "Shang Hai"
    end := "Su Zhou"
    tripIDs, tripDate, err := q.QueryHighSpeedTicket([2]string{start, end}, BaseDate)
    if err != nil {
        return fmt.Errorf("failed to query ticket: %v", err)
    }
    if len(tripIDs) == 0 {
        return fmt.Errorf("no trips available")
    }
    return q.Preserve(start, end, tripIDs, true, tripDate)
}

func createPaidOrder(q *Query) error {
    // First create unpaid order
    if err := createUnpaidOrder(q); err != nil {
        return fmt.Errorf("failed to create unpaid order: %v", err)
    }
    
    // Wait a short time for the order to be processed
    time.Sleep(time.Millisecond * 100)
    
    // Query the created order and pay for it
    pairs, err := q.QueryOrders([]int{0}, false)
    if err != nil {
        return fmt.Errorf("failed to query orders: %v", err)
    }
    if len(pairs) == 0 {
        return fmt.Errorf("no unpaid orders found")
    }
    
    orderID, tripID := pairs[0][0], pairs[0][1]
    return q.PayOrder(orderID, tripID)
}

func createCollectedOrder(q *Query) error {
    // First create paid order
    if err := createPaidOrder(q); err != nil {
        return fmt.Errorf("failed to create paid order: %v", err)
    }
    
    // Wait a short time for the payment to be processed
    time.Sleep(time.Millisecond * 200)
    
    // Query the paid order and collect it
    pairs, err := q.QueryOrders([]int{1}, false)
    if err != nil {
        return fmt.Errorf("failed to query orders: %v", err)
    }
    if len(pairs) == 0 {
        return fmt.Errorf("no paid orders found")
    }
    
    return q.CollectTicket(pairs[0][0])
}

func createConsignedOrder(q *Query) error {
    // First create paid order
    if err := createPaidOrder(q); err != nil {
        return fmt.Errorf("failed to create paid order: %v", err)
    }
    
    // Wait a short time for the payment to be processed
    time.Sleep(time.Millisecond * 200)
    
    // Query order info and consign it
    ordersList, err := q.QueryOrdersAllInfo(false)
    if err != nil {
        return fmt.Errorf("failed to query orders: %v", err)
    }
    if len(ordersList) == 0 {
        return fmt.Errorf("no orders found for consignment")
    }
    
    return q.PutConsign(ordersList[0])
}