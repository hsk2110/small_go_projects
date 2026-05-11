package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	statusCode int
	duration   time.Duration
	err        error
	rawURL     string
}

func checkURL(urlString string, results chan Result, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	start := time.Now()

	req, getErr := http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
	if getErr != nil {
		results <- Result{
			err:    getErr,
			rawURL: urlString,
		}
		return
	}

	resp, getErr := http.DefaultClient.Do(req)

	durr := time.Since(start)

	if getErr != nil {
		results <- Result{
			err:      getErr,
			duration: durr,
			rawURL:   urlString,
		}
		return
	}
	defer resp.Body.Close()
	results <- Result{
		statusCode: resp.StatusCode,
		duration:   durr,
		rawURL:     urlString,
	}
}

func main() {

	myURLs := []string{"https://google.com", "https://gobyexample.com", "https://youtube.com", "https://tum.de", "https://amazon.com"}

	results := make(chan Result, 5)
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, u := range myURLs {
		wg.Add(1)
		go checkURL(u, results, &wg, ctx)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for current_result := range results {
		if current_result.err != nil {
			fmt.Printf("Error: %v\n Duration: %v\n URL: %v\n\n", current_result.err, current_result.duration, current_result.rawURL)
		} else {
			fmt.Printf("Status Code: %d\n Duration: %v\n URL: %v\n\n", current_result.statusCode, current_result.duration, current_result.rawURL)
		}
	}

}
