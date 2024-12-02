package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"sync"
	"math/rand"

	"github.com/temoto/robotstxt" // Go library for parsing robots.txt
	"golang.org/x/net/html"
)

type Sitemap struct {
	URL   string    `json:"url"`
	Links []*Sitemap `json:"links"`
}

var visited = make(map[string]bool)
var mu sync.Mutex // Mutex to ensure thread safety for visited map

// List of user agents to rotate
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.0",
	"Mozilla/5.0 (Windows NT 6.1; rv:34.0) Gecko/20100101 Firefox/34.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
}

var sem = make(chan struct{}, 10) // Limit concurrency

// Maximum number of concurrent requests to be made
const maxConcurrentRequests = 10

func getRandomUserAgent() string {
	// Get a random index from the userAgents slice
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	return userAgents[rand.Intn(len(userAgents))]
}

func crawl(startURL string, maxDepth, depth int, wg *sync.WaitGroup) *Sitemap {
	defer wg.Done() // Decrement the counter when done
	if depth > maxDepth {
		log.Printf("Reached max depth for: %s (Depth: %d)\n", startURL, depth)
		return nil
	}

	normalizedURL := normalizeURL(startURL)
	mu.Lock()
	if visited[normalizedURL] {
		mu.Unlock()
		log.Printf("Already visited: %s (Depth: %d)\n", normalizedURL, depth)
		return nil
	}
	visited[normalizedURL] = true
	mu.Unlock()

	log.Printf("Crawling URL: %s (Depth: %d)\n", normalizedURL, depth)

	// Check robots.txt
	domain := getDomain(normalizedURL)
	if !isAllowedByRobots(domain, normalizedURL) {
		log.Printf("Blocked by robots.txt: %s\n", normalizedURL)
		return nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", normalizedURL, nil)
	if err != nil {
		log.Printf("Error creating request for URL %s: %v\n", normalizedURL, err)
		return nil
	}

	// Set a random User-Agent
	req.Header.Set("User-Agent", getRandomUserAgent())

	// Limit concurrency
	sem <- struct{}{} // Acquire a slot
	resp, err := client.Do(req)
	<-sem // Release the slot
	if err != nil {
		log.Printf("Error fetching URL %s: %v\n", normalizedURL, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s (%d)\n", resp.Status, resp.StatusCode)
		return nil
	}

	sitemap := &Sitemap{URL: normalizedURL}
	tokenizer := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			log.Printf("Finished parsing: %s (Depth: %d)\n", normalizedURL, depth)
			return sitemap
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						absoluteLink := resolveURL(normalizedURL, link)
						if absoluteLink != "" && isSameDomain(normalizedURL, absoluteLink) {
							wg.Add(1) // Increment the WaitGroup counter
							go func(link string) {
								child := crawl(link, 2, depth+1, wg)
								if child != nil {
									sitemap.Links = append(sitemap.Links, child)
								}
							}(absoluteLink)
						}
					}
				}
			}
		}
	}
}

func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Error parsing URL: %s, %v\n", rawURL, err)
		return rawURL
	}
	u.Fragment = ""
	u.RawQuery = ""
	return strings.TrimRight(u.String(), "/")
}

func resolveURL(baseURL, href string) string {
	u, err := url.Parse(href)
	if err != nil {
		log.Printf("Error parsing URL: %s, %v\n", href, err)
		return ""
	}
	if u.IsAbs() {
		return href
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing base URL: %s, %v\n", baseURL, err)
		return ""
	}
	return base.ResolveReference(u).String()
}

func isSameDomain(baseURL, checkURL string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing base URL: %s, %v\n", baseURL, err)
		return false
	}
	check, err := url.Parse(checkURL)
	if err != nil {
		log.Printf("Error parsing check URL: %s, %v\n", checkURL, err)
		return false
	}
	return base.Hostname() == check.Hostname()
}

func isAllowedByRobots(domain, path string) bool {
    robotsURL := fmt.Sprintf("%s/robots.txt", domain)
    resp, err := http.Get(robotsURL)
    if err != nil || resp.StatusCode != http.StatusOK {
        log.Printf("robots.txt not found or inaccessible for domain %s: assuming allowed\n", domain)
        return true // Assume allowed if robots.txt is inaccessible
    }
    defer resp.Body.Close()

    robotsData, err := robotstxt.FromResponse(resp)
    if err != nil {
        log.Printf("Error parsing robots.txt for domain %s: %v. Assuming allowed.\n", domain, err)
        return true
    }

    allowed := robotsData.TestAgent(path, "Mozilla/5.0")
    if !allowed {
        log.Printf("Blocked by robots.txt: %s\n", path)
    }
    return allowed
}

func getDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Error parsing URL: %s, %v\n", rawURL, err)
		return rawURL
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Println("Received request:", r.URL)

	query := r.URL.Query()
	startURL := query.Get("url")
	if startURL == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		log.Println("Error: Missing url parameter")
		return
	}

	// Starting the crawl process
	log.Printf("Starting crawl for URL: %s\n", startURL)
	var wg sync.WaitGroup
	wg.Add(1)
	sitemap := crawl(startURL, 2, 0, &wg)
	wg.Wait() // Wait for all crawlers to finish

	// Returning the sitemap as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sitemap)
	log.Printf("Crawling completed in %v\n", time.Since(startTime))
}

func main() {
	http.HandleFunc("/crawl", handler)
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
