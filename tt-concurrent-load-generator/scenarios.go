package main

import (
	"log"
	"strings"
	"time"
)

var highspeedWeights = map[bool]int{true: 60, false: 40}

func QueryAndCancel(q *Query) {
	log.Println("Starting QueryAndCancel operation")
	pairs := make([][2]string, 0)
	var err error

	rfw := RandomFromWeighted(highspeedWeights)

	//if len(pairs) == 0 {
	//    log.Println("No orders found for cancellation")
	//    return
	//}

	for len(pairs) == 0 {
		if rfw {
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

func QueryAndPreserve(q *Query) {
	log.Println("Starting QueryAndPreserve operation")
	start := ""
	end := ""
	var tripIDs []string
	var tripDate string
	var err error

	highSpeed := RandomFromWeighted(highspeedWeights)
	if highSpeed {
		start = "Shang Hai"
		end = "Su Zhou"
		log.Printf("Querying high-speed ticket from %s to %s for date %s", start, end, BaseDate.Format("2006-01-02"))
		tripIDs, tripDate, err = q.QueryHighSpeedTicket([2]string{start, end}, BaseDate)
	} else {
		start = "Shang Hai"
		end = "Nan Jing"
		log.Printf("Querying normal ticket from %s to %s for date %s", start, end, BaseDate.Format("2006-01-02"))
		tripIDs, tripDate, err = q.QueryNormalTicket([2]string{start, end}, BaseDate)
	}

	if err != nil {
		log.Printf("Error querying tickets: %v", err)
		return
	}

	log.Printf("Found %d trips. Trip date: %s", len(tripIDs), tripDate)

	if len(tripIDs) == 0 {
		log.Printf("No trips available from %s to %s on %s", start, end, BaseDate.Format("2006-01-02"))
		return
	}

	log.Println("Attempting to preserve ticket")
	err = q.Preserve(start, end, tripIDs, highSpeed, tripDate)
	if err != nil {
		log.Printf("Error preserving ticket: %v", err)
		return
	}

	log.Printf("Ticket preserved successfully from %s to %s for %s", start, end, tripDate)
}

func QueryAndConsign(q *Query) {
	log.Println("Starting QueryAndConsign operation")
	var list []map[string]interface{}
	var err error

	if RandomFromWeighted(highspeedWeights) {
		log.Println("Querying high-speed orders for consignment")
		list, err = q.QueryOrdersAllInfo(false)
	} else {
		log.Println("Querying normal orders for consignment")
		list, err = q.QueryOrdersAllInfo(true)
	}

	if err != nil {
		log.Printf("Error querying orders for consignment: %v", err)
		return
	}

	if len(list) == 0 {
		log.Println("No orders found for consignment")
		return
	}

	// Try consigning orders until one succeeds or we run out of orders
	for _, selectedOrder := range list {
		orderID, ok := selectedOrder["orderId"].(string)
		if !ok {
			log.Println("Error: orderId not found or not a string in selected order")
			continue
		}

		log.Printf("Attempting to consign order %s", orderID)

		err = q.PutConsign(selectedOrder)
		if err != nil {
			log.Printf("Error putting consign for order %s: %v", orderID, err)
			if strings.Contains(err.Error(), "403") {
				log.Printf("403 Forbidden error for order %s. This order may not be in a state that allows consignment. Trying next order.", orderID)
				continue
			}
			// For other errors, we'll stop trying
			return
		}

		log.Printf("Order %s successfully queried and consigned", orderID)
		return
	}

	log.Println("Unable to consign any of the queried orders")
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
	log.Println("Starting QueryAndRebook operation")
	var pairs [][2]string
	var err error

	if RandomFromWeighted(highspeedWeights) {
		log.Println("Querying high-speed orders for rebooking")
		pairs, err = q.QueryOrders([]int{0, 1}, false)
	} else {
		log.Println("Querying normal orders for rebooking")
		pairs, err = q.QueryOrders([]int{0, 1}, true)
	}

	if err != nil {
		log.Printf("Error querying orders for rebooking: %v", err)
		return
	}

	if len(pairs) == 0 {
		log.Println("No orders found for rebooking")
		return
	}

	pair := RandomFromList(pairs).([2]string)
	orderID, tripID := pair[0], pair[1]

	log.Printf("Selected order %s (Trip: %s) for rebooking", orderID, tripID)

	// First, cancel the order
	err = q.CancelOrder(orderID, q.UID)
	if err != nil {
		log.Printf("Error cancelling order %s: %v", orderID, err)
		return
	}
	log.Printf("Order %s successfully canceled", orderID)

	// Now attempt to rebook
	newTripID := tripID // For simplicity, we're rebooking to the same trip
	newDate := time.Now().Format("2006-01-02")
	newSeatType := RandomFromList([]string{"2", "3"}).(string)

	err = q.RebookTicket(orderID, tripID, newTripID, newDate, newSeatType)
	if err != nil {
		log.Printf("Error rebooking ticket for order %s: %v", orderID, err)
		return
	}

	log.Printf("Order %s successfully rebooked to trip %s on %s with seat type %s", orderID, newTripID, newDate, newSeatType)
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
