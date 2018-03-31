
package main

import (
	"os"
	"encoding/json"
	"fmt"
	"strconv"
	"net/http"
)

import "time"
import "math/rand"

type Configuration struct {
	Suffix string `json:"suffix"`
}

func hexFromInt64(st int64) string {
	encode := strconv.FormatInt(st,16)
	if len(encode) < 16 {
		ori := "0000000000000000"
		encode = ori[:16-len(encode)] + encode
	}
	
	return encode
}

func currentTime() int64 {

	var seconds int64
	var micro_seconds int64

	now := time.Now()
	nano := now.UnixNano()

	micro_seconds = nano / 1000000;
	seconds = micro_seconds / 1000;
	micro_seconds = micro_seconds % 1000;

	return ((seconds & 0xffffffff) << 16) | (micro_seconds & 0xff)
}

func idGernerater(in chan int64, out chan string, suffix string) {
	var offset int64 = 0
	var timestamp int64 = 0

	for nano := range in {
		nano = currentTime()
		if nano > timestamp {
			offset = rand.Int63n(65536)
			seq := hexFromInt64((nano << 16) | offset)
			out <-  seq + suffix

			timestamp = nano
			offset += 1
			if offset > 65536 {
				timestamp += 1
				offset = 0
			}
		} else {
			seq := hexFromInt64((nano << 16) | offset)
			out <-  seq + suffix

			offset += 1
			if offset >= 65536 {
				timestamp += 1
				offset = 0
			}
		}
	}

}

func welcome(w http.ResponseWriter, r *http.Request) {
	in <- 1
	id := <- out
	fmt.Fprintf(w, "%s\n", id)
}

var in chan int64
var out chan string

func init() {
	in = make(chan int64)
	out = make(chan string)
}

func main() {
	var conf Configuration

	file, _ := os.Open("conf.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("read config file error:", err)
		return
	}

	go idGernerater(in, out, conf.Suffix)

	http.HandleFunc("/", welcome)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("start go server error!")
	}
}
