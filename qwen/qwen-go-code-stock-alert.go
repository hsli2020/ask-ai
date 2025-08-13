package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	UserID        string
	Username      string
	WatchedStocks map[string]*Stock
	Alarms        []*Alarm
}

type Stock struct {
	StockCode    string
	StockName    string
	CurrentPrice float64
}

type Alarm struct {
	Type      string
	Price     float64
	User      *User
	StockCode string
}

var stocks = make(map[string]*Stock)
var users = make(map[string]*User)

func fetchStockPrices(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data struct {
		Stocks []struct {
			Code  string  `json:"code"`
			Price float64 `json:"price"`
		} `json:"stocks"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	for _, s := range data.Stocks {
		stock, exists := stocks[s.Code]
		if exists {
			stock.CurrentPrice = s.Price
		} else {
			stocks[s.Code] = &Stock{
				StockCode:    s.Code,
				CurrentPrice: s.Price,
			}
		}
	}

	return nil
}

// BreakThrough 突破 / FallBelow 跌破
func checkAlarms() {
	for _, user := range users {
		for _, alarm := range user.Alarms {
			stock, exists := stocks[alarm.StockCode]
			if !exists {
				continue
			}
			switch alarm.Type {
			case "break":
				if stock.CurrentPrice >= alarm.Price {
					notifyUser(user, stock, alarm)
				}
			case "fall below":
				if stock.CurrentPrice <= alarm.Price {
					notifyUser(user, stock, alarm)
				}
			}
		}
	}
}

func notifyUser(user *User, stock *Stock, alarm *Alarm) {
	fmt.Printf("通知用户 %s：股票 %s 的价格 %f 触发了 %s 预警\n",
		user.Username, stock.StockName, stock.CurrentPrice, alarm.Type)
}

func main() {
	// 初始化股票
	stocks["AAPL"] = &Stock{StockCode: "AAPL", StockName: "Apple Inc.", CurrentPrice: 0.0}
	stocks["GOOGL"] = &Stock{StockCode: "GOOGL", StockName: "Alphabet Inc.", CurrentPrice: 0.0}

	// 初始化用户
	user1 := &User{
		UserID:        "u1",
		Username:      "Alice",
		WatchedStocks: make(map[string]*Stock),
		Alarms:        []*Alarm{},
	}
	user1.WatchedStocks["AAPL"] = stocks["AAPL"]
	user1.Alarms = append(user1.Alarms, &Alarm{Type: "break", Price: 155.0, User: user1, StockCode: "AAPL"})
	user1.Alarms = append(user1.Alarms, &Alarm{Type: "fall below", Price: 145.0, User: user1, StockCode: "AAPL"})
	users[user1.UserID] = user1

	// 添加更多用户和股票
	// ...

	// 获取股票价格
	err := fetchStockPrices("http://example.com/stocks")
	if err != nil {
		fmt.Println("无法获取股票价格:", err)
		return
	}

	// 检查预警
	checkAlarms()
}
