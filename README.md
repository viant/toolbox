# Toolbox - go utility library

[![GoReportCard](https://goreportcard.com/badge/github.com/viant/toolbox)](https://goreportcard.com/report/github.com/viant/toolbox)
[![GoDoc](https://godoc.org/github.com/viant/toolbox?status.svg)](https://godoc.org/github.com/viant/toolbox)

This library is compatible with Go 1.8+

Please refer to [`CHANGELOG.md`](CHANGELOG.md) if you encounter breaking changes.

- [Motivation](#Motivation)
- [Collection Utilities](#Collection-Utilities)
- [Converter && Conversion Utilities](#Conversion-Utilities)
- [Struct Utilities](#Struct-Utilities)
- [Function Utilities](#Function-Utilities)
- [Time Utilities](#TimeUtilities)
- [Storage API](#storage)
- [Data substitution](data/)
- [Text Utilities](text/)
- [ServiceRouter](#ServiceRouter)
- [Decoder and Encoder](#DecoderandEncoder)
- [Logger](#Logger)
- [BatchLimiter](#BatchLimiter)
- [AST Based FileSetInfo](#ast-based-filesetinfo)
- [License](#License)
- [Credits and Acknowledgements](#Credits-and-Acknowledgements)



## Motivation

This library was developed as part of [Datastore Connectivity](https://github.com/viant/dsc/) and Testibility libraries: ([Assertly](https://github.com/viant/assertly),  [Datastore testing](https://github.com/viant/dsunit/), [End to end testing](https://github.com/viant/endly/)) 
as a way to share utilities, and other abstractions that may be useful in other projects.


<a name="Collection-Utilities"></a>

### Collection Utilities


#### Iterator

Example
```go
	slice := []string{"a", "z", "c"}
	iterator := toolbox.NewSliceIterator(slice)
    value := ""
    for iterator.HasNext() {
        iterator.Next(&value)
        ...
    }
```

#### Slice utilities

The following methods work on **any slice type.**

**ProcessSlice**

Example
```go
	var aSlice interface{}
	
	toolbox.ProcessSlice(aSlice, func(item interface{}) bool {
    		...
    		return true //to continue to next element return true
    })
	
```


**ProcessSliceWithIndex**

Example:
```go
	var aSlice interface{}
	
	toolbox.ProcessSlice(aSlice, func(index int, item interface{}) bool {
    		...
    		return true //to continue to next element return true
    })
	
```


**IndexSlice**

Example:
```go
    type Foo struct{
		id int
		name string
	}

	var aSlice = []Foo{ Foo{1, "A"}, Foo{2, "B"} }
	var indexedMap = make(map[int]Foo)
	
	toolbox.IndexSlice(aSlice, indexedMap, func(foo Foo) int {
		return foo.id
	})
	
	
```


**CopySliceElements**

Example:
```go
   source := []interface{}{
   		"abc", "def", "cyz",
   	}
   	var target = make([]string, 0)
   	toolbox.CopySliceElements(source, &target)
	
	
```


**FilterSliceElements**

Example:
```go
	source := []interface{}{
		"abc", "def", "cyz","adc",
	}
	var target = make([]string, 0)
	
	toolbox.FilterSliceElements(source, func(item string) bool {
		return strings.HasPrefix(item, "a") //this matches any elements starting with a
	}, &target)
```


**HasSliceAnyElements**

Example:
```go
    source := []interface{}{
		"abc", "def", "cyz","adc",
	}
	toolbox.HasSliceAnyElements(source, "cyz")
```


**SliceToMap**

Example:
```go
    var source = []Foo{ Foo{1, "A"}, Foo{2, "B"} }
	var target = make(map[int]string)
	toolbox.SliceToMap(source, target, func(foo Foo) int {
		return foo.id
	},
	func(foo Foo) string {
		return foo.name
	})
	

```


**TransformSlice**

Example:
```go
type Product struct{ vendor, name string }
	products := []Product{
		Product{"Vendor1", "Product1"},
		Product{"Vendor2", "Product2"},
		Product{"Vendor1", "Product3"},
		Product{"Vendor1", "Product4"},
	}
	var vendors=make([]string, 0)
	toolbox.TransformSlice(products, &vendors, func(product Product) string {
		return product.vendor
	})
```


#### Map utilities

**ProcessMap**

The following methods work on **any map type.**

Example:
```go
    var aMap interface{}
	toolbox.ProcessMap(aMap, func(key, value interface{}) bool {
    		...
    		return true //to continue to next element return true
    })

```


**CopyMapEntries**

Example:
```go
    type Foo struct{id int;name string}
	
	source := map[interface{}]interface{} {
		1: Foo{1, "A"},
		2: Foo{2, "B"},
	}
	var target = make   (map[int]Foo)

	toolbox.CopyMapEntries(source, target)
```


**MapKeysToSlice**

Example:
```go
    aMap := map[string]int {
		"abc":1,
		"efg":2,
	}
	var keys = make([]string, 0)
	toolbox.MapKeysToSlice(aMap, &keys)
```
	

**GroupSliceElements**

Example:
```go
	type Product struct{vendor,name string}
    	products := []Product{
    		Product{"Vendor1", "Product1"},
    		Product{"Vendor2", "Product2"},
    		Product{"Vendor1", "Product3"},
    		Product{"Vendor1", "Product4"},
    	}
    
    	productsByVendor := make(map[string][]Product)
    	toolbox.GroupSliceElements(products, productsByVendor, func(product Product) string {
    		return product.vendor
    	})
```


**SliceToMultimap**

```go	
	type Product struct {
    		vendor, name string
    		productId    int
    	}
    
    	products := []Product{
    		Product{"Vendor1", "Product1", 1},
    		Product{"Vendor2", "Product2", 2},
    		Product{"Vendor1", "Product3", 3},
    		Product{"Vendor1", "Product4", 4},
    	}
    
    	productsByVendor := make(map[string][]int)
    	toolbox.SliceToMultimap(products, productsByVendor, func(product Product) string {
    		return product.vendor
    	},
    	func(product Product) int {
    		return product.productId
    	})
```

<a name="Conversion-Utilities"></a>
### Converter && Conversion Utilities

Converter transforms, data between any compatible or incompatible data type including struct/basicType/map/slice/interface{}
On top of that it supports custom tag to map field to target data type (i.e map)


```go
    myStruct :=  //some struct ...
    myMap := make(map[string]interface{})
    converter := NewConverter(dateLayout, keyTag) 	
    err = converter.AssignConverted(&myMap, myStruct)
    err = converter.AssignConverted(myStruct, myMap) 
```


<a name="Struct-Utilities"></a>
### Struct Utilities


**ScanStructMethods**

Scan struct methods

```go
    service := New()
    err = toolbox.ScanStructMethods(service, 1, func(method reflect.Method) error {
		fmt.Printf("%v\n", method.Name)
		return nil
	})

```

**ProcessStruct**

Scan struct fields

```go
   service := New()
    err = toolbox.ProcessStruct(service,
        func(field reflect.StructField, value reflect.Value) error {
            fmt.Print(field.Type.Name)
            return nil
    })

```


<a name="Function-Utilities"></a>
### Function Utilities




<a name="TimeUtilities"></a>
### Time Utilities	
	
**DateFormatToLayout**

Java date format style to go date layout conversion.


```go	
    dateLaout := toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z")
    timeValue, err := time.Parse(dateLaout, "2016-02-22 12:32:01 UTC")
```

**TimeAt** 

Provide dynamic semantic for creating time object

```go
    
    tomorrow, err = TimeAt("tomorrow")//tomorrow in local timezone
    timeInUTC, err := TimeAt("2 days ago in UTC") //or 2DayAgoInUTC
    yesterdayUTC, err := TimeAt("yesterdayInUTC")//yesterday in UTC
    hoursAhead, err := TimeAt("50 hours ahead")

```

**TimeDiff**

Provide dynamic semantic for creating time object based on time dif

```go

    lastyear, _ := time.Parse(DateFormatToLayout("yyyy-MM-dd"), "2017-01-01")
    ts1, e := TimeDiff(lastyear, "50 hours earlier")
    ts2, e := TimeDiff(lastyear, "3 days before in Poland")
	
```

**DayElapsed** 
```go

    t0, _ := time.Parse(DateFormatToLayout("yyyy-MM-dd hh:mm:ss"), "2017-01-01 12:00:00")
    dayElapsedInT0, err := ElapsedDay(t0) //0.5
	
```

**ElapsedToday**
```go

    elapscedInLocalTz, err := ElapsedTodayInPct("")  
    elapscedInUTC, err := ElapsedToday("UTC")
	
```
 
**RemainingToday**
```go

    elapscedInLocalTz, err := RemainingTodayInPct("")
    elapscedInUTC, err := RemainingToday("UTC")
	
```

**AtTime**
```go
    atTime := &AtTime{
        WeekDay: "*",
        Hour:    "*",
        Minute:  "10,30",
	}
    
    //returns the nearest future time for xx:10 or xx:30  
    nextScheduleTime := atTime.Next(time.Now)
```
<a name="storage"></a>

## Storage


[Storage API](storage/README.md) provides unified way of accessing local or remote storage system.  

This API has been deprecated, please consider using [Abstract Storage](https://github.com/viant/afs)

**Example** 

```go
    import (
    	"github.com/viant/toolbox/storage"
    	_ "github.com/viant/toolbox/storage/gs"	
    )
    
    
    destinationURL := "gs://myBucket/set1/content.gz"
    destinationCredentialFile = "gs-secret.json"
    storageService, err := storage.NewServiceForURL(destinationURL, destinationCredentialFile)

```



<a name="tet"></a>
### Text utilities

**ToCaseFormat** 

```go
    formatted := toolbox.ToCaseFormat(text, toolbox.CaseLowerUnderscore, toolbox.CaseLowerCamel)

```


<a name="Tokenizer"></a>
### Tokenizer

<a name="ServiceRouter"></a>
### ServiceRouter

This ServiceRouter provides simple WebService Endpoint abstractin and RESET Client utilities.


Take as example of a ReverseService defined as follow

```go

type ReverseService interface {
        Reverse(values []int) []int 
}

type reverseService struct{}

func (r *reverseService) Reverse(values []int) []int {
	var result = make([]int, 0)
	for i := len(values) - 1; i >= 0; i-- {
		result = append(result, values[i])
	}

	return result
}

```

In order to define Endpoint for this service,  define a server, a router with the service routes;



```qo


type Server struct {
    service ReverseService
    port string
}

func (s *Server) Start() {
    
    router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/reverse/{ids}",
			Handler:    s.service.Reverse,
			Parameters: []string{"ids"}, 
		},
		toolbox.ServiceRouting{
			HTTPMethod: "POST",
			URI:        "/v1/reverse/",
			Handler:    s.service.Reverse,
			Parameters: []string{"ids"},
		})
		
        http.HandleFunc("/v1/", func(writer http.ResponseWriter, reader *http.Request) {
            err := router.Route(writer, reader)
            if err != nil {
                response.WriteHeader(http.StatusInternalServerError)
            }
        })
    
        fmt.Printf("Started test server on port %v\n", port)
        log.Fatal(http.ListenAndServe(":"+port, nil))
}

```
**ServiceRouting** parameters define handler parameters that can be extracted from URI, QueryString, or from Post Body (json payload)
In addition two special parameter names are supported: @httpRequest, @httpResponseWriter  to pass in request, and response object respectively.


The REST client utility invoking our reverse service may look as follow
 
 
```go

               var result = make([]int, 0)
               err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse/1,7,3", nil, &result)
               //...
               err := toolbox.RouteToService("post", "http://127.0.0.1:8082/v1/reverse/", []int{1, 7, 3}, &result)

``` 


By default a service router uses reflection to call the matched routes handler, it is possible to avoid reflection overhead by providing the custom handler invoker.

```go


var ReverseInvoker = func(serviceRouting *toolbox.ServiceRouting, request *http.Request, response http.ResponseWriter, uriParameters map[string]interface{}) error {
	var function = serviceRouting.Handler.(func(values []int) []int)
	idsParam := uriParameters["ids"]
	ids := idsParam.([]string)
	values := make([]int, 0)
	for _, item := range ids {
		values = append(values, toolbox.AsInt(item))
	}
	var result = function(values)
	err := toolbox.WriteServiceRoutingResponse(response, request, serviceRouting, result)
	if err != nil {
		return err
	}
	return nil
}

//...

 
        router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/reverse/{ids}",
			Handler:    s.service.Reverse,
			Parameters: []string{"ids"},
			HandlerInvoker: ReverseInvoker,
		})
//...		
		

```




<a name="DecoderandEncoder"></a>
### Decoder and Encoder 

#### Decoder

This library defines DecoderFactory interface to delegate  decoder creation, 
This library comes with standard JSON and UnMarshaler (protobuf) factory implementation.

 Example
  
```go
    factory :=toolbox.NewJsonDecoderFactory()
    ....
    
    decoder := factory.Create(reader)
    foo := &Foo{}
    err = decoder.Decode(foo)



    marshalerFactory := toolbox.NewUnMarshalerDecoderFactory()
    decoder := marshalerFactory.Create(reader)
    foo := &Foo{}
    err = decoder.Decode(foo)
```


#### Encoder

This library defines EncoderFactory interface to delegate encoder creation, 
This library comes with standard JSON and Marshaler (protobuf) factory implementation.

 Example
  
```go
        factory :=toolbox.NewJsonEncoderFactory()
        ....
        buffer := new(bytes.Buffer)
        
        
        decoder := factory.Create(buffer)
        err = decoder.Encode(foo)
    
    
    
        marshalerFactory := toolbox.NewMarshalerEncoderFactory()
        decoder := marshalerFactory.Create(buffer)
        err = decoder.Encode(foo)
```



<a name="Logger"></a>
### Logger


This library provides a file logger implementation that optimizes writes.
Log messages are queues until max queue size or flush frequency are met.
On top of that Ctrl-C also forces immediate log messages flush to disk.

File template support java style time format to manage rotation on the file name level.

```go
    logger, err := toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:           "test",
		FileTemplate:      "/tmp/test[yyyyMMdd-hhmm].log",
		QueueFlashCount:   250,
		MaxQueueSize:      500,
		FlushFrequencyInMs: 2000,
		MaxIddleTimeInSec: 1,
	}, toolbox.FileLoggerConfig{
       		LogType:           "transaction",
       		FileTemplate:      "/tmp/transaction[yyyyMMdd-hhmm].log",
       		QueueFlashCount:   250,
       		MaxQueueSize:      500,
       		FlushFrequencyInMs:2000,
       		MaxIddleTimeInSec: 1,
       	},
	)

    logger.Log(&toolbox.LogMessage{
        MessageType: "test",
        Message:     message
    })
    
    logger.Log(&toolbox.LogMessage{
            MessageType: "transaction",
            Message:     message
        })
```
		
<a name="BatchLimiter"></a>
### BatchLimiter


This library provides a batch limiter, that enables controling number of active go routines.


```go

     var tasks []*Task
     var batchSize = 4
	 limiter:= toolbox.NewBatchLimiter(batchSize, len(tasks))
   	 for i, _ :=  range tasks {
            go func(task *Task) {
                    limiter.Acquire()
                    defer limiter.Done()
                    task.Run();
        	}(tasks[i])
	}
	limiter.Wait()

```

### AST Based FileSetInfo 


```go
    pkgPath := ""
	source := path.Join(pkgPath)
	filesetInfo, err :=toolbox.NewFileSetInfo(source)
    myType := fileSetInfo.Type("MyType")
    fields := myType.Fields()
    method := myType.Receivers
``` 




## GoCover

[![GoCover](https://gocover.io/github.com/viant/toolbox)](https://gocover.io/github.com/viant/toolbox)

	
	
<a name="License"></a>
## License

The source code is made available under the terms of the Apache License, Version 2, as stated in the file `LICENSE`.

Individual files may be made available under their own specific license,
all compatible with Apache License, Version 2. Please see individual files for details.


<a name="Credits-and-Acknowledgements"></a>

##  Credits and Acknowledgements

**Library Author:** Adrian Witas

**Contributors:**
