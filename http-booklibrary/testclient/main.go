package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/cloudmachinery/apps/http-booklibrary/contracts"
	"github.com/google/uuid"
)

func main() {
	var (
		mode string
		host string
		port int
	)
	flag.StringVar(&mode, "mode", "correctness", "test mode (correctness, concurrency)")
	flag.StringVar(&host, "host", "localhost", "server host")
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	c := newClient(host, port)

	switch mode {
	case "correctness":
		correctness(c)
	case "concurrency":
		concurrency(c)
	default:
		panic("unknown mode")
	}
}

func correctness(c *client) {
	fmt.Println("Make sure the server has clean state before running the test.")

	// Add books
	guide1, err := c.AddBook("978-3-16-148410-0", "The Hitchhiker's Guide to the Galaxy")
	requireNoError(err)
	requireEquals(guide1.Id, 1)

	guide2, err := c.AddBook("978-3-16-148410-1", "The Hitchhiker's Guide to the Galaxy 2")
	requireNoError(err)
	requireEquals(guide2.Id, 2)

	_, err = c.AddBook(guide1.ISBN, guide1.Title)
	requireError(err)

	// Get books
	books, err := c.GetBooks()
	requireNoError(err)
	requireEquals(books, []*contracts.Book{guide1, guide2})

	// Get book by ISBN
	b, err := c.GetBookByISBN(guide2.ISBN)
	requireNoError(err)
	requireEquals(b, guide2)

	// Get book by ID
	b, err = c.GetBookById(guide1.Id)
	requireNoError(err)
	requireEquals(b, guide1)

	// Delete book
	err = c.DeleteBook(guide1.Id)
	requireNoError(err)

	// There should be only one book left
	books, err = c.GetBooks()
	requireNoError(err)
	requireEquals(books, []*contracts.Book{guide2})

	_, err = c.GetBookById(guide1.Id)
	requireError(err)

	_, err = c.GetBookByISBN(guide1.ISBN)
	requireError(err)

	err = c.DeleteBook(guide1.Id)
	requireNoError(err)

	fmt.Println("Correctness test passed")
}

func concurrency(c *client) {
	var (
		duration = time.Second * 10
		readers  = 10
		mutators = 10
	)

	fmt.Printf("Running concurrency test for %v with %d readers and %d mutators\n", duration, readers, mutators)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ids := newSafeSet[int]()
	isbns := newSafeSet[string]()

	var wg sync.WaitGroup
	wg.Add(readers + mutators)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()

			for ctx.Err() == nil {
				switch rand.Intn(3) {
				case 0:
					_, _ = c.GetBooks()
				case 1:
					_, _ = c.GetBookById(ids.FirstOrDefault(rand.Intn(100)))
				case 2:
					_, _ = c.GetBookByISBN(isbns.FirstOrDefault("978-3-16-148410-0"))
				}
			}
		}()
	}

	for i := 0; i < mutators; i++ {
		go func() {
			defer wg.Done()

			for ctx.Err() == nil {
				switch rand.Intn(3) {
				// Add book
				case 0:
					isbn := uuid.New().String()
					_, err := c.AddBook(isbn, fmt.Sprintf("Book %s", isbn))
					if err != nil {
						fmt.Printf("cannot add book: %v\n", err)
					} else {
						isbns.Add(isbn)
					}
				// Delete existing book
				case 1:
					id, ok := ids.TryGetFirst()
					if !ok {
						continue
					}

					if id != 0 {
						err := c.DeleteBook(id)
						if err != nil {
							fmt.Printf("cannot delete book: %v\n", err)
						} else {
							ids.Remove(id)
						}
					}
				// Delete non-existing book
				case 2:
					id := rand.Intn(1000000)
					_ = c.DeleteBook(id)
					ids.Remove(id)
				}
			}
		}()
	}

	wg.Wait()

	fmt.Println("Concurrency test passed")
}

func requireNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func requireError(err error) {
	if err == nil {
		panic("expected error")
	}
}

func requireEquals(a, b any) {
	if !reflect.DeepEqual(a, b) {
		panic(fmt.Sprintf("expected %v, got %v", a, b))
	}
}

type safeSet[T comparable] struct {
	mx  sync.RWMutex
	set map[T]struct{}
}

func newSafeSet[T comparable]() *safeSet[T] {
	return &safeSet[T]{set: make(map[T]struct{})}
}

func (s *safeSet[T]) Add(v T) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.set[v] = struct{}{}
}

func (s *safeSet[T]) Remove(v T) {
	s.mx.Lock()
	defer s.mx.Unlock()

	delete(s.set, v)
}

func (s *safeSet[T]) FirstOrDefault(def T) T {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for v := range s.set {
		return v
	}

	return def
}

func (s *safeSet[T]) TryGetFirst() (T, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for v := range s.set {
		return v, true
	}

	var def T
	return def, false
}
