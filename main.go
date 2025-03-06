package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var methodRegex = regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH) `)

type Request struct {
	Method string `json:"method"`
	Header []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"header,omitempty"`
	Body struct {
		Mode string `json:"mode"`
		Raw  string `json:"raw,omitempty"`
	} `json:"body,omitempty"`
	URL struct {
		Raw      string   `json:"raw"`
		Protocol string   `json:"protocol,omitempty"`
		Host     []string `json:"host,omitempty"`
		Path     []string `json:"path,omitempty"`
	} `json:"url"`
}

type Item struct {
	Name    string  `json:"name"`
	Request Request `json:"request"`
}

type PostmanCollection struct {
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Item []Item `json:"item"`
}

func parseURL(url string) (structuredURL struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol,omitempty"`
	Host     []string `json:"host,omitempty"`
	Path     []string `json:"path,omitempty"`
},
) {
	structuredURL.Raw = url
	if strings.HasPrefix(url, "http://") {
		structuredURL.Protocol = "http"
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		structuredURL.Protocol = "https"
		url = strings.TrimPrefix(url, "https://")
	}
	parts := strings.Split(url, "/")
	structuredURL.Host = []string{parts[0]}
	if len(parts) > 1 {
		structuredURL.Path = parts[1:]
	}
	return
}

func parseHTTPFile(filename string) (PostmanCollection, error) {
	file, err := os.Open(filename)
	if err != nil {
		return PostmanCollection{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	postman := PostmanCollection{
		Info: struct {
			Name   string `json:"name"`
			Schema string `json:"schema"`
		}{
			Name:   "Generated Collection",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
	}

	var method, url string
	var headers []map[string]string
	var requestBody strings.Builder
	isBody := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "###") {
			continue
		}

		if methodRegex.MatchString(line) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				method, url = parts[0], parts[1]
				headers = nil
				isBody = false
				requestBody.Reset()
			}
		} else if strings.Contains(line, ":") && !isBody {
			parts := strings.SplitN(line, ":", 2)
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			headers = append(headers, map[string]string{"key": key, "value": value})
		} else if strings.HasPrefix(line, "{") {
			isBody = true
			requestBody.WriteString(line)
		} else if isBody {
			requestBody.WriteString(line)
		}

		if method != "" && url != "" {
			item := Item{
				Name: url,
				Request: Request{
					Method: method,
					URL:    parseURL(url),
				},
			}

			for _, h := range headers {
				item.Request.Header = append(item.Request.Header, struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}{Key: h["key"], Value: h["value"]})
			}

			if requestBody.Len() > 0 {
				item.Request.Body.Mode = "raw"
				item.Request.Body.Raw = requestBody.String()
			}

			postman.Item = append(postman.Item, item)
			method, url, headers = "", "", nil
			requestBody.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return PostmanCollection{}, err
	}

	return postman, nil
}

func main() {
	collection, err := parseHTTPFile("requests.http")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	output, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = os.WriteFile("postman_collection.json", output, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Postman collection generated: postman_collection.json")
}
