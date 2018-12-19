# Expandable Collection User Defined Function

### Usage

```go

    aMap := data.NewMap();
    aMap.Put("ts", "2015-02-11")
    udf.Register(aMap)
    expanded := aMap.ExpandAsText(`$FormatTime($ts, "yyyy")`)
 
```

_The list of defined UDFs:_

-  Length, Len returns length of slice, map or string
-  AsMap - convert source into a map, it accepts data structure, or JSON, YAML literal
-  AsCollection - convert source into a slice, it accepts data structure, or JSON, YAML literal
-  AsData - convert source into a map or slice, it accepts data structure, or JSON, YAML literal
-  AsInt - convert source into a an int
-  AsFloat - convert source into a a float
-  AsBool  - convert source into a boolean
-  FormatTime, takes two arguments, date or now, followed by java style date format
-  Values - returns map values
-  Keys  - return map keys
-  IndexOf - returns index of matched slice element
-  Join - join slice element with supplied separator