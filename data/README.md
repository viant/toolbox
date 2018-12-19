# Data utilities


### Expandable Map & Collection

Expandable structure enable nested data structure to substitution source for other data structure or text.
Data substitution expression starts with $ sign, you can use path where dot or [index] allows access data sub node.

- [ExpandAsText](#ExpandAsText)
- [Expand](#Expand)


### ExpandAsText

ExpandAsText expands any expression that has satisfied dependencies, meaning only expression that path is present
can be expanded, otherwise expression is left unchanged.
In case when UDF is used, expression is expanded if UDF does not return an error.

**Usage:**

```go
    aMap := Map(map[string]interface{}{
		"key1": 1,
		"key2": map[string]interface{}{
			"subKey1":10,
			"subKey2":20,
		},
		"key3": "subKey2",
		"array": []interface{}{
			111, 222, 333,
		},
		"slice": []interface{}{
			map[string]interface{}{
				"attr1":111,
				"attr2":222,
			},
		},
	})
	expandedText := aMap.ExpandAsText(`1: $key1, 
2: ${array[2]}  
3: $key2.subKey1 
4: $key2[$key3] ${slice[0].attr1}  
5: ${(key1 + 1) * 3} 
6: $abc
7: end
`)
	
	
/* expands to 
1: 1, 
2: 333  
3: 10 
4: 20 111  
5: 6 
6: $abc
7: end
*/
```

## Expand arbitrary data structure


```go
    aMap := Map(map[string]interface{}{
		"key1": 1,
		"key2": map[string]interface{}{
			"subKey1":10,
			"subKey2":20,
		},
		"key3": "subKey2",
		"array": []interface{}{
			111, 222, 333,
		},
		"slice": []interface{}{
			map[string]interface{}{
				"attr1":111,
				"attr2":222,
			},
		},
	})
    
    data := map[string]interface{}{
    	"k1": "$key1",
    	"k2":"$array",
    	"k3":"$key2",
    }
    expanded := aMap.Expand(data)
    /* expands to
        map[string]interface{}{
        	"k1": 1,
        	"k2": []interface{}{111, 222, 333},
        	"k3": map[string]interface{}{
                "subKey1":10,
                "subKey2":20,
            },
        }
    */
    
```


# UDF expandable User defined function

You can add dynamic data substitution by registering function in top level map.

```go
type Udf func(interface{}, Map) (interface{}, error)
```

```go
        aMap: data.Map(map[string]interface{}{
            "dailyCap":   100,
            "overallCap": 2,
            "AsFloat": func(source interface{}, state Map) (interface{}, error) {
                return toolbox.AsFloat(source), nil
            },
        })

        expanded := aMap.Expand("$AsFloat($dailyCap)")
        //expands to actual float: 100.0

```

[Predefined UDF](udf)


### Compacted slice

Using a generic data structure in a form []map[string]interface{} is extremely memory inefficient, 
CompactedSlice addresses the memory inefficiency by storing the new item values in a slice 
and by mapping corresponding fields to the item slice index positions. 
On top of that any neighboring nil values can be compacted too. 


**Usage**

```go

    collection := NewCompactedSlice(true, true)

    for i := 0;i<10;i++ {
        collection.Add(map[string]interface{}{
            "f1":  i+1,
            "f12": i+10,
            "f15": i*20,
            "f20": i+4,
            "f11": nil,
            "f12": nil,
            "f13": nil,
            "f14": "",
        })
	}
    
    collection.Range(func(data interface{}) (bool, error) {
        actual = append(actual, toolbox.AsMap(data))
        return true, nil
    })
    
    
 
```