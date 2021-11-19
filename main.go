package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AggregatedFee struct {
	Timestamp int64   `json:"t"`
	Fee       float64 `json:"v"`
}

type ethDB interface {
	AggregateFeeByHour() ([]AggregatedFee, error)
}

type inmemoryDB struct {
	data []AggregatedFee
}

func (iDB inmemoryDB) AggregateFeeByHour() ([]AggregatedFee, error) {
	return iDB.data, nil
}

type handler struct {
	db ethDB
}

func (h handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(rw, "Invalid HTTP Method, only HTTP GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := h.db.AggregateFeeByHour()
	if err != nil {
		log.Println(err)
		http.Error(rw, "Could not reach datasource", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(rw)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&data); err != nil {
		log.Println(err)
		http.Error(rw, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
}

func main() {
	data := []AggregatedFee{
		{time.Now().Unix(), 10.5},
		{time.Now().Unix(), 11.244},
	}

	h := handler{inmemoryDB{data}}

	if err := http.ListenAndServe(":8080", h); err != nil {
		log.Println("Error starting webserver: ", err)
	}

}
