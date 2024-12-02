package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url" // Correctly imported
	"strings"
	"golang.org/x/net/html"
)

type PageData struct {
	URL      string
	Sitemap  []string
	HasError bool
	ErrorMsg string
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/crawl", crawlHandler)

	// Start the server
	log.Println("Starting server on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

// homeHandler renders the HTML UI
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Crawler</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        h1 { text-align: center; }
        input, button { padding: 10px; margin: 10px; width: 300px; }
        pre { background-color: #f4f4f4; padding: 10px; }
    </style>
</head>
<body>
    <h1>Web Crawler</h1>
    <form action="/crawl" method="get">
        <label for="url">Enter URL to crawl:</label><br>
        <input type="text" id="url" name="url" value="{{.URL}}" required>
        <button type="submit">Start Crawling</button>
    </form>
    {{if .HasError}}
        <div style="color: red;">Error: {{.ErrorMsg}}</div>
    {{end}}
    {{if .Sitemap}}
        <h3>Crawled URLs:</h3>
        <pre>{{range .Sitemap}}{{.}}
        {{end}}</pre>
    {{end}}
</body>
</html>
`))

	tmpl.Execute(w, nil)
}

// crawlHandler handles the crawling of the provided URL and returns the sitemap
func crawlHandler(w http.ResponseWriter, r *http.Request) {
	urlStr := r.URL.Query().Get("url")
	pageData := PageData{URL: urlStr}

	if urlStr == "" {
		pageData.HasError = true
		pageData.ErrorMsg = "URL is required."
		renderTemplate(w, pageData)
		return
	}

	// Normalize and validate the root URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		pageData.HasError = true
		pageData.ErrorMsg = "Invalid URL format."
		renderTemplate(w, pageData)
		return
	}

	// Initialize a set to avoid duplicate URLs
	visited := make(map[string]bool)
	var sitemap []string

	// Start crawling the provided URL
	err = crawlPage(parsedURL.String(), visited, &sitemap)
	if err != nil {
		pageData.HasError = true
		pageData.ErrorMsg = fmt.Sprintf("Error crawling URL: %v", err)
		renderTemplate(w, pageData)
		return
	}

	pageData.Sitemap = sitemap
	renderTemplate(w, pageData)
}

// renderTemplate renders the HTML template with the page data
func renderTemplate(w http.ResponseWriter, data PageData) {
	tmpl := template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Crawler</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        h1 { text-align: center; }
        input, button { padding: 10px; margin: 10px; width: 300px; }
        pre { background-color: #f4f4f4; padding: 10px; }
    </style>
</head>
<body>
    <h1>Web Crawler</h1>
    <form action="/crawl" method="get">
        <label for="url">Enter URL to crawl:</label><br>
        <input type="text" id="url" name="url" value="{{.URL}}" required>
        <button type="submit">Start Crawling</button>
    </form>
    {{if .HasError}}
        <div style="color: red;">Error: {{.ErrorMsg}}</div>
    {{end}}
    {{if .Sitemap}}
        <h3>Crawled URLs:</h3>
        <pre>{{range .Sitemap}}{{.}}
        {{end}}</pre>
    {{end}}
</body>
</html>
`))

	tmpl.Execute(w, data)
}

// crawlPage crawls a single page and collects links
func crawlPage(pageURL string, visited map[string]bool, sitemap *[]string) error {
	// If we've already visited this page, skip it
	if visited[pageURL] {
		return nil
	}
	visited[pageURL] = true

	// Fetch the page content
	resp, err := http.Get(pageURL)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %v", err)
	}
	defer resp.Body.Close()

	// Parse the HTML content
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Extract links from the page
	err = extractLinks(doc, pageURL, visited, sitemap)
	if err != nil {
		return fmt.Errorf("failed to extract links: %v", err)
	}

	return nil
}

// extractLinks extracts all the links from a parsed HTML document
func extractLinks(n *html.Node, baseURL string, visited map[string]bool, sitemap *[]string) error {
	// If the current node is an anchor <a> tag, extract the href attribute
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				link := attr.Val
				// If it's a relative URL, resolve it to an absolute URL
				if !strings.HasPrefix(link, "http") {
					link = baseURL + link
				}

				// Add the link to the sitemap if it's not visited
				if !visited[link] {
					visited[link] = true
					*sitemap = append(*sitemap, link)
					// Recursively crawl the linked page
					err := crawlPage(link, visited, sitemap)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Continue traversing the HTML tree recursively
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := extractLinks(c, baseURL, visited, sitemap)
		if err != nil {
			return err
		}
	}

	return nil
}
