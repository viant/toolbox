package data

import (
	"sync"
	"github.com/viant/toolbox"
	"sync/atomic"
)

type field struct {
	Name  string
	Index int
}

type nilGroup int

//CompactedSlice represented a compacted slice to represent object collection
type CompactedSlice struct {
	omitEmpty    bool
	compressNils bool
	lock         *sync.RWMutex
	fieldNames   map[string]*field
	fields       []*field
	data         [][]interface{}
	size         int64
}

func (s *CompactedSlice) Size() int {
	return int(atomic.LoadInt64(&s.size))
}

func (s *CompactedSlice) index(fieldName string) int {
	s.lock.RLock()
	f, ok := s.fieldNames[fieldName]
	s.lock.RUnlock()
	if ok {
		return f.Index
	}
	f = &field{Name: fieldName, Index: len(s.fieldNames)}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.fieldNames[fieldName] = f
	s.fields = append(s.fields, f)
	return f.Index
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
		if ! ok {
			out[index] = item
			index++
			continue
		}
		for j := 0; j < int(nilGroup); j++ {
			out[index] = nil
			index++
		}
	}
	for i := index; i<len(out);i++ {
		out[i] = nil
	}
}

func (s *CompactedSlice) Add(data map[string]interface{}) {
	var initSize = len(s.fieldNames)
	if initSize < len(data) {
		initSize = len(data)
	}
	atomic.AddInt64(&s.size, 1)
	var record = make([]interface{}, initSize)
	for k, v := range data {
		i := s.index(k)
		if ! (i < len(record)) {
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
		for _, field := range fields {
			index := field.Index
			var value = record[index]
			if value == nil {
				continue
			}
			aMap[field.Name] = value
		}
		if next, err := handler(aMap); ! next || err != nil {
			return err
		}
	}
	return nil
}

//NewCompactedSlice create new compacted slice
func NewCompactedSlice(omitEmpty, compressNils bool) *CompactedSlice {
	return &CompactedSlice{
		omitEmpty:    omitEmpty,
		compressNils: compressNils,
		fields:       make([]*field, 0),
		fieldNames:   make(map[string]*field),
		data:         make([][]interface{}, 0),
		lock:         &sync.RWMutex{},
	}
}
