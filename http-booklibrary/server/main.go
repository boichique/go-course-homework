package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudmachinery/apps/http-booklibrary/contracts"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 18080, "port to listen on")
	flag.Parse()

	// handler -> store -> db (in-memory)
	s := newBookStore()
	h := newHandler(s)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), h)
	if err != nil {
		fmt.Println(err)
	}
}

type handler struct {
	store *bookStore
	mux   *http.ServeMux
}

// Instead of using a global http.ServeMux we create a new one for the handler
// Handler registers all routes and implements http.Handler interface, so it can be used with http.ListenAndServe
func newHandler(store *bookStore) *handler {
	h := &handler{
		store: store,
		mux:   http.NewServeMux(),
	}
	h.mux.HandleFunc("/books", h.handle)

	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) {
	// both id and isbn query parameters are optional
	id := r.URL.Query().Get("id")
	isbn := r.URL.Query().Get("isbn")

	// in case both are provided -> bad request
	if id != "" && isbn != "" {
		http.Error(w, "both id and isbn provided", http.StatusBadRequest)
		return
	}

	// in case only one is provided -> return book by id or isbn
	if id != "" {
		intId, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		book := h.store.GetBookById(intId)
		h.sendBook(w, book)
		return
	}

	if isbn != "" {
		book := h.store.GetBookByISBN(isbn)
		h.sendBook(w, book)
		return
	}

	// in case none is provided -> return all books sorted by their ids
	books := h.store.GetAllBooks()
	h.sendBooks(w, books)
}

func (h *handler) handlePost(w http.ResponseWriter, r *http.Request) {
	// add new book. Use contracts.CreateBookRequest as a request body
	var request contracts.CreateBookRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.ISBN == "" {
		http.Error(w, "isbn cannot be empty", http.StatusBadRequest)
		return
	}

	// In case ISBN exists, return http.StatusConflict
	id, err := h.store.AddBook(request.ISBN, request.Title)
	if err != nil {
		if err == ErrISBNAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	book := h.store.GetBookById(id)
	h.sendBook(w, book)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// id query parameter is required
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "no id provided", http.StatusBadRequest)
		return
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if h.store.DeleteBook(intId) {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "book not found", http.StatusNotFound)
}

func (h *handler) sendBook(w http.ResponseWriter, book *contracts.Book) {
	if book == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(book)
}

func (h *handler) sendBooks(w http.ResponseWriter, books []*contracts.Book) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(books)
}
