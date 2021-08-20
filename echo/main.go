package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Echo struct {
}

func (e Echo) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	all, _ := ioutil.ReadAll(request.Body)
	m := make(map[string]interface{}, 0)
	m["body"] = all
	m["uri"] = request.URL.String()
	m["header"] = request.Header
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(200)
	_ = json.NewEncoder(writer).Encode(m)
}

func main() {
	log.Println("It works")
	log.Fatal(http.ListenAndServe(":8080", Echo{}))
}
