package toolbox

//Ranger represents an abstraction that has ability range over collection item
type Ranger interface {
	Range(func(item interface{}) (bool, error)) error
}
