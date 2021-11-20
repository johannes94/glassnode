package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type AggregatedFee struct {
	Hour      int64   `json:"t"`
	HourlyFee float64 `json:"v"`
}

type ethDB interface {
	AggregateFeeByHour() ([]AggregatedFee, error)
}

var query string = `
SELECT CAST(extract(EPOCH FROM date_trunc('hour', sub.ts)) AS INT) AS hour, SUM(gas_payed)* 10 ^ -18 AS hourly_fee FROM 
	(SELECT t.gas_used*t.gas_price AS gas_payed, t.block_time AS ts FROM 
		transactions AS t LEFT JOIN contracts AS c
		ON t.from = c.address OR t.to = c.address
		WHERE c.address IS NULL
) AS sub 
GROUP BY hour;
`

type psqlDB struct {
	con *sql.DB
}

func (pDB psqlDB) AggregateFeeByHour() (result []AggregatedFee, err error) {
	rows, err := pDB.con.Query(query)
	if err != nil {
		return
	}

	for rows.Next() {
		dest := AggregatedFee{}
		err = rows.Scan(&dest.Hour, &dest.HourlyFee)
		if err != nil {
			return
		}

		result = append(result, dest)
	}

	return
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
	rw.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(&data); err != nil {
		log.Println(err)
		http.Error(rw, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

}

func main() {

	host := os.Getenv("ETH_DB_HOST")
	user := os.Getenv("ETH_DB_USER")
	pwd := os.Getenv("ETH_DB_PASSWORD")
	dbname := os.Getenv("ETH_DB_NAME")

	cStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", user, dbname, pwd, host)
	con, err := sql.Open("postgres", cStr)
	if err != nil {
		log.Fatal("error connecting to database", err)
	}

	h := handler{psqlDB{con}}

	log.Println("API starts listening")
	if err := http.ListenAndServe(":8080", h); err != nil {
		log.Println("Error starting webserver: ", err)
	}
}