package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var (
	Web1   = FakeSearch("web", "The Go Programming Language", "http://golang.org")
	Web2   = FakeSearch("web", "The Go Programming Language", "http://golang.org")
	Image1 = FakeSearch("image", "The Go gopher", "https://blog.golang.org/gopher/gopher.png")
	Image2 = FakeSearch("image", "The Go gopher", "https://blog.golang.org/gopher/gopher.png")
	Video1 = FakeSearch("video", "Concurrency Is Not Parallelism", "https://www.youtube.com/watch?v=qmg1CF3gZQ0")
	Video2 = FakeSearch("video", "Concurrency Is Not Parallelism", "https://www.youtube.com/watch?v=qmg1CF3gZQ0")
)

var (
	replicatedWeb   = First(Web1, Web2)
	replicatedImage = First(Image1, Image2)
	replicatedVideo = First(Video1, Video2)
)

type Result struct {
	Title, URL string
}

type SearchFunc func(query string) Result

func First(replicas ...SearchFunc) SearchFunc {
	return func(query string) Result {
		c := make(chan Result, len(replicas))
		searchReplica := func(i int) {
			c <- replicas[i](query)
		}

		for i := range replicas {
			go searchReplica(i)
		}

		return <-c
	}
}

func FakeSearch(kind, title, url string) SearchFunc {
	return func(query string) Result {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		return Result{
			Title: fmt.Sprintf("%s(%q): %s", kind, query, title),
			URL:   url,
		}
	}
}

func Search(query string, timeout time.Duration) ([]Result, error) {
	timer := time.After(timeout)
	c := make(chan Result, 3)

	go func() { c <- replicatedWeb(query) }()
	go func() { c <- replicatedImage(query) }()
	go func() { c <- replicatedVideo(query) }()

	var results []Result

	for i := 0; i < 3; i++ {
		select {
		case result := <-c:
			results = append(results, result)
		case <-timer:
			return results, errors.New("timed out")
		}
	}

	return results, nil
}

func main() {
	start := time.Now()
	results, err := Search("golang", 80*time.Millisecond)
	elapsed := time.Since(start)

	fmt.Println(results)
	fmt.Println(elapsed, err)
}