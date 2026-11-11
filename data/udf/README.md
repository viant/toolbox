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

#### Calendar date offsets

`FormatTime(value, format, timezone, truncateFormat, offset)` accepts an optional
fifth argument object containing integer `years`, `months`, and `days` fields.
Omitted fields are zero; unknown fields, fractional values, and values outside
the platform's `int` range fail. JSON numeric values must also be integral.
The existing two-, three-, and four-argument forms retain their behavior.
In the five-argument form, use `""` as the fourth argument for no truncation.

```go
state := data.NewMap()
udf.Register(state)
state.Put("offset", data.Map{"days": -30})
// 2026-10-01
result := state.ExpandAsText(`$FormatTime("2026-10-31", "yyyy-MM-dd", "UTC", "", $offset)`)
```

The existing parser resolves the input and converts it to the supplied timezone
before applying Go's `time.AddDate`. Calendar days retain local wall-clock time
across daylight-saving transitions rather than adding fixed 24-hour durations.
An explicit timestamp remains an instant: timezone conversion can change its
calendar date before the offset is applied. Existing truncation, `weekday`, and
`secondsOfDay` handling then operate on the shifted time.

Month and year offsets use Go's normalization rather than clamping: subtracting
one month from October 31 produces October 1 because September 31 normalizes into
October. For a previous-month start, explicitly anchor November 1 and then apply
`{"months": -1}` to obtain October 1. February 29 plus one year similarly
normalizes to March 1 in a non-leap year.

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
