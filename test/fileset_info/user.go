package fileset_info

import (
	"fmt"
	"time"
)

type Addresses []Address
type Ints []int
type Z string

type Country struct {
	Code string
	Name string
}

//Address represents a test struct
type Address struct {
	Country
	City string
}

type A Address

//User represents a test struct
type User struct { //my comments
	///abc comment
	ID             *int //  comment1 type
	Name           string
	DateOfBirth    time.Time `foo="bar"`
	Address        Address
	AddressPointer *Address
	Addresses      Addresses
	Ints           []int
	Ints2          Ints
	M              map[string][]string
	C              chan *bool
	Appointments   []time.Time
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
