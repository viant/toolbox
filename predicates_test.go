package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
	"time"
)

func TestWithinPredicate(t *testing.T) {
	targetTime := time.Unix(1465009041, 0)
	predicate := toolbox.NewWithinPredicate(targetTime, 2, "")
	timeValue := time.Unix(1465009042, 0)
	assert.True(t, predicate.Apply(timeValue))

	timeValue = time.Unix(1465009044, 0)
	assert.False(t, predicate.Apply(timeValue))
}

func TestBetweenPredicate(t *testing.T) {

	predicate := toolbox.NewBetweenPredicate(10, 20)

	assert.False(t, predicate.Apply(9))
	assert.True(t, predicate.Apply(10))
	assert.True(t, predicate.Apply(11))
	assert.False(t, predicate.Apply(21))

}

func TestInPredicate(t *testing.T) {

	{
		predicate := toolbox.NewInPredicate("10", 20, "a")
		assert.False(t, predicate.Apply(9))
		assert.True(t, predicate.Apply(10))
		assert.False(t, predicate.Apply(15))
		assert.True(t, predicate.Apply("a"))
		assert.True(t, predicate.Apply(20))
		assert.False(t, predicate.Apply(21))
	}
}

func TestComparablePredicate(t *testing.T) {
	{
		predicate := toolbox.NewComparablePredicate(">", "1")
		assert.True(t, predicate.Apply(3))
		assert.False(t, predicate.Apply(1))
	}
	{
		predicate := toolbox.NewComparablePredicate("<", "1")
		assert.True(t, predicate.Apply(0))
		assert.False(t, predicate.Apply(3))
	}
	{
		predicate := toolbox.NewComparablePredicate("!=", "1")
		assert.True(t, predicate.Apply(0))
		assert.False(t, predicate.Apply(1))
	}

}

func TestNewLikePredicate(t *testing.T) {
	{
		predicate := toolbox.NewLikePredicate("abc%efg")
		assert.False(t, predicate.Apply("abefg"))
		assert.True(t, predicate.Apply("abcefg"))

	}
	{
		predicate := toolbox.NewLikePredicate("abc%")
		assert.True(t, predicate.Apply("abcfg"))

	}
}

func TestNewComparablePredicate(t *testing.T) {

	{
		predicate := toolbox.NewComparablePredicate("=", "abc")
		assert.True(t, predicate.Apply("abc"))
		assert.False(t, predicate.Apply("abdc"))
	}
	{
		predicate := toolbox.NewComparablePredicate("!=", "abc")
		assert.True(t, predicate.Apply("abcc"))
		assert.False(t, predicate.Apply("abc"))
	}

	{
		predicate := toolbox.NewComparablePredicate(">=", 3)
		assert.True(t, predicate.Apply(10))
		assert.False(t, predicate.Apply(1))
	}

	{
		predicate := toolbox.NewComparablePredicate("<=", 3)
		assert.True(t, predicate.Apply(1))
		assert.False(t, predicate.Apply(10))
	}

}

func TestNewInPredicate(t *testing.T) {
	predicate := toolbox.NewInPredicate(1.2, 1.5)
	assert.True(t, predicate.Apply("1.2"))
	assert.False(t, predicate.Apply("1.1"))
}
