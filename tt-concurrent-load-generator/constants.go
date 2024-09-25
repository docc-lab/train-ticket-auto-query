package main

import "time"

var BaseDate = time.Date(2024, 9, 29, 0, 0, 0, 0, time.UTC)

func init() {
    var err error
    BaseDate, err = time.Parse("2006-01-02", "2024-09-29") // Default date
    if err != nil {
        panic("Failed to parse base date: " + err.Error())
    }
}