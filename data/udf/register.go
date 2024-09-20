package udf

import "github.com/viant/toolbox/data"

var Predefined = data.Map{
	"AsInt":                  AsInt,
	"AsString":               AsString,
	"AsFloat":                AsFloat,
	"AsFloat32":              AsFloat32,
	"AsFloat32Ptr":           AsFloat32Ptr,
	"AsBool":                 AsBool,
	"AsMap":                  AsMap,
	"AsData":                 AsData,
	"AsCollection":           AsCollection,
	"AsJSON":                 AsJSON,
	"Type":                   Type,
	"Join":                   Join,
	"Split":                  Split,
	"Keys":                   Keys,
	"StringKeys":             StringKeys,
	"Values":                 Values,
	"Length":                 Length,
	"Len":                    Length,
	"IndexOf":                IndexOf,
	"FormatTime":             FormatTime,
	"QueryEscape":            QueryEscape,
	"QueryUnescape":          QueryUnescape,
	"Base64Encode":           Base64Encode,
	"Base64Decode":           Base64Decode,
	"Base64RawURLEncode":     Base64RawURLEncode,
	"Base64RawURLDecode":     Base64RawURLDecode,
	"Base64DecodeText":       Base64DecodeText,
	"TrimSpace":              TrimSpace,
	"Elapsed":                Elapsed,
	"Sum":                    Sum,
	"Count":                  Count,
	"AsNumber":               AsNumber,
	"Select":                 Select,
	"Rand":                   Rand,
	"Concat":                 Concat,
	"Merge":                  Merge,
	"AsStringMap":            AsStringMap,
	"PackInt32sTo64":         PackInt32sTo64,
	"Replace":                Replace,
	"ToLower":                ToLower,
	"ToUpper":                ToUpper,
	"AsNewLineDelimitedJSON": AsNewLineDelimitedJSON,
	"LoadJSON":               LoadJSON,
}

func Register(aMap data.Map) {
	if aMap == nil {
	}
	udfs, ok := aMap[data.UDFKey]
	if !ok {
		aMap[data.UDFKey] = Predefined
		return
	}
	var udfMap = data.Map{}
	if prevMap, ok := udfs.(data.Map); ok {
		for k, v := range prevMap {
			udfMap[k] = v
		}
	}
	for k, v := range Predefined {
		udfMap[k] = v
	}
	aMap[data.UDFKey] = udfMap
}
