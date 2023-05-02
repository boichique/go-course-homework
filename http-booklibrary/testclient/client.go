package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudmachinery/apps/http-booklibrary/contracts"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

func newClient(host string, port int) *client {
	return &client{
		baseURL:    fmt.Sprintf("http://%s:%d/books", host, port),
		httpClient: &http.Client{},
	}
}

func (c *client) GetBookById(id int) (*contracts.Book, error) {
	resp, err := c.doGet(fmt.Sprintf("%s?id=%d", c.baseURL, id))
	if err != nil {
		return nil, err
	}

	return readBook(resp.Body)
}

func (c *client) GetBookByISBN(isbn string) (*contracts.Book, error) {
	resp, err := c.doGet(fmt.Sprintf("%s?isbn=%s", c.baseURL, isbn))
	if err != nil {
		return nil, err
	}

	return readBook(resp.Body)
}

func (c *client) GetBooks() ([]*contracts.Book, error) {
	resp, err := c.doGet(c.baseURL)
	if err != nil {
		return nil, err
	}

	return readBooks(resp.Body)
}

func (c *client) AddBook(isbn, title string) (*contracts.Book, error) {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(&contracts.CreateBookRequest{
		ISBN:  isbn,
		Title: title,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	return readBook(resp.Body)
}

func (c *client) DeleteBook(id int) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s?id=%d", c.baseURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	return err
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// it's common to check the status code of the response like that
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *client) doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func readBook(r io.ReadCloser) (*contracts.Book, error) {
	defer r.Close()

	var book contracts.Book
	err := json.NewDecoder(r).Decode(&book)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func readBooks(r io.ReadCloser) ([]*contracts.Book, error) {
	defer r.Close()

	var books []*contracts.Book
	err := json.NewDecoder(r).Decode(&books)
	if err != nil {
		return nil, err
	}

	return books, nil
}
