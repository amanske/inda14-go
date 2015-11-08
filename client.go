package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	server := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
	for {
		before := time.Now()
		//res := Get(server[0])
		//res := Read(server[0], time.Second)
		res := MultiRead(server, time.Second)
		after := time.Now()
		fmt.Println("Response:", *res)
		fmt.Println("Time:", after.Sub(before))
		fmt.Println()
		time.Sleep(500 * time.Millisecond)
	}
}

type Response struct {
	Body       string
	StatusCode int
}

// Get makes an HTTP Get request and returns an abbreviated response.
// Status code 200 means that the request was successful.
// The function returns &Response{"", 0} if the request fails
// and it blocks forever if the server doesn't respond.
func Get(url string) *Response {
	res, err := http.Get(url)
	if err != nil {
		return &Response{}
	}
	// res.Body != nil when err == nil
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	return &Response{string(body), res.StatusCode}
}

//This method calls the function Get() and lets a go routine run
//until a response is available (or it has timed out).
//
//Bug 1: If we look at Get() and Read(), they both have access to
//the poiner to 'res' (type Response). This could result in an error
//if timeout would occur,since no one would be there to recieve on the channel.
//We can fix this by giving the channel a buffer that could store a Response
//if no one were the to recieve it. This also solves bug 2 which was that the
//routines could not terminate if timeout was hit.
func Read(url string, timeout time.Duration) (res *Response) {
	done := make(chan *Response, 1)
	go func() {
		done <- Get(url) //buffer allows to terminate even if no one reads
	}()
	select {
	case res = <-done:
	case <-time.After(timeout):
		res = &Response{"Gateway timeout\n", 504}
	}
	return
}

// MultiRead makes an HTTP Get request to each url and returns
// the response of the first server to answer with status code 200.
// If none of the servers answer before timeout, the response is
// 503 â€“ Service unavailable.
func MultiRead(urls []string, timeout time.Duration) (res *Response) {
	resChan := make(chan *Response, len(urls))

	for _, url := range urls {
		go func() {
			read := Read(url, timeout)
			if read.StatusCode == 200 {
				resChan <- read
			}
		}()
	}

	select {
	case res = <-resChan:
	case <-time.After(timeout):
		res = &Response{"Service unavailable\n", 503}
	}

	return
}
