package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// http://arslan.io/thread-safe-set-data-structure-in-go

type Set struct {
	m map[Match]bool
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m: make(map[Match]bool),
	}
}

// Add add
func (s *Set) Add(item Match) {
	s.Lock()
	defer s.Unlock()
	s.m[item] = true
}

// Remove deletes the specified item from the map
func (s *Set) Remove(item Match) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}

// Has looks for the existence of an item
func (s *Set) Has(item Match) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}

// Len returns the number of items in a set.
func (s *Set) Len() int {
	return len(s.List())
}

// Clear removes all items from the set
func (s *Set) Clear() {
	s.Lock()
	defer s.Unlock()
	s.m = make(map[Match]bool)
}

// IsEmpty checks for emptiness
func (s *Set) IsEmpty() bool {
	if s.Len() == 0 {
		return true
	}
	return false
}

// Set returns a slice of all items
func (s *Set) List() []Match {
	s.RLock()
	defer s.RUnlock()
	list := make([]Match, 0)
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

//////////////////////////////

func clangQuery(source string, query string, args []string) string {
	allArgs := append([]string{source}, args...)
	proc := exec.Command("clang-query", allArgs...)
	pin, _ := proc.StdinPipe()
	pout, _ := proc.StdoutPipe()
	// perr, _ := proc.StderrPipe()
	proc.Start()
	pin.Write([]byte(query + "\n"))
	pin.Close()
	out, _ := ioutil.ReadAll(pout)
	proc.Wait()
	return string(out)
}

func getExtraArgs(args []string) [][]string {
	for k, v := range args {
		if v == "--" {
			return [][]string{args[k+1:]}
		}
	}
	return [][]string{}
}

type Match struct {
	info string
}

func (m Match) String() string {
	return m.info
}

func ParseMatches(matches string) []Match {
	var results []Match
	lines := strings.Split(matches, "\n")
	if len(lines) < 6 {
		return []Match{}
	}

	var match []string
	activeMatch := false
	for _, line := range lines {
		if len(line) > 5 && line[0:5] == "Match" {
			activeMatch = true
			if len(match) != 0 { // the match on the stack is done
				results = append(results, Match{strings.Join(match, "\n")})
				match = []string{}
			}
		} else if activeMatch {
			match = append(match, line)
		}
	}
	if len(match) != 0 {
		// Store the currently processing match.
		// The last line contains the number of matches which we do not care about.
		results = append(results, Match{strings.Join(match[:int(math.Max(0., float64(len(match)-2)))], "\n")})
		match = []string{}
	}

	return results
}

func getFlags() ([]string, []string) {
	// parse args
	flag.Parse()
	args := flag.Args()
	files := []string{}
	clangArgs := []string{}
	for k, v := range args {
		if v == "---" {
			files = args[:k]
			clangArgs = args[k+1:]
			return files, clangArgs
		}
		files = args
	}
	return files, clangArgs
}

func main() {
	files, clangArgs := getFlags()
	query, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	// result set
	matches := NewSet()

	// limited concurrency, http://jmoiron.net/blog/limiting-concurrency-in-go/
	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// spawn workers
	for _, file := range files {
		sem <- true
		go func(this_file string) {
			defer func() { <-sem }()
			for _, match := range ParseMatches(clangQuery(this_file, query, clangArgs)) {
				matches.Add(match)
			}
		}(file)
	}
	// join workers
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	// print results
	for i, v := range matches.List() {
		fmt.Printf("Match #%d:\n", i+1)
		fmt.Println(v)
	}
	fmt.Printf("\n%d matches\n", matches.Len())
}
