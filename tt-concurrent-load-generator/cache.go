package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/semaphore"
	"io"
	"log"
	"net/http"
	"time"
)

type OrderCacheManager struct {
	OrdersCache      []map[string]interface{}
	OCTime           time.Time
	OrdersCacheOther []map[string]interface{}
	OCTimeOther      time.Time
	QuerySem         *semaphore.Weighted
	other            bool
}

var OCManager OrderCacheManager

func InitOCM() {
	OCManager.QuerySem = semaphore.NewWeighted(int64(ThreadCount))
}

func UpdateOrderCache(q *Query) {
	var url string
	if OCManager.other {
		url = fmt.Sprintf("%s/api/v1/orderOtherService/orderOther/refresh", q.Address)
	} else {
		url = fmt.Sprintf("%s/api/v1/orderservice/order/refresh", q.Address)
	}

	payload := map[string]string{
		"loginId": q.UID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
	}

	log.Printf("Querying orders with payload: %s", string(jsonPayload))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.Token)

	resp, err := q.Client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
	}

	// log.Printf("Order query response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		log.Printf("Unexpected response structure: %+v", result)
	}

	if len(data) > 0 {
		if OCManager.other {
			OCManager.OrdersCacheOther = make([]map[string]interface{}, len(data))
			OCManager.OCTimeOther = time.Now()

			for i, order := range data {
				orderMap, ok := order.(map[string]interface{})
				if !ok {
					log.Printf("Unexpected order structure: %+v", order)
					continue
				}

				orderInfo := map[string]interface{}{
					"accountId":   orderMap["accountId"],
					"targetDate":  time.Now().Format("2006-01-02 15:04:05"),
					"orderId":     orderMap["id"],
					"from":        orderMap["from"],
					"to":          orderMap["to"],
					"trainNumber": orderMap["trainNumber"],
					"status":      orderMap["status"],
				}

				OCManager.OrdersCacheOther[i] = orderInfo
			}
		} else {
			OCManager.OrdersCache = make([]map[string]interface{}, len(data))
			OCManager.OCTime = time.Now()

			for i, order := range data {
				orderMap, ok := order.(map[string]interface{})
				if !ok {
					log.Printf("Unexpected order structure: %+v", order)
					continue
				}

				orderInfo := map[string]interface{}{
					"accountId":   orderMap["accountId"],
					"targetDate":  time.Now().Format("2006-01-02 15:04:05"),
					"orderId":     orderMap["id"],
					"from":        orderMap["from"],
					"to":          orderMap["to"],
					"trainNumber": orderMap["trainNumber"],
					"status":      orderMap["status"],
				}

				OCManager.OrdersCache[i] = orderInfo
			}
		}
	}

	OCManager.other = !OCManager.other
}
