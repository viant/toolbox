/*
*
*
* Copyright 2012-2016 Viant.
*
* Licensed under the Apache License, Version 2.0 (the "License"); you may not
*  use this file except in compliance with the License. You may obtain a copy of
*  the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
* WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
*  License for the specific language governing permissions and limitations under
*  the License.
*
 */
package toolbox

import "reflect"

//Iterator represents generic iterator.
type Iterator interface {

	//HasNext returns true if iterator has next element.
	HasNext() bool

	//Next sets item pointer with next element.
	Next(itemPointer interface{})
}

type sliceIterator struct {
	sliceValue reflect.Value
	index      int
}

func (i *sliceIterator) HasNext() bool {
	return i.index < i.sliceValue.Len()
}

func (i *sliceIterator) Next(itemPointer interface{}) {
	AssertKind(itemPointer, reflect.Ptr, "itemPointer")
	value := i.sliceValue.Index(i.index)
	i.index++
	itemPointerValue := reflect.ValueOf(itemPointer)
	itemPointerValue.Elem().Set(value)
}

//NewSliceIterator creates a new slice iterator.
func NewSliceIterator(slice interface{}) Iterator {
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	return &sliceIterator{sliceValue: sliceValue}
}
