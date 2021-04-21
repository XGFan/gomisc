package main

import (
	"fmt"
	"math"
)

func main() {
	//repayment()
	AverageCapitalPlusInterest(2700000, 5.68, 360)
	AverageInterest(2700000, 5.68, 360)
}
func AverageInterest(total, rate float32, months int) {
	fixed := total / float32(months)
	for i := 1; i <= months; i++ {
		shouldPay := fixed + (total * rate / 100 / 12)
		total -= fixed
		fmt.Printf("[%d]Pay: %.2f, Total:%.2f\n", i, shouldPay, total)
	}
}

func AverageCapitalPlusInterest(total, rate float64, months int) {
	rateMonthly := rate / 12 / 100
	pay := total * rateMonthly * math.Pow(rateMonthly+1, float64(months)) / (math.Pow(rateMonthly+1, float64(months)) - 1)
	fmt.Printf("Pay: %.2f\n", pay)
}

func repayment() int {
	total := 700000.0
	rate := 3.25
	account := 200000.0
	income := 9000.0
	var i int
	f := total / 360
	for i = 1; account <= total; i++ {
		shouldPay := f + (total * rate / 100 / 12)
		total -= total / 360
		account += income - shouldPay
		fmt.Println(i, "total:", total, "should", shouldPay, "account:", account)
		if i%12 == 0 {
			total -= account - 100
			account = 100
			fmt.Println("Reload", "total:", total, "account:", account)
		}
	}
	return i
}
