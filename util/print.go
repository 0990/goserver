package util

import (
	"fmt"
	"time"
)

func PrintCurrNano(title string) {
	fmt.Println(title, ":", time.Now().UnixNano()/1000, time.Now().Unix())
}

func PrintGoroutineID(title string) {
	fmt.Println(title, "goroutineID:", GetGoroutineID())
}
