package operations

import (
    "log"

    ".."
)

func QueryRoute(q *main.Query) {
    routeID := "92708982-77af-4318-be25-57ccb0ff69ad"

    err := q.QueryRoute(routeID)
    if err != nil {
        log.Printf("Error querying route: %v", err)
        return
    }

    log.Printf("Successfully queried route %s", routeID)
}