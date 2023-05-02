package contracts

type Book struct {
	Id    int    `json:"id"`
	ISBN  string `json:"isbn"` // should be unique
	Title string `json:"title"`
}

type CreateBookRequest struct {
	ISBN  string `json:"isbn"`
	Title string `json:"title"`
}
