package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockDB struct {
	aggregateFeeByHour func() ([]AggregatedFee, error)
}

func (m mockDB) AggregateFeeByHour() ([]AggregatedFee, error) {
	return m.aggregateFeeByHour()
}

func TestHandler(t *testing.T) {
	successExample := []AggregatedFee{
		{time.Now().Unix(), 10.5},
		{time.Now().Unix(), 9.3},
	}

	successB, err := json.MarshalIndent(&successExample, "", "  ")
	if err != nil {
		t.Fatal("error initializing testdata", err)
	}
	successStr := string(successB) + "\n"

	tt := []struct {
		description      string
		httpMethod       string
		expectedStatus   int
		expectedResponse string
		aggregateFunc    func() ([]AggregatedFee, error)
	}{
		{
			description:      "should return method not allowed for method not GET",
			httpMethod:       http.MethodPost,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedResponse: "Invalid HTTP Method, only HTTP GET is allowed\n",
		},
		{
			description:      "should return internal server error for db err",
			httpMethod:       http.MethodGet,
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: "Error communicating with datasource\n",
			aggregateFunc: func() ([]AggregatedFee, error) {
				return []AggregatedFee{}, errors.New("db error")
			},
		},
		{
			description:      "should return OK and data",
			httpMethod:       http.MethodGet,
			expectedStatus:   http.StatusOK,
			expectedResponse: successStr,
			aggregateFunc: func() ([]AggregatedFee, error) {
				return successExample, nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			req, err := http.NewRequest(tc.httpMethod, "", nil)
			if err != nil {
				t.Fatal("error building request", err)
			}
			db := mockDB{tc.aggregateFunc}
			rr := httptest.NewRecorder()
			h := handler{db}
			h.ServeHTTP(rr, req)

			if rr.Result().StatusCode != tc.expectedStatus {
				t.Fatalf("wrong status code, want: %v, got: %v", tc.expectedStatus, rr.Result().StatusCode)
			}

			gotData, err := ioutil.ReadAll(rr.Body)
			if err != nil {
				t.Fatal("error reading response", err)
			}

			if string(gotData) != tc.expectedResponse {
				t.Fatalf("wrong response body\nwant: %v\ngot: %v", tc.expectedResponse, string(gotData))
			}
		})
	}
}
