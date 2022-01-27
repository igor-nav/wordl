package main

import (
  _ "embed"
  "fmt"
  "strings"
)

//go:embed 5-letter-words.txt
var data string

var allWords []string

func init() {
  allWords = strings.Split(data, "\n")
}

func main() {
  fmt.Printf("Starting with a dictionary of %d words\n", len(allWords))
}
