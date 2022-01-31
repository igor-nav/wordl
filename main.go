package main

import (
  _ "embed"
  "fmt"
  "runtime"
  "sync"
)

//go:embed 5-letter-words.txt
var data string

var allWords []string
var wordScore map[string]int64
var scoreCache [][]int

func init() {
  // Read the dictionary from the file.
  wordScore = make(map[string]int64)
  for len(data) > 0 {
    var word string
    var score int64
    k, err := fmt.Sscanf(data, "%s\t%d\n", &word, &score)
    if err != nil || k != 2 {
      panic(fmt.Sprintf("can't read '%s': %q", data[:20], err))
    }
    allWords = append(allWords, word)
    wordScore[word] = score
    data = data[7 + Length(score):]
  }
}

// Length returns the number of digits in 'x', a non-negative integer.
func Length(x int64) int {
  if x == 0 {
    return 1
  }
  ans := 0
  for x > 0 {
    x /= 10
    ans++
  }
  return ans
}

func TestLength() {
  if x := Length(0); x != 1 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Length(1); x != 1 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Length(99); x != 2 {
    panic(fmt.Sprintf("%d", x))
  }
  if x := Length(100); x != 3 {
    panic(fmt.Sprintf("%d", x))
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
  for _, guess := range words {
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
  words := make([]int, len(allWords))
  for i := 0; i < len(words); i++ {
    words[i] = i
  }
  for len(words) > 1 {
    guess := algo(words)[0]
    fmt.Printf("There are %d possible words remaining\n", len(words))
    fmt.Printf("Your next guess should be: %s\n", allWords[guess])
    fmt.Printf("What does the game say? ")
    var reply int
    if k, err := fmt.Scanf("%d\n", &reply); k != 1 || err != nil {
      panic(err)
    }
    words = SplitFast(words, guess)[reply]
  }
  fmt.Printf("And the answer is... %s!\n", allWords[words[0]])
}

// PrecomputeScores caches all possible results of Score().
func PrecomputeScores() {
  n := len(allWords)
  threads := runtime.NumCPU()
  var wg sync.WaitGroup
  scoreCache = make([][]int, n)

  for k := 0; k < threads; k++ {
    wg.Add(1)
    go func(k int) {
      for i := k; i < n; i += threads {
        scoreCache[i] = make([]int, n)
        for j := 0; j < n; j++ {
          scoreCache[i][j] = Score(allWords[i], allWords[j])
        }
      }
      wg.Done()
    }(k)
  }
  wg.Wait()
}

func main() {
  TestScore()
  TestLength()

  fmt.Printf("Starting with a dictionary of %d words\n", len(allWords))
  PrecomputeScores()
  fmt.Printf("Hold on. Computing...\n")

  words := make([]int, len(allWords))
  for i := 0; i < len(words); i++ {
    words[i] = i
  }
  for i, x := range Eval(GreedyMinimax, words) {
    if x > 0 {
      fmt.Printf("  %d tries to guess %4d words\n", i, x)
    }
  }

  Play(GreedyMinimax)
}
