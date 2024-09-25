package operations

import (
    "log"

    ".."
)

func QueryFood(q *main.Query) {
    placePair := [2]string{"Shang Hai", "Su Zhou"}
    trainNumber := "D1345"

    foods, err := q.QueryFood(placePair, trainNumber)
    if err != nil {
        log.Printf("Error querying food: %v", err)
        return
    }

    if len(foods) == 0 {
        log.Println("No food options available")
        return
    }

    log.Printf("Found %d food options", len(foods))
    for _, food := range foods {
        log.Printf("Food: %s, Price: %.2f, Type: %d, Station: %s, Store: %s",
            food["foodName"], food["foodPrice"], food["foodType"],
            food["stationName"], food["storeName"])
    }
}