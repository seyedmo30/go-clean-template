package integration

// jsonplaceholder user structure
type jpAddressGeo struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}
type jpAddress struct {
	Street  string       `json:"street"`
	Suite   string       `json:"suite"`
	City    string       `json:"city"`
	Zipcode string       `json:"zipcode"`
	Geo     jpAddressGeo `json:"geo"`
}
type jpCompany struct {
	Name        string `json:"name"`
	CatchPhrase string `json:"catchPhrase"`
	Bs          string `json:"bs"`
}
type jpUser struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Address  jpAddress `json:"address"`
	Phone    string    `json:"phone"`
	Website  string    `json:"website"`
	Company  jpCompany `json:"company"`
}
