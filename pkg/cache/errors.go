/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cache

import "fmt"

type CacheError struct {
	Problem      ProblemType
	WrappedError error
}

type ProblemType string

const NotJsonifiable ProblemType = ProblemType("not jsonifiable")
const ExceedsTotalCacheSize = ProblemType("exceeds total cache size")
const ExceedsCacheSize = ProblemType("exceeds cache size")
const ObjectToLarge = ProblemType("object to large")

const NoItem = ProblemType("no item")

func (t *CacheError) Error() string {
	var wrapped string
	if t.WrappedError != nil {
		wrapped = t.WrappedError.Error()
	} else {
		wrapped = "no wrapped error"
	}
	return fmt.Sprintf("%s wrapped - %s ", t.Problem, wrapped)
}

func NewCacheError(problem ProblemType, wrapped error) *CacheError {
	ret := new(CacheError)
	ret.Problem = problem
	ret.WrappedError = wrapped
	return ret
}
