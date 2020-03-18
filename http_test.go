package simplehttp

import (
	"fmt"
	"testing"
)

func TestClient_Get(t *testing.T) {
	httpClient := NewClient("https://samples.openweathermap.org/")
	query := map[string]interface{}{
		"q":     "London,uk",
		"appid": "b6907d289e10d714a6e88b30761fae22",
	}
	res, err := httpClient.Get("/data/2.5/weather", query, nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(res))

	httpClient = NewClient("https://samples.openweathermap.org/data/2.5/weather")
	res, err = httpClient.Get("", query, nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(res))
}
