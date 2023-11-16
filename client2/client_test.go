/*
Copyright © 2019 Kondukto
*/
package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/kondukto-io/kdt/klog"
)

func TestNewRequestPath(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	req, err := client.newRequest("GET", "/test/path", nil)
	if err != nil {
		t.Fatal(err)
	}

	got := req.URL.String()
	expected := "http://localhost:8080/test/path"
	if got != expected {
		t.Fatalf("wrong request url: expected: %s got: %s", expected, got)
	}
}

func TestDo(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://jsonplaceholder.typicode.com")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = u

	req, err := client.newRequest("GET", "/users/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	type user struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Address  struct {
			Street  string `json:"street"`
			Suite   string `json:"suite"`
			City    string `json:"city"`
			Zipcode string `json:"zipcode"`
			Geo     struct {
				Lat string `json:"lat"`
				Lng string `json:"lng"`
			} `json:"geo"`
		} `json:"address"`
		Phone   string `json:"phone"`
		Website string `json:"website"`
		Company struct {
			Name        string `json:"name"`
			CatchPhrase string `json:"catchPhrase"`
			Bs          string `json:"bs"`
		} `json:"company"`
	}

	var someone user

	resp, err := client.do(req, &someone)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		klog.Fatalf("HTTP response status code: %d", resp.StatusCode)
		t.Fatal("HTTP response not OK")
	}

	j, err := json.Marshal(&someone)
	if err != nil {
		t.Fatal(err)
	}

	got := string(j)
	expected := `{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    "street": "Kulas Light",
    "suite": "Apt. 556",
    "city": "Gwenborough",
    "zipcode": "92998-3874",
    "geo": {
      "lat": "-37.3159",
      "lng": "81.1496"
    }
  },
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org",
  "company": {
    "name": "Romaguera-Crona",
    "catchPhrase": "Multi-layered client-server neural-net",
    "bs": "harness real-time e-markets"
  }
}`

	// Converting the expected string to compact JSON to ease comparison
	buf := new(bytes.Buffer)
	if err := json.Compact(buf, []byte(expected)); err != nil {
		t.Fatal()
	}
	expectedCompact := buf.String()

	if got != expectedCompact {
		t.Fatal("incorrect response body")
	}
}
