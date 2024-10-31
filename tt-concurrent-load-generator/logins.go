package main

type LoginManager struct {
	Usernames []string
	Passwords []string
	Usages    map[int]bool
}

var Manager LoginManager

func Init() {

}
