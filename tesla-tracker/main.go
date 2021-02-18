package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-rod/rod"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ctx = context.Background()

func Notify(str string) {
	client := http.Client{}
	_, err := client.Get("https://api.day.app/secret/" + str)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	page := rod.New().MustConnect().MustPage("https://www.tesla.cn/model3/design#battery")
	el := page.MustWaitLoad().MustElements(".group--options_block_title")
	all := rdb.HGetAll(ctx, "tesla").Val()
	newMap := make(map[string]interface{}, 0)
	for _, element := range el {
		carType := element.MustElement(".group--options_block--name").MustText()
		priceStr := element.MustElement(".price-not-included").MustText()
		if priceStr == "" || carType == "" {
			log.Fatal("can't get type or price", element)
		}
		priceStr = strings.ReplaceAll(priceStr, ",", "")
		priceStr = strings.ReplaceAll(priceStr, "¥", "")
		priceStr = strings.ReplaceAll(priceStr, "*", "")
		priceStr = strings.TrimSpace(priceStr)
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			log.Fatal(err)
		}
		if all[carType] == "" {
			Notify(fmt.Sprintf("New Car [%s], Price: %d", carType, price))
		} else {
			prevPrice, _ := strconv.Atoi(all[carType])
			if prevPrice != price {
				Notify(fmt.Sprintf("Car [%s] Price From %d to %d", carType, prevPrice, price))
			}
		}
		delete(all, carType)
		newMap[carType] = priceStr
	}
	for carType, price := range all {
		Notify(fmt.Sprintf("Car [%s](%s) Has Been Removed", carType, price))
	}
	_, err := rdb.HSet(ctx, "tesla", newMap).Result()
	if err != nil {
		log.Fatal(err)
	}
}
