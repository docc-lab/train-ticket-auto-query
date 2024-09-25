package main

import (
    "math/rand"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func RandomBoolean() bool {
    return rand.Intn(2) == 1
}


func RandomFromList(list interface{}) interface{} {
    switch v := list.(type) {
    case []string:
        if len(v) == 0 {
            return ""
        }
        return v[rand.Intn(len(v))]
    case [][2]string:
        if len(v) == 0 {
            return [2]string{"", ""}
        }
        return v[rand.Intn(len(v))]
    case []map[string]interface{}:
        if len(v) == 0 {
            return map[string]interface{}{}
        }
        return v[rand.Intn(len(v))]
    case []interface{}:
        if len(v) == 0 {
            return nil
        }
        return v[rand.Intn(len(v))]
    case map[string]interface{}:
        if len(v) == 0 {
            return nil
        }
        keys := make([]string, 0, len(v))
        for k := range v {
            keys = append(keys, k)
        }
        randomKey := keys[rand.Intn(len(keys))]
        return v[randomKey]
    default:
        return nil
    }
}

func RandomFromWeighted(weights map[bool]int) bool {
    total := 0
    for _, weight := range weights {
        total += weight
    }
    r := rand.Intn(total)
    for k, weight := range weights {
        r -= weight
        if r <= 0 {
            return k
        }
    }
    return false
}

func RandomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func RandomPhone() string {
    const digits = "0123456789"
    b := make([]byte, rand.Intn(8)+8)
    for i := range b {
        b[i] = digits[rand.Intn(len(digits))]
    }
    return string(b)
}