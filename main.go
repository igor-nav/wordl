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
  // Read the dictionary from the file.
  allWords = strings.Split(data, "\n")
  if allWords[len(allWords) - 1] == "" {
    allWords = allWords[:len(allWords) - 1]
  }
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

// Split groups the words in 'targets' by their answer to the 'guess'.
func Split(targets []string, guess string) map[int][]string {
  ans := make(map[int][]string)
  for _, target := range targets {
    s := Score(target, guess)
    ans[s] = append(ans[s], target)
  }
  return ans
}

// GreedyMinimax finds the guess that minimizes the largest remaining word set.
// It returns the list of best guesses (most likely just one).
func GreedyMinimax(words []string) []string {
  minimax := len(words)
  bestGuesses := make([]string, 0)
  for _, guess := range words {
    s := Split(words, guess)
    maxSize := 0
    for _, v := range s {
      if len(v) > maxSize {
        maxSize = len(v)
      }
    }
    if maxSize < minimax {
      minimax = maxSize
      bestGuesses = bestGuesses[:0]
    }
    if maxSize == minimax {
      bestGuesses = append(bestGuesses, guess)
    }
  }
  return bestGuesses
}

// Eval evaluates an algorithm on a set of words.
// Returns a histogram of how many guesses it takes to win, by target word.
func Eval(algo func([]string)[]string, words []string) []int {
  guess := algo(words)[0]
  ans := make([]int, 2)
  for reply, subWords := range Split(words, guess) {
    if reply == 22222 {
      ans[1]++
    } else {
      subAns := Eval(algo, subWords)
      for i, x := range subAns {
        if len(ans) <= i + 1 {
          ans = append(ans, 0)
        }
        ans[i + 1] += x
      }
    }
  }
  return ans
}

func TestEval() {
  var x []int

  x = Eval(GreedyMinimax, []string{"plane", "train"})
  if len(x) != 3 || x[0] != 0 || x[1] != 1 || x[2] != 1 {
    panic(fmt.Sprintf("%q", x))
  }

  x = Eval(GreedyMinimax, []string{"count", "mount", "zyzzy"})
  if len(x) != 3 || x[0] != 0 || x[1] != 1 || x[2] != 2 {
    panic(fmt.Sprintf("%q", x))
  }
}

// Play plays the game interactively using the given algorithm.
func Play(algo func([]string)[]string) {
  fmt.Printf("Lets play...\n")
  words := allWords
  for len(words) > 1 {
    guess := algo(words)[0]
    fmt.Printf("There are %d possible words remaining\n", len(words))
    fmt.Printf("Your next guess should be: %s\n", guess)
    fmt.Printf("What does the game say? ")
    var reply int
    if k, err := fmt.Scanf("%d\n", &reply); k != 1 || err != nil {
      panic(err)
    }
    words = Split(words, guess)[reply]
  }
  fmt.Printf("And the answer is... %s!\n", words[0])
}

func main() {
  TestScore()
  TestEval()

  fmt.Printf("Starting with a dictionary of %d words\n", len(allWords))

  fmt.Printf("Hold on. Computing...\n")
  for i, x := range Eval(GreedyMinimax, allWords) {
    if x > 0 {
      fmt.Printf("  %d tries to guess %4d words\n", i, x)
    }
  }

  Play(GreedyMinimax)
}
