package toolbox

import (
	"reflect"
	"sync"
)

//TrueValueProvider is a function that returns true, it takes one parameters which ignores,
//This provider can be used to make map from slice like map[some type]bool

//ProcessSliceAsync iterates over any slice, it calls handler with each element asynchronously
func ProcessSliceAsync(slice interface{}, handler func(item interface{}) bool) {
	//The common cases with reflection for speed
	var wg sync.WaitGroup
	if aSlice, ok := slice.([]interface{}); ok {
		wg.Add(len(aSlice))
		for _, item := range aSlice {
			go func(item interface{}) {
				defer wg.Done()
				handler(item)
			}(item)

		}
		wg.Wait()
		return
	}
	if aSlice, ok := slice.([]map[string]interface{}); ok {
		wg.Add(len(aSlice))
		for _, item := range aSlice {
			go func(item interface{}) {
				defer wg.Done()
				handler(item)
			}(item)

		}
		wg.Wait()
		return
	}
	//The common cases with reflection for speed
	if aSlice, ok := slice.([]string); ok {
		wg.Add(len(aSlice))
		for _, item := range aSlice {
			go func(item interface{}) {
				defer wg.Done()
				handler(item)
			}(item)

		}
		wg.Wait()
		return
	}

	//The common cases with reflection for speed
	if aSlice, ok := slice.([]int); ok {
		wg.Add(len(aSlice))
		for _, item := range aSlice {
			go func(item interface{}) {
				defer wg.Done()
				handler(item)
			}(item)
		}
		wg.Wait()
		return
	}

	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	wg.Add(sliceValue.Len())
	for i := 0; i < sliceValue.Len(); i++ {
		go func(item interface{}) {
			defer wg.Done()
			handler(item)
		}(sliceValue.Index(i).Interface())

	}
	wg.Wait()
}

//IndexSlice reads passed in slice and applies function that takes a slice item as argument to return a key value.
//passed in resulting map needs to match key type return by a key function, and accept slice item type as argument.
func IndexSliceAsync(slice, resultingMap, keyFunction interface{}) {
	var lock = sync.RWMutex{}
	mapValue := DiscoverValueByKind(resultingMap, reflect.Map)
	ProcessSliceAsync(slice, func(item interface{}) bool {
		result := CallFunction(keyFunction, item)
		lock.Lock() //otherwise, fatal error: concurrent map writes
		defer lock.Unlock()
		mapValue.SetMapIndex(reflect.ValueOf(result[0]), reflect.ValueOf(item))
		return true
	})
}

//SliceToMap reads passed in slice to to apply the key and value function for each item. Result of these calls is placed in the resulting map.
func SliceToMapAsync(sourceSlice, targetMap, keyFunction, valueFunction interface{}) {
	//optimized case
	var wg sync.WaitGroup
	var lock = sync.RWMutex{}
	if stringBoolMap, ok := targetMap.(map[string]bool); ok {
		if stringSlice, ok := sourceSlice.([]string); ok {
			if valueFunction, ok := keyFunction.(func(string) bool); ok {
				if keyFunction, ok := keyFunction.(func(string) string); ok {
					wg.Add(len(stringSlice))
					for _, item := range stringSlice {
						go func(item string) {
							defer wg.Done()
							key := keyFunction(item)
							value := valueFunction(item)
							lock.Lock()
							defer lock.Unlock()
							stringBoolMap[key] = value
						}(item)
					}
					wg.Wait()
					return
				}
			}
		}
	}

	mapValue := DiscoverValueByKind(targetMap, reflect.Map)
	ProcessSliceAsync(sourceSlice, func(item interface{}) bool {
		key := CallFunction(keyFunction, item)
		value := CallFunction(valueFunction, item)
		lock.Lock()
		defer lock.Unlock()
		mapValue.SetMapIndex(reflect.ValueOf(key[0]), reflect.ValueOf(value[0]))
		return true
	})
}

func ProcessSliceWithIndexAsync(slice interface{}, handler func(index int, item interface{}) bool) {
	var wg sync.WaitGroup
	if aSlice, ok := slice.([]interface{}); ok {
		wg.Add(len(aSlice))
		for i, item := range aSlice {
			go func(i int, item interface{}) {
				defer wg.Done()
				handler(i, item)
			}(i, item)
		}
		wg.Wait()
		return
	}
	if aSlice, ok := slice.([]string); ok {
		wg.Add(len(aSlice))
		for i, item := range aSlice {
			go func(i int, item interface{}) {
				defer wg.Done()
				handler(i, item)
			}(i, item)
		}
		wg.Wait()
		return
	}
	if aSlice, ok := slice.([]int); ok {
		wg.Add(len(aSlice))
		for i, item := range aSlice {
			go func(i int, item interface{}) {
				defer wg.Done()
				handler(i, item)
			}(i, item)
		}
		wg.Wait()
		return
	}

	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	wg.Add(sliceValue.Len())
	for i := 0; i < sliceValue.Len(); i++ {
		go func(i int, item interface{}) {
			defer wg.Done()
			handler(i, item)
		}(i, sliceValue.Index(i).Interface())
	}
	wg.Wait()
}
