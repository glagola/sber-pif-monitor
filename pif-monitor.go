package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func processError(err error) {
	if err != nil {
		panic(err)
	}
}

func getUrl(fund int, startDate, endDate string) string {
	const URL = "http://www.sberbank-am.ru/visible/chart/complexFundRates?fund=%d&startDate=%s&endDate=%s"

	return fmt.Sprintf(URL, fund, startDate, endDate)
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func getPeriod(startDate time.Time) (string, string) {
	return formatDate(startDate), formatDate(time.Now())
}

func fixJson(jsonBlob []byte) []byte {
	s := string(jsonBlob)

	s = strings.Replace(s, "(", "", -1)
	s = strings.Replace(s, "date", `"date"`, -1)
	s = strings.Replace(s, "price", `"price"`, -1)
	s = strings.Replace(s, "dollar", `"dollar"`, -1)
	s = strings.Replace(s, ");", "", -1)

	return []byte(s)
}

func getJson(url string) (jsn []byte) {
	res, err := http.Get(url)
	defer res.Body.Close()

	processError(err)
	jsn, err = ioutil.ReadAll(res.Body)
	processError(err)

	return
}

type Quote struct {
	Date  time.Time
	Price float64
}

func parseQuotes(jsonBlob []byte) []Quote {

	type quote struct {
		Date  int64   `json:"date"`
		Price float64 `json:"price"`
	}

	var quotes []quote
	err := json.Unmarshal(jsonBlob, &quotes)
	processError(err)

	var res []Quote

	for _, q := range quotes {
		res = append(res, Quote{Date: time.Unix(q.Date/1000, 0), Price: q.Price})
	}

	return res
}

func getStartEndPrice(quotes []Quote) (float64, float64) {
	startQuote, endQuote := quotes[0], quotes[0]

	for _, q := range quotes {
		if q.Date.Before(startQuote.Date) {
			startQuote = q
		}

		if q.Date.After(endQuote.Date) {
			endQuote = q
		}
	}

	return startQuote.Price, endQuote.Price
}

type Investment struct {
	Fund      int
	Purchased time.Time
	shares    float64
}

func main() {
	purchaseDate, _ := time.Parse("02.01.2006", "26.02.2014")
	purchaseDate2, _ := time.Parse("02.01.2006", "09.06.2014")

	var funds = map[int]string{
		12: "Телекоммуникации и Технологии",
		16: "Потребительский сектор",
		29: "Глобальный Интернет",
	}

	var config = []Investment{
		{Fund: 12, Purchased: purchaseDate, shares: 4.9607604},
		{Fund: 29, Purchased: purchaseDate, shares: 21.8402868},
		{Fund: 16, Purchased: purchaseDate, shares: 12.691966},
		{Fund: 12, Purchased: purchaseDate2, shares: 17.703110},
	}

	total := .0

	for _, inv := range config {
		startDate, endDate := getPeriod(inv.Purchased)
		startPrice, endPrice := getStartEndPrice(parseQuotes(fixJson(getJson(getUrl(inv.Fund, startDate, endDate)))))

		profit := (endPrice - startPrice) * inv.shares
		total += profit

		fmt.Printf("%s: %.02f\r\n", funds[inv.Fund], profit)
	}

	fmt.Print("---------------------------------------------\r\n")
	fmt.Printf("Итого: %.02f\r\n", total)
}
