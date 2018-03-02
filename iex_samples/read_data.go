package main

import (
	"encoding/json"
	"fmt"
	"time"

	realtime "github.com/135yshr/intrinio-realtime-go-sdk"
)

const (
	cUserName = "YOUR_INTRINIO_API_USERNAME"
	cPassword = "YOUR_INTRINIO_API_PASSWORD"
)

var (
	quoteCount int
	tradeCount int
)

func main() {
	fmt.Println("start!")
	c := realtime.New(cUserName, cPassword, realtime.IEX)
	c.OnQuote(quoteHandler)
	c.OnError(errorHandler)

	if err := c.Connect(); err != nil {
		fmt.Println(err)
		return
	}
	defer c.Disconnect()
	fmt.Println("connected!")

	c.Join("AAPL")
	fmt.Println("please wait for 1min!")
	time.Sleep(time.Minute)
	c.LeaveAll()

	fmt.Printf("quote = %d, trade = %d\n", quoteCount, tradeCount)
}

func quoteHandler(d map[string]interface{}) {
	j, err := json.Marshal(d)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(j))
}

func errorHandler(err error) {
	fmt.Printf("It is serious! An error has occurred!! error = %v\n", err)
}
