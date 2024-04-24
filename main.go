package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
	Protocol: 3,  // specify 2 for RESP 2 or 3 for RESP 3
})

func getUrl() string {
	file, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	words := ""
	scanner := bufio.NewScanner(file)
	for i := 0; i < 3; i++ {
		random := int64(rand.Intn(2252))
		file.Seek(5*random, 0)
		scanner.Scan()
		words += scanner.Text() + "-"
	}
	return words[:len(words)-1]
}

func writeToRedis(key string, value string) {
	rdb.Set(ctx, key, value, 0)
}

func getFromRedis(key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func setKey(w http.ResponseWriter, r *http.Request) {
	value := r.FormValue("value")
	url := getUrl()
	println(url, value)
	writeToRedis(url, value)
	fmt.Fprintf(w, "<html><body><h1>"+url+"<h1></body></html>")
}

func redirect(w http.ResponseWriter, r *http.Request) {
	words := r.URL.Path[1:]
	val, err := getFromRedis(words)

	if err != nil {
		io.WriteString(w, "Key not found\n"+words+"\n")
	} else {
		// redirect to the value of the key
		if val[:4] != "http" {
			val = "http://" + val
		}
		http.Redirect(w, r, val, http.StatusMovedPermanently)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 1 {
		redirect(w, r)
	} else {
		http.FileServer(http.Dir("./static/index.html"))
		t, _ := template.ParseFiles("./static/index.html")
		t.Execute(w, t)
	}
}

func main() {
	http.HandleFunc("/set", setKey)
	http.HandleFunc("/", handleGet)

	print("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
