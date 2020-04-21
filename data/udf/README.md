# Expandable Collection User Defined Function

### Usage


#### Data substitution

```go

    aMap := data.NewMap();
    aMap.Put("ts", "2015-02-11")
    udf.Register(aMap)
    expanded := aMap.ExpandAsText(`$FormatTime($ts, "yyyy")`)
 
```
#### Data node selection


```go

	holder := data.NewMap()
    collection := data.NewCollection()
    collection.Push(map[string]interface{}{
        "amount": 2,
        "id":2,
        "name":"p1",
        "vendor":"v1",
    })
    collection.Push(map[string]interface{}{
        "amount": 12,
        "id":3,
        "name":"p2",
        "vendor":"v2",
    })
    holder.SetValue("node1.obj", collection)

	records, err := Select([]interface{}{"node1/obj/*", "id", "name:product"}, holder)


```

#### The list of defined UDFs

-  Length, Len returns length of slice, map or string
-  AsMap - convert source into a map, it accepts data structure, or JSON, YAML literal
-  AsCollection - convert source into a slice, it accepts data structure, or JSON, YAML literal
-  AsData - convert source into a map or slice, it accepts data structure, or JSON, YAML literal
-  AsInt - convert source into a an int
-  AsFloat - convert source into a a float
-  AsBool  - convert source into a boolean
-  AsNumber - converts to either int or float
-  FormatTime, takes two arguments, date or now, followed by java style date format
-  Values - returns map values
-  Keys  - return map keys
-  IndexOf - returns index of matched slice element
-  Join - join slice element with supplied separator
-  Split - split text by separator
-  Sum - sums values for matched Path, i.e. $Sum('node1/obj/*/amount')
-  Count - counts values for matched Path, i.e. $Sum('node1/obj/*/amount')
-  Select - selects attribute for matched path, i.e $Select("node1/obj/*", "id", "name:product")
-  QueryEscape - url escape
-  QueryUnescape - url unescape
-  Base64Encode
-  Base64DecodeText
-  TrimSpace
-  Elapsed elapsed time  
-  Rand
-  Replace
-  ToLower
-  ToUpper
-  AsNewLineDelimitedJSON