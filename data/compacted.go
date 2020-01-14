package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/viant/toolbox"
	"reflect"
	"sync"
	"sync/atomic"
)

type Field struct {
	Name  string
	Type  reflect.Type
	index int
}

type nilGroup int

//CompactedSlice represented a compacted slice to represent object collection
type CompactedSlice struct {
	omitEmpty    bool
	compressNils bool
	lock         *sync.RWMutex
	fieldNames   map[string]*Field
	fields       []*Field
	data         [][]interface{}
	size         int64
	RawEncoding bool
}


func (d CompactedSlice) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.Write([]byte("["))
	if err != nil {
		return nil, err
	}
	i := 0
	if err = d.Range(func(item interface{}) (b bool, err error) {
		if i > 0 {
			_, err := buf.Write([]byte(","))
			if err != nil {
				return false, err
			}
		}
		i++
		data, err :=json.Marshal(item)
		if err != nil {
			return false, err
		}
		_, err = buf.Write(data)
		return err == nil, err
	});err != nil {
		return nil, err
	}
	if _, err := buf.Write([]byte("]")); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}



func (s *CompactedSlice) Fields() []*Field {
	return s.fields
}

//Size returns size of collection
func (s *CompactedSlice) Size() int {
	return int(atomic.LoadInt64(&s.size))
}

func (s *CompactedSlice) index(fieldName string, value interface{}) int {
	s.lock.RLock()
	f, ok := s.fieldNames[fieldName]
	s.lock.RUnlock()
	if ok {
		return f.index
	}
	f = &Field{Name: fieldName, index: len(s.fieldNames), Type: reflect.TypeOf(value)}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.fieldNames[fieldName] = f
	s.fields = append(s.fields, f)
	return f.index
}

func expandIfNeeded(size int, data []interface{}) []interface{} {
	if size >= len(data) {
		for i := len(data); i < size; i++ {
			data = append(data, nil)
		}
	}
	return data
}

func (s *CompactedSlice) compress(data []interface{}) []interface{} {
	var compressed = make([]interface{}, 0)
	var nilCount = 0
	for _, item := range data {
		if item != nil {
			switch nilCount {
			case 0:
			case 1:
				compressed = append(compressed, nil)
			default:
				compressed = append(compressed, nilGroup(nilCount))
			}
			compressed = append(compressed, item)
			nilCount = 0
			continue
		}
		nilCount++
	}
	return compressed
}

func (s *CompactedSlice) uncompress(in, out []interface{}) {
	var index = 0
	for i := 0; i < len(in); i++ {
		var item = in[i]
		nilGroup, ok := item.(nilGroup)
		if !ok {
			out[index] = item
			index++
			continue
		}
		for j := 0; j < int(nilGroup); j++ {
			out[index] = nil
			index++
		}
	}
	for i := index; i < len(out); i++ {
		out[i] = nil
	}
}

//Add adds data to a collection
func (s *CompactedSlice) Add(data map[string]interface{}) {
	var initSize = len(s.fieldNames)
	if initSize < len(data) {
		initSize = len(data)
	}
	atomic.AddInt64(&s.size, 1)
	var record = make([]interface{}, initSize)
	for k, v := range data {
		i := s.index(k, v)
		if !(i < len(record)) {
			record = expandIfNeeded(i+1, record)
		}
		if s.omitEmpty {
			if toolbox.IsString(v) {
				if toolbox.AsString(v) == "" {
					v = nil
				}
			} else if toolbox.IsInt(v) {
				if toolbox.AsInt(v) == 0 {
					v = nil
				}
			} else if toolbox.IsFloat(v) {
				if toolbox.AsFloat(v) == 0.0 {
					v = nil
				}
			}
		}
		record[i] = v
	}
	if s.compressNils {
		record = s.compress(record)
	}
	s.data = append(s.data, record)
}

func (s *CompactedSlice) mapNamesToFieldPositions(names []string) ([]int, error) {
	var result = make([]int, 0)
	for _, name := range names {
		field, ok := s.fieldNames[name]
		if !ok {
			return nil, fmt.Errorf("failed to lookup Field: %v", name)
		}
		result = append(result, field.index)
	}
	return result, nil
}

//SortedRange sort collection by supplied index and then call for each item supplied handler callback
func (s *CompactedSlice) SortedRange(indexBy []string, handler func(item interface{}) (bool, error)) error {
	s.lock.Lock()
	fields := s.fields
	data := s.data
	s.data = [][]interface{}{}
	s.lock.Unlock()
	indexByPositions, err := s.mapNamesToFieldPositions(indexBy)
	if err != nil {
		return err
	}

	var indexedRecords = make(map[interface{}][]interface{})
	var record = make([]interface{}, len(s.fields))
	var key interface{}
	for _, item := range data {
		atomic.AddInt64(&s.size, -1)
		if s.compressNils {
			s.uncompress(item, record)
		} else {
			record = item
		}
		key = indexValue(indexByPositions, item)
		indexedRecords[key] = item
	}

	keys, err := sortKeys(key, indexedRecords)
	if err != nil {
		return err
	}
	for _, key := range keys {
		item := indexedRecords[key]
		if s.compressNils {
			s.uncompress(item, record)
		} else {
			record = item
		}

		var aMap = map[string]interface{}{}
		recordToMap(fields, record, aMap)
		if next, err := handler(aMap); !next || err != nil {
			return err
		}

	}
	return nil
}

//SortedIterator returns sorted iterator
func (s *CompactedSlice) SortedIterator(indexBy []string) (toolbox.Iterator, error) {
	s.lock.Lock()
	fields := s.fields
	data := s.data
	s.data = [][]interface{}{}
	s.lock.Unlock()
	if len(indexBy) == 0 {
		return nil, fmt.Errorf("indexBy was empty")
	}
	indexByPositions, err := s.mapNamesToFieldPositions(indexBy)
	if err != nil {
		return nil, err
	}
	var record = make([]interface{}, len(fields))
	var indexedRecords = make(map[interface{}][]interface{})
	var key interface{}
	for _, item := range data {
		atomic.AddInt64(&s.size, -1)
		if s.compressNils {
			s.uncompress(item, record)
		} else {
			record = item
		}
		key = indexValue(indexByPositions, record)
		indexedRecords[key] = item
	}

	data = nil
	keys, err := sortKeys(key, indexedRecords)
	if err != nil {
		return nil, err
	}
	atomic.AddInt64(&s.size, int64(-len(data)))
	return &iterator{
		size: len(indexedRecords),
		provider: func(index int) (map[string]interface{}, error) {
			if index >= len(indexedRecords) {
				return nil, fmt.Errorf("index: %d out bounds:%d", index, len(data))
			}
			key := keys[index]
			item := indexedRecords[key]
			if s.compressNils {
				s.uncompress(item, record)
			} else {
				record = item
			}
			var aMap = map[string]interface{}{}
			recordToMap(fields, record, aMap)
			return aMap, nil
		},
	}, nil
}

//Range iterate over slice, and remove processed data from the compacted slice
func (s *CompactedSlice) Range(handler func(item interface{}) (bool, error)) error {
	s.lock.Lock()
	fields := s.fields
	data := s.data
	s.data = [][]interface{}{}
	s.lock.Unlock()

	var record = make([]interface{}, len(s.fields))
	for _, item := range data {
		atomic.AddInt64(&s.size, -1)
		if s.compressNils {
			s.uncompress(item, record)
		} else {
			record = item
		}
		var aMap = map[string]interface{}{}
		recordToMap(fields, record, aMap)
		if next, err := handler(aMap); !next || err != nil {
			return err
		}
	}
	return nil
}

//Ranger moves data from slice to ranger
func (s *CompactedSlice) Ranger() toolbox.Ranger {
	s.lock.Lock()
	clone := &CompactedSlice{
		data:         s.data,
		fields:       s.fields,
		size:         s.size,
		omitEmpty:    s.omitEmpty,
		compressNils: s.compressNils,
		lock:         &sync.RWMutex{},
		fieldNames:   s.fieldNames,
	}
	s.data = [][]interface{}{}
	atomic.StoreInt64(&s.size, 0)
	s.lock.Unlock()
	return clone
}

//Iterator returns a slice iterator
func (s *CompactedSlice) Iterator() toolbox.Iterator {
	s.lock.Lock()
	fields := s.fields
	data := s.data
	s.data = [][]interface{}{}
	s.lock.Unlock()
	atomic.AddInt64(&s.size, int64(-len(data)))

	var record = make([]interface{}, len(fields))
	return &iterator{
		size: len(data),
		provider: func(index int) (map[string]interface{}, error) {
			if index >= len(data) {
				return nil, fmt.Errorf("index: %d out bounds:%d", index, len(data))
			}
			item := data[index]
			if s.compressNils {
				s.uncompress(item, record)
			} else {
				record = item
			}
			var aMap = map[string]interface{}{}
			recordToMap(fields, record, aMap)
			return aMap, nil
		},
	}
}

type iterator struct {
	size     int
	provider func(index int) (map[string]interface{}, error)
	index    int
}

//HasNext returns true if iterator has next element.
func (i *iterator) HasNext() bool {
	return i.index < i.size
}

//Next sets item pointer with next element.
func (i *iterator) Next(itemPointer interface{}) error {
	record, err := i.provider(i.index)
	if err != nil {
		return err
	}
	switch pointer := itemPointer.(type) {
	case *map[string]interface{}:
		*pointer = record
	case *interface{}:
		*pointer = record
	default:
		return fmt.Errorf("unsupported type: %T, expected *map[string]interface{}", itemPointer)
	}
	i.index++
	return nil
}

//NewCompactedSlice create new compacted slice
func NewCompactedSlice(omitEmpty, compressNils bool) *CompactedSlice {
	return &CompactedSlice{
		omitEmpty:    omitEmpty,
		compressNils: compressNils,
		fields:       make([]*Field, 0),
		fieldNames:   make(map[string]*Field),
		data:         make([][]interface{}, 0),
		lock:         &sync.RWMutex{},
	}
}
