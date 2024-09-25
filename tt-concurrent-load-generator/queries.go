package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/http/cookiejar"
    _url "net/url"
    "time"
)

type Query struct {
    Address string
    UID     string
    Token   string
    Client  *http.Client
    Cookies []*http.Cookie
}

func NewQuery(address string) *Query {
    jar, _ := cookiejar.New(nil)
    return &Query{
        Address: address,
        Client:  &http.Client{Jar: jar},
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
    
    // Create a cookie jar to handle cookies
    jar, _ := cookiejar.New(nil)
    q.Client.Jar = jar

    // Set initial cookies
    initialCookies := []*http.Cookie{
        {Name: "JSESSIONID", Value: "9ED5635A2A892A4BA31E7E98533A279D"},
        {Name: "YsbCaptcha", Value: "025080CF8BA94594B09E283F17815444"},
    }
    u, _ := _url.Parse(url)
    q.Client.Jar.SetCookies(u, initialCookies)

    payload := map[string]string{
        "username": username,
        "password": password,
    }
    jsonPayload, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Proxy-Connection", "keep-alive")
    req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
    req.Header.Set("X-Requested-With", "XMLHttpRequest")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
    req.Header.Set("Origin", url)
    req.Header.Set("Referer", fmt.Sprintf("%s/client_login.html", q.Address))
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

    resp, err := q.Client.Do(req)
    if err != nil {
        return fmt.Errorf("login request failed: %v", err)
    }
    defer resp.Body.Close()

    // Store cookies for future requests
    q.Cookies = q.Client.Jar.Cookies(u)

    // Read and print the raw response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response body: %v", err)
    }
    fmt.Printf("Raw response2: %s\n", string(body))

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
        return fmt.Errorf("failed to decode response: %v", err)
    }

    data, ok := result["data"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("unexpected response structure: %v", result)
    }

    userId, ok := data["userId"].(string)
    if !ok {
        return fmt.Errorf("userId not found in response")
    }
    q.UID = userId

    token, ok := data["token"].(string)
    if !ok {
        return fmt.Errorf("token not found in response")
    }
    q.Token = token

    return nil
}

func (q *Query) QueryHighSpeedTicket(placePair [2]string, date time.Time) ([]string, string, error) {
    return q.queryTicket(placePair, date, true)
}

func (q *Query) QueryNormalTicket(placePair [2]string, date time.Time) ([]string, string, error) {
    return q.queryTicket(placePair, date, false)
}

func (q *Query) queryTicket(placePair [2]string, date time.Time, isHighSpeed bool) ([]string, string, error) {
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
        return nil, "", fmt.Errorf("query ticket failed: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, "",fmt.Errorf("failed to read response body: %v", err)
    }
    fmt.Printf("Raw response1: %s\n", string(body))

    if resp.StatusCode != http.StatusOK {
        return nil, "",fmt.Errorf("query ticket failed with status code: %d", resp.StatusCode)
    }

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, "",fmt.Errorf("failed to decode response: %v", err)
    }

    data, ok := result["data"].([]interface{})
    if !ok {
        return []string{}, "", nil // Return empty slice if no data
    }

    tripIDs := make([]string, 0, len(data))
    var tripDate string
    for _, trip := range data {
        tripMap := trip.(map[string]interface{})
        tripID := tripMap["tripId"].(map[string]interface{})
        tripIDs = append(tripIDs, tripID["type"].(string)+tripID["number"].(string))
        
        // Extract the date from the startTime
        startTime := tripMap["startTime"].(string)
        tripDate = startTime[:10] // Assuming the format is "YYYY-MM-DD HH:MM:SS"
    }

    return tripIDs, tripDate, nil
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
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // Add authorization header
    req.Header.Set("Authorization", "Bearer "+q.Token)

    resp, err := q.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to query contacts: status %d, body: %s", resp.StatusCode, string(body))
    }

    resp, err = q.Client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    data, ok := result["data"].([]interface{})
    if !ok {
        return nil, fmt.Errorf("unexpected response structure")
    }

    contactIDs := make([]string, len(data))
    for i, contact := range data {
        contactMap, ok := contact.(map[string]interface{})
        if !ok {
            return nil, fmt.Errorf("unexpected contact structure")
        }
        contactIDs[i] = contactMap["id"].(string)
    }

    return contactIDs, nil
}

func (q *Query) Preserve(start, end string, tripIDs []string, isHighSpeed bool, date string) error {
    if len(tripIDs) == 0 {
        return fmt.Errorf("no trips available for preservation")
    }

    var url string
    if isHighSpeed {
        url = fmt.Sprintf("%s/api/v1/preserveservice/preserve", q.Address)
    } else {
        url = fmt.Sprintf("%s/api/v1/preserveotherservice/preserveOther", q.Address)
    }

    contacts_result, err := q.QueryContacts()
    if err != nil {
        return fmt.Errorf("failed to query contacts: %v", err)
    }

    if len(contacts_result) == 0 {
        return fmt.Errorf("no contacts found")
    }

    contactsId := RandomFromList(contacts_result).(string)

    payload := map[string]interface{}{
        "accountId":  q.UID,
        "contactsId": contactsId,
        "tripId":     RandomFromList(tripIDs).(string),
        "seatType":   RandomFromList([]string{"2", "3"}),
        "date":       date,
        "from":       start,
        "to":         end,
        "assurance":  "0",
        "foodType":   "0",
    }

    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+q.Token)

    log.Printf("Sending preserve request to %s with payload: %s", url, string(jsonPayload))

    resp, err := q.Client.Do(req)
    if err != nil {
        return fmt.Errorf("preserve request failed: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    log.Printf("Preserve response: %s", string(body))

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("preserve failed with status code: %d, body: %s", resp.StatusCode, string(body))
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

func (q *Query) RebookTicket(oldOrderID, oldTripID, newTripID string, date time.Time) error {
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