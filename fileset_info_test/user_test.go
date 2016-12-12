package fileset_info

import (
	"fmt"
	"time"
)

//Address represents a test struct
type Address struct {
	City string
}

//User represents a test struct
type User struct { //my comments
	///abc comment
	ID          *int //  comment1 type
	Name        string
	DateOfBirth time.Time `foo="bar"`
	Address     Address
	Ints        []int
	M           map[string][]string
	C           chan *bool
}

//Test represents a test method
func (u User) Test() {
	fmt.Printf("Abc %v", u)
}

//Test1 represents a test method
func (u User) Test1() bool {
	fmt.Printf("Abc %v", u)
	return false
}
