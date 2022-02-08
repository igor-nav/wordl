package main

import (
  _ "embed"
  "fmt"
  "runtime"
  "sync"
)

//go:embed words1.txt
var words1txt string

//go:embed words2.txt
var words2txt string

var words1 []string
var words2 []string
var scoreCache [][]int

func init() {
  words1 = ReadFile(words1txt)
  words2 = ReadFile(words2txt)
}

func ReadFile(data string) []string {
  allWords := make([]string, 0, 11000)
  for len(data) > 0 {
    var word string
    k, err := fmt.Sscanf(data, "%s\n", &word)
    if err != nil || k != 1 {
      panic(fmt.Sprintf("can't read '%s': %q", data[:20], err))
    }
    allWords = append(allWords, word)
    data = data[6:]
  }
  return allWords
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

// SplitFast is Split() that works on word indices rather than strings.
func SplitFast(targets []int, guess int) map[int][]int {
  ans := make(map[int][]int)
  for _, target := range targets {
    s := scoreCache[target][guess]
    ans[s] = append(ans[s], target)
  }
  return ans
}

// GreedyMinimax finds the guess that minimizes the largest remaining word set.
// It returns the list of best guesses (most likely just one).
func GreedyMinimax(words []int) []int {
  minimax := len(words)
  bestGuesses := make([]int, 0)
  //for _, guess := range words {
  for guess, _ := range words1 {  // DEBUG
    s := SplitFast(words, guess)
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

// GreedyLookahead2 tries all pairs of guesses and picks the best.
func GreedyLookahead2(words []int) []int {
  threads := runtime.NumCPU()
  var wg sync.WaitGroup
  minimaxes := make([]int, len(words1) + len(words2))

  for k := 0; k < threads; k++ {
    wg.Add(1)
    go func(k int) {
      for guess := k; guess < len(words1) + len(words2); guess++ {
        minimaxes[guess] = len(words) * 101
        s := SplitFast(words, guess)
        var maxChunk []int
        for _, v := range s {
          if len(v) > len(maxChunk) {
            maxChunk = v
          }
        }
        for guess2, _ := range words1 {
          s2 := SplitFast(maxChunk, guess2)
          maxSize := 0
          for _, v2 := range s2 {
            maxSize = len(v2) * 100 + len(maxChunk)
          }
          if maxSize < minimaxes[guess] {
            minimaxes[guess] = maxSize
          }
        }
      }
      wg.Done()
    }(k)
  }
  wg.Wait()

  minimax := len(words) * 101
  bestGuesses := make([]int, 0)
  for guess := 0; guess < len(words1) + len(words2); guess++ {
    if minimaxes[guess] < minimax {
      minimax = minimaxes[guess]
      bestGuesses = bestGuesses[:0]
    }
    if minimaxes[guess] == minimax {
      bestGuesses = append(bestGuesses, guess)
    }
  }
  return bestGuesses
}

// Eval evaluates an algorithm on a set of words.
// Returns a histogram of how many guesses it takes to win, by target word.
func Eval(algo func([]int)[]int, words []int) []int {
  guess := algo(words)[0]
  ans := make([]int, 2)
  for reply, subWords := range SplitFast(words, guess) {
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

// Play plays the game interactively using the given algorithm.
func Play(algo func([]int)[]int) {
  fmt.Printf("Lets play...\n")
  words := make([]int, len(words1))
  for i := 0; i < len(words); i++ {
    words[i] = i
  }
  for len(words) > 1 {
    guess := algo(words)[0]
    fmt.Printf("There are %d possible words remaining", len(words))
    if len(words) < 8 {
      fmt.Printf(":")
      for _, w := range words {
        fmt.Printf(" %s", words1[w])
      }
    }
    fmt.Printf("\n")
    if guess < len(words1) {
      fmt.Printf("Your next guess should be: %s\n", words1[guess])
    } else {
      fmt.Printf("Your next guess should be: %s\n", words2[guess - len(words1)])
    }
    fmt.Printf("What does the game say? ")
    var reply int
    if k, err := fmt.Scanf("%d\n", &reply); k != 1 || err != nil {
      panic(err)
    }
    words = SplitFast(words, guess)[reply]
  }
  fmt.Printf("And the answer is... %s!\n", words1[words[0]])
}

// PrecomputeScores caches all possible results of Score().
func PrecomputeScores() {
  n := len(words1)
  threads := runtime.NumCPU()
  var wg sync.WaitGroup
  scoreCache = make([][]int, n)

  for k := 0; k < threads; k++ {
    wg.Add(1)
    go func(k int) {
      for i := k; i < n; i += threads {
        scoreCache[i] = make([]int, n + len(words2))
        for j := 0; j < n; j++ {
          scoreCache[i][j] = Score(words1[i], words1[j])
        }
        for j, w := range words2 {
          scoreCache[i][n + j] = Score(words1[i], w)
        }
      }
      wg.Done()
    }(k)
  }
  wg.Wait()
}

func main() {
  TestScore()

  fmt.Printf("Starting with a dictionary of %d words\n", len(words1))
  PrecomputeScores()
  fmt.Printf("Hold on. Computing...\n")
/*
  words := make([]int, len(words1))
  for i := 0; i < len(words); i++ {
    words[i] = i
  }
  for i, x := range Eval(GreedyMinimax, words) {
    if x > 0 {
      fmt.Printf("  %d tries to guess %4d words\n", i, x)
    }
  }
*/
  //Play(GreedyMinimax)
  Play(GreedyLookahead2)
}
