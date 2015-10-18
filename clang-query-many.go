package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

//////////////////////////////
// Set from http://arslan.io/thread-safe-set-data-structure-in-go

type Set struct {
	m map[string]bool
	sync.RWMutex
}

func New() *Set {
	return &Set{
		m: make(map[string]bool),
	}
}

// Add add
func (s *Set) Add(item string) {
	s.Lock()
	defer s.Unlock()
	s.m[item] = true
}

// Remove deletes the specified item from the map
func (s *Set) Remove(item string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}

// Has looks for the existence of an item
func (s *Set) Has(item string) bool {
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
	s.m = make(map[string]bool)
}

// IsEmpty checks for emptiness
func (s *Set) IsEmpty() bool {
	if s.Len() == 0 {
		return true
	}
	return false
}

// Set returns a slice of all items
func (s *Set) List() []string {
	s.RLock()
	defer s.RUnlock()
	list := make([]string, 0)
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

//////////////////////////////

func clangQuery(source string, query string, args []string) string {
	proc := exec.Command("clang-query", append([]string{source}, args...)...)
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

func parseMatches(matches string) []string {
	var files []string
	lines := strings.Split(matches, "\n")
	if len(lines) <= 6 {
		return []string{}
	}
	for _, line := range lines {
		if strings.Contains(line, ": note: \"root\" binds here") {
			files = append(files, strings.TrimSuffix(line, ": note: \"root\" binds here"))
		}
	}
	return files
}

func getFlags() ([]string, []string) {
	// parse args
	flag.Parse()
	args := flag.Args()
	files := []string{}
	clangArgs := []string{}
	for k, v := range args {
		if v == "--" {
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
	matches := New()

	// limited concurrency, http://jmoiron.net/blog/limiting-concurrency-in-go/
	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// spawn workers
	for _, file := range files {
		file, _ = strconv.Unquote(file)
		sem <- true
		go func(this_file string) {
			defer func() { <-sem }()
			fmt.Println(this_file)
			for _, match := range parseMatches(clangQuery(this_file, query, clangArgs)) {
				// fmt.Println(match)
				matches.Add(match)
			}
		}(file)
	}
	// join workers
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	// print results
	for k, _ := range matches.m {
		fmt.Println(k)
	}
}
