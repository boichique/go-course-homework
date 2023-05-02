package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/cloudmachinery/apps/http-booklibrary/contracts"
)

var ErrISBNAlreadyExists = fmt.Errorf("isbn already exists")

type bookStore struct {
	mx       sync.RWMutex
	lastId   int
	idToBook map[int]*contracts.Book
	isbnToId map[string]*contracts.Book
}

// newBookStore creates a new bookStore ready to use
func newBookStore() *bookStore {
	return &bookStore{
		mx:       sync.RWMutex{},
		lastId:   0,
		idToBook: make(map[int]*contracts.Book),
		isbnToId: make(map[string]*contracts.Book),
	}
}

// AddBook adds a book to the store and returns the id of the book or ErrISBNAlreadyExists if the ISBN already exists
func (s *bookStore) AddBook(isbn string, title string) (id int, err error) {
	if book := s.GetBookByISBN(isbn); book != nil {
		return 0, ErrISBNAlreadyExists
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	s.lastId++
	id = s.lastId
	book := contracts.Book{
		Id:    id,
		ISBN:  isbn,
		Title: title,
	}
	s.isbnToId[isbn] = &book
	s.idToBook[id] = &book

	return id, nil
}

// GetBookById returns a book by id or nil if the book does not exist
func (s *bookStore) GetBookById(id int) *contracts.Book {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if book, ok := s.idToBook[id]; ok {
		return book
	}
	return nil
}

// GetBookByISBN returns a book by isbn or nil if the book does not exist
func (s *bookStore) GetBookByISBN(isbn string) *contracts.Book {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if book, ok := s.isbnToId[isbn]; ok {
		return book
	}
	return nil
}

// GetAllBooks returns all books sorted by id
func (s *bookStore) GetAllBooks() []*contracts.Book {
	s.mx.RLock()

	books := make([]*contracts.Book, 0, len(s.idToBook))
	for _, book := range s.idToBook {
		books = append(books, book)
	}

	s.mx.RUnlock()

	sort.Slice(books, func(i, j int) bool {
		return books[i].Id < books[j].Id
	})

	return books
}

// DeleteBook deletes a book by id and returns true if the book was deleted
func (s *bookStore) DeleteBook(id int) bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	if book, ok := s.idToBook[id]; ok {
		isbn := book.ISBN
		delete(s.idToBook, id)
		delete(s.isbnToId, isbn)
		return true
	}
	return false
}
