package main

import (
	"context"
	"io"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func writeToRedis(key string, value string) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		Protocol: 3,  // specify 2 for RESP 2 or 3 for RESP 3
	})

	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func getFromRedis(key string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		Protocol: 3,  // specify 2 for RESP 2 or 3 for RESP 3
	})

	return rdb.Get(ctx, key).Result()
}

func setKey(w http.ResponseWriter, r *http.Request) {
	// read from formdata
	key := r.FormValue("key")
	value := r.FormValue("value")
	writeToRedis(key, value)
	newValue, err := getFromRedis(key)
	if err != nil {
		io.WriteString(w, "Key set failed\n"+key+" : "+value+"\n"+err.Error())
	}
	io.WriteString(w, "Key set successfully\n"+key+" : "+newValue)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	val, err := getFromRedis(r.URL.Path[5:])

	if err != nil {
		io.WriteString(w, "Key not found\n"+r.URL.Path[5:]+"\n"+err.Error())
	} else {
		// redirect to the value of the key
		http.Redirect(w, r, val, http.StatusMovedPermanently)
	}
}

func main() {
	http.HandleFunc("/set", setKey)
	http.HandleFunc("/api/", redirect)
	print("Server is running on port 8080")
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":8080", nil)
}
