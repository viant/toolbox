package toolbox

//Predicate represents a generic predicate
type Predicate interface {
	//Apply checks if predicate is true
	Apply(value interface{}) bool
}
