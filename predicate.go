package toolbox



//Predicate represents a generic predicate
type Predicate interface {
	Apply(value interface{}) bool
}



