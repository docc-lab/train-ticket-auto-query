package main

import (
    "math/rand"
    "time"
)

var BaseDate time.Time

func init() {
    rand.Seed(time.Now().UnixNano())
}

// UpdateBaseDate sets BaseDate to a random date within 30 days of the original BaseDate
func UpdateBaseDate() {
    daysToAdd := rand.Intn(30) // Random number of days to add (0 to 29)
    BaseDate = BaseDate.AddDate(0, 0, daysToAdd)
}