package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "log"
)

type Query struct {
    Address string
    UID     string
    Token   string
    Client  *http.Client
}

func NewQuery(address string) *Query {
    return &Query{
        Address: address,
        Client:  &http.Client{},
    }
}

type tokenTransport struct {
    Token string
    Base  http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req.Header.Set("Authorization", "Bearer "+t.Token)
    return t.Base.RoundTrip(req)
}

func (q *Query) Login(username, password string) error {
    url := fmt.Sprintf("%s/api/v1/users/login", q.Address)
    payload := map[string]string{
        "username": username,
        "password": password,
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].(map[string]interface{})
    q.UID = data["userId"].(string)
    q.Token = data["token"].(string)
    q.Client.Transport = &tokenTransport{Token: q.Token, Base: http.DefaultTransport}

    return nil
}

func (q *Query) QueryHighSpeedTicket(placePair [2]string, date time.Time) ([]string, error) {
    return q.queryTicket(placePair, date, true)
}

func (q *Query) QueryNormalTicket(placePair [2]string, date time.Time) ([]string, error) {
    return q.queryTicket(placePair, date, false)
}

func (q *Query) queryTicket(placePair [2]string, date time.Time, isHighSpeed bool) ([]string, error) {
    var url string
    if isHighSpeed {
        url = fmt.Sprintf("%s/api/v1/travelservice/trips/left", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/travel2service/trips/left", q.Address)
    }

    payload := map[string]string{
        "departureTime": date.Format("2006-01-02"),
        "startPlace":    placePair[0],
        "endPlace":      placePair[1],
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query ticket failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].([]interface{})
    tripIDs := make([]string, len(data))
    for i, trip := range data {
        tripMap := trip.(map[string]interface{})
        tripID := tripMap["tripId"].(map[string]interface{})
        tripIDs[i] = tripID["type"].(string) + tripID["number"].(string)
    }

    return tripIDs, nil
}

func (q *Query) QueryHighSpeedTicketParallel(placePair [2]string, date time.Time) ([]string, error) {
    url := fmt.Sprintf("%s/api/v1/travelservice/trips/left_parallel", q.Address)
    payload := map[string]string{
        "departureTime": date.Format("2006-01-02"),
        "startPlace":    placePair[0],
        "endPlace":      placePair[1],
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query high speed ticket parallel failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].([]interface{})
    tripIDs := make([]string, len(data))
    for i, trip := range data {
        tripMap := trip.(map[string]interface{})
        tripID := tripMap["tripId"].(map[string]interface{})
        tripIDs[i] = tripID["type"].(string) + tripID["number"].(string)
    }

    return tripIDs, nil
}

func (q *Query) QueryOrders(orderTypes []int, queryOther bool) ([][2]string, error) {
    var url string
    if queryOther {
        url = fmt.Sprintf("%s/api/v1/orderOtherService/orderOther/refresh", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/orderservice/order/refresh", q.Address)
    }

    payload := map[string]string{
        "loginId": q.UID,
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query orders failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].([]interface{})
    var pairs [][2]string

    for _, order := range data {
        orderMap := order.(map[string]interface{})
        status := int(orderMap["status"].(float64))
        for _, t := range orderTypes {
            if status == t {
                pair := [2]string{orderMap["id"].(string), orderMap["trainNumber"].(string)}
                pairs = append(pairs, pair)
                break
            }
        }
    }

    return pairs, nil
}

func (q *Query) CancelOrder(orderID, uuid string) error {
    url := fmt.Sprintf("%s/api/v1/cancelservice/cancel/%s/%s", q.Address, orderID, uuid)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("cancel order failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) CollectTicket(orderID string) error {
    url := fmt.Sprintf("%s/api/v1/executeservice/execute/collected/%s", q.Address, orderID)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("collect ticket failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) EnterStation(orderID string) error {
    url := fmt.Sprintf("%s/api/v1/executeservice/execute/execute/%s", q.Address, orderID)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("enter station failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) QueryAssurances() ([]map[string]string, error) {
    url := fmt.Sprintf("%s/api/v1/assuranceservice/assurances/types", q.Address)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query assurances failed with status code: %d", resp.StatusCode)
    }

    // As per the Python implementation, we're returning a fixed value
    return []map[string]string{{"assurance": "1"}}, nil
}

func (q *Query) QueryFood(placePair [2]string, trainNum string) ([]map[string]interface{}, error) {
    url := fmt.Sprintf("%s/api/v1/foodservice/foods/%s/%s/%s/%s", q.Address, time.Now().Format("2006-01-02"), placePair[0], placePair[1], trainNum)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query food failed with status code: %d", resp.StatusCode)
    }

    // As per the Python implementation, we're returning a fixed value
    return []map[string]interface{}{
        {
            "foodName":   "Soup",
            "foodPrice":  3.7,
            "foodType":   2,
            "stationName": "Su Zhou",
            "storeName":  "Roman Holiday",
        },
    }, nil
}

func (q *Query) QueryContacts() ([]string, error) {
    url := fmt.Sprintf("%s/api/v1/contactservice/contacts/account/%s", q.Address, q.UID)

    req, _ := http.NewRequest("GET", url, nil)
    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query contacts failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].([]interface{})
    contactIDs := make([]string, len(data))
    for i, contact := range data {
        contactMap := contact.(map[string]interface{})
        contactIDs[i] = contactMap["id"].(string)
    }

    return contactIDs, nil
}

func (q *Query) Preserve(start, end string, tripIDs []string, isHighSpeed bool) error {
    var url string
    if isHighSpeed {
        url = fmt.Sprintf("%s/api/v1/preserveservice/preserve", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/preserveotherservice/preserveOther", q.Address)
    }

    payload := map[string]interface{}{
        "accountId":  q.UID,
        "contactsId": "",
        "tripId":     RandomFromList(tripIDs),
        "seatType":   RandomFromList([]string{"2", "3"}),
        "date":       time.Now().Format("2006-01-02"),
        "from":       start,
        "to":         end,
        "assurance":  "0",
        "foodType":   "0",
    }

    if RandomBoolean() {
        payload["assurance"] = "1"
    }

    contacts, err := q.QueryContacts()
    if err == nil && len(contacts) > 0 {
        payload["contactsId"] = RandomFromList(contacts)
    }

    if RandomBoolean() {
        food, err := q.QueryFood([2]string{start, end}, payload["tripId"].(string))
        if err == nil && len(food) > 0 {
            selectedFood := food[0]
            for k, v := range selectedFood {
                payload[k] = v
            }
        }
    }

    if RandomBoolean() {
        payload["consigneeName"] = RandomString(10)
        payload["consigneePhone"] = RandomPhone()
        payload["consigneeWeight"] = RandomFromList([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
        payload["handleDate"] = time.Now().Format("2006-01-02")
    }

    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("preserve failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    if result["data"] != "Success" {
        return fmt.Errorf("preserve failed: %v", result["data"])
    }

    return nil
}

func (q *Query) PutConsign(consignInfo map[string]interface{}) error {
    url := fmt.Sprintf("%s/api/v1/consignservice/consigns", q.Address)

    consignInfo["handleDate"] = time.Now().Format("2006-01-02")
    consignInfo["targetDate"] = time.Now().Format("2006-01-02 15:04:05")
    consignInfo["consignee"] = "32"
    consignInfo["phone"] = "12345677654"
    consignInfo["weight"] = "32"
    consignInfo["id"] = ""
    consignInfo["isWithin"] = false

    jsonPayload, _ := json.Marshal(consignInfo)

    req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("put consign failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) PayOrder(orderID, tripID string) error {
    url := fmt.Sprintf("%s/api/v1/inside_pay_service/inside_payment", q.Address)

    payload := map[string]string{
        "orderId": orderID,
        "tripId":  tripID,
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("pay order failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) RebookTicket(oldOrderID, oldTripID, newTripID string) error {
    url := fmt.Sprintf("%s/api/v1/rebookservice/rebook", q.Address)

    payload := map[string]string{
        "oldTripId": oldTripID,
        "orderId":   oldOrderID,
        "tripId":    newTripID,
        "date":      time.Now().Format("2006-01-02"),
        "seatType":  RandomFromList([]string{"2", "3"}).(string),
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("rebook ticket failed with status code: %d", resp.StatusCode)
    }

    return nil
}

func (q *Query) QueryOrdersAllInfo(queryOther bool) ([]map[string]interface{}, error) {
    var url string
    if queryOther {
        url = fmt.Sprintf("%s/api/v1/orderOtherService/orderOther/refresh", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/orderservice/order/refresh", q.Address)
    }

    payload := map[string]string{
        "loginId": q.UID,
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query orders all info failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    data := result["data"].([]interface{})
    orderInfoList := make([]map[string]interface{}, len(data))

    for i, order := range data {
        orderMap := order.(map[string]interface{})
        orderInfo := map[string]interface{}{
            "accountId":  orderMap["accountId"],
            "targetDate": time.Now().Format("2006-01-02 15:04:05"),
            "orderId":    orderMap["id"],
            "from":       orderMap["from"],
            "to":         orderMap["to"],
        }
        orderInfoList[i] = orderInfo
    }

    return orderInfoList, nil
}

func (q *Query) QueryAdminBasicPrice() (*http.Response, error) {
    url := fmt.Sprintf("%s/api/v1/adminbasicservice/adminbasic/prices", q.Address)

    resp, err := q.Client.Get(url)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode == http.StatusOK {
        log.Println("Query price success")
        return resp, nil
    } else {
        log.Printf("Query price failed with status code: %d", resp.StatusCode)
        return nil, fmt.Errorf("query price failed")
    }
}

func (q *Query) QueryAdminBasicConfig() (*http.Response, error) {
    url := fmt.Sprintf("%s/api/v1/adminbasicservice/adminbasic/configs", q.Address)

    resp, err := q.Client.Get(url)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode == http.StatusOK {
        log.Println("Config success")
        return resp, nil
    } else {
        log.Printf("Config failed with status code: %d", resp.StatusCode)
        return nil, fmt.Errorf("config failed")
    }
}

func (q *Query) QueryRoute(routeId string) error {
    var url string
    if routeId == "" {
        url = fmt.Sprintf("%s/api/v1/routeservice/routes", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/routeservice/routes/%s", q.Address, routeId)
    }

    resp, err := q.Client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        log.Printf("Query routeId success")
    } else {
        log.Printf("Query routeId: %s fail, code: %d, text: %s", routeId, resp.StatusCode, resp.Status)
        return fmt.Errorf("query route failed")
    }

    return nil
}

func (q *Query) QueryAdminTravel() error {
    url := fmt.Sprintf("%s/api/v1/admintravelservice/admintravel", q.Address)

    resp, err := q.Client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)
        if result["status"] == float64(1) {
            log.Println("Success to query admin travel")
            return nil
        }
    }

    log.Printf("Failed to query admin travel with status code: %d", resp.StatusCode)
    return fmt.Errorf("query admin travel failed")
}