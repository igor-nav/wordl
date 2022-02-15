package main

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "strings"
  "regexp"
)

// fetch returns the contents of a webpage.
func fetch(path string) []byte {
  url := "https://nytimes.com/games/wordle/" + path
  resp, err := http.Get(url)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  html, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }
  return html
}

// findJs finds the name of the JavaScript file in the HTML.
func findJs(html []byte) string {
  re := regexp.MustCompile(`<script src="(main.[^.]+.js)">`)
  m := re.FindSubmatch(html)
  if len(m) < 2 {
    panic("No <script> found in index.html")
  }
  return string(m[1])
}

// findArray finds an array of strings in JavaScript.
func findArray(js []byte, name string) []string {
  re := regexp.MustCompile(fmt.Sprintf(`[, ]%s=\["([a-z",]+)"\]`, name))
  m := re.FindSubmatch(js)
  if len(m) < 2 {
    panic(fmt.Sprintf("Cannot find an array called %s in JS", name))
  }
  return strings.Split(string(m[1]), `","`)
}

// formatArray converts an array of strings into file contents.
func formatArray(arr []string) []byte {
  return append([]byte(strings.Join(arr, "\n")), '\n')
}

func main() {
  indexHtml := fetch("index.html")
  js := fetch(findJs(indexHtml))
  ma := findArray(js, "Ma")
  oa := findArray(js, "Oa")
  fmt.Printf("len(Ma) = %d; len(Oa) = %d;\n", len(ma), len(oa))
  os.WriteFile("../words1.txt", formatArray(ma), 0644)
  os.WriteFile("../words2.txt", formatArray(oa), 0644)
}
