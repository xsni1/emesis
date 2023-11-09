package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "golang.org/x/time/rate"
)

type RateLimiter struct{}

// global rate limits?
func (rl *RateLimiter) Wait(b *Bucket, cost int) {
	// if we don't have enough "requests" in the buckets
	// we have to wait until we do
	fmt.Println("WAIT FUNC")
	fmt.Println("bucket remaining: ", b.remaining, "cost: ", cost)
	fmt.Println("reset after: ", b.resetTime, "now: ", time.Now())
	if b.remaining < cost && b.resetTime.After(time.Now()) {
		fmt.Println("waiting")
		time.Sleep(b.resetTime.Sub(time.Now()))
		return
	}
	fmt.Println("not waiting")

	return
}

type HttpClient struct {
	buckets     map[string]*Bucket
	rateLimiter *RateLimiter
}

type Bucket struct {
	remaining int
	resetTime time.Time
}

func (c *HttpClient) getOrCreateBucket(key string) *Bucket {
	if b, ok := c.buckets[key]; ok {
		return b
	}

	b := &Bucket{
		remaining: 1,
	}

	c.buckets[key] = b

	return b
}

func (c *HttpClient) Do(url string) {
	fmt.Println("===========================")
	fmt.Println("==== REQUEST")

	b := c.getOrCreateBucket(url)

	c.rateLimiter.Wait(b, 1)

	// res, _ := http.Get(url)
	httpClient := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bot OTEyODUxMTQ1Mjg0Nzg0MTM5.YZ184A.23WLcn5A0Kgrbhjah2ABoc8RDM0")
	req.Header.Set("Accept", "application/json")

	res, _ := httpClient.Do(req)
	bytes, _ := io.ReadAll(res.Body)

	fmt.Println(string(bytes))

	resetAfter := res.Header.Get("X-RateLimit-Reset-After")
	remaining := res.Header.Get("X-RateLimit-Remaining")
	parsedResetAfter, _ := strconv.ParseFloat(resetAfter, 64)
	parsedRemaining, _ := strconv.Atoi(remaining)
	seconds, miliseconds := math.Modf(parsedResetAfter)

	fmt.Println("pre parsing remaining: ", remaining, "reset after: ", resetAfter)
	fmt.Println("remaining: ", parsedRemaining, "reset after: ", parsedResetAfter, "secs: ", seconds, "milis: ", miliseconds)

	b.resetTime = time.Now().Add(time.Duration(seconds) * time.Second).Add(time.Duration(miliseconds) * time.Millisecond)
	b.remaining = parsedRemaining

	fmt.Println(b.resetTime, b.remaining)
}

func main() {
	godotenv.Load()

	httpClient := HttpClient{
		buckets: make(map[string]*Bucket),
	}
	for {
		httpClient.Do("https://discord.com/api/v10/channels/445681005475397643/messages?after=523587110851182617&limit=2")
	}
}
