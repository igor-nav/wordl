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

// Score returns the response to a guess, as a 5-digit integer.
func Score(target, guess string) int {
  used, ans := 0, 0
  for j, m := 0, 10000; j < 5; j, m = j + 1, m / 10 {
    if guess[j] == target[j] {
      ans += m * 2
      used |= (1 << j)
      continue
    }
    for i := 0; i < 5; i++ {
      if (used & (1 << i)) == 0 && target[i] == guess[j] && target[i] != guess[i] {
        ans += m
        used |= (1 << i)
        break
      }
    }
  }
  return ans
}

func TestScore() {
  if x := Score("mount", "strew"); x != 1000 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Score("mount", "topaz"); x != 12000 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Score("mount", "booth"); x != 2010 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Score("mount", "motif"); x != 22100 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Score("mount", "mount"); x != 22222 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Score("boobs", "bobos"); x != 22112 {
    panic(fmt.Sprintf("%d", x))
  }
}

func main() {
  TestScore()
  fmt.Printf("Starting with a dictionary of %d words\n", len(allWords))
}
