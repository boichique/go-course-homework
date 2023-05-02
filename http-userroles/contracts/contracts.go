package contracts

type User struct {
	Email    string   `json:"email"` // unique key
	FullName string   `json:"full_name"`
	Roles    []string `json:"roles"`
}
