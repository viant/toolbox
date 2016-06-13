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

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

var jsonContentType = "application/json"
var textPlainContentType = "text/plain"

//ServiceRouting represents a simple web services routing rule, which is matched with http request
type ServiceRouting struct {
	URI                 string      //matching uri
	Handler             interface{} //has to be func
	HTTPMethod          string
	Parameters          []string
	ContentTypeEncoders map[string]EncoderFactory //content type encoder factory
	ContentTypeDecoders map[string]DecoderFactory //content type decoder factory
}

func (sr ServiceRouting) getDecoderFactory(contentType string) DecoderFactory {
	if sr.ContentTypeDecoders != nil {
		if factory, found := sr.ContentTypeDecoders[contentType]; found {
			return factory
		}
	}
	return NewJSONDecoderFactory()
}

func (sr ServiceRouting) getEncoderFactory(contentType string) EncoderFactory {
	if sr.ContentTypeDecoders != nil {
		if factory, found := sr.ContentTypeEncoders[contentType]; found {
			return factory
		}
	}
	return NewJSONEncoderFactory()
}

func (sr ServiceRouting) extractParameterFromBody(parameterName string, targetType reflect.Type, request *http.Request) (interface{}, error) {
	targetValuePointer := reflect.New(targetType)
	contentType := getContentTypeOrJSONContentType(request.Header.Get("Content-Type"))
	decoderFactory := sr.getDecoderFactory(contentType)
	decoder := decoderFactory.Create(request.Body)
	if !strings.Contains(parameterName, ":") {
		err := decoder.Decode(targetValuePointer.Interface())
		if err != nil {
			return nil, fmt.Errorf("Unable to extract %Tv due to %v", targetValuePointer.Interface(), err)
		}
	} else {
		var valueMap = make(map[string]interface{})
		pair := strings.SplitN(parameterName, ":", 2)
		valueMap[pair[1]] = targetValuePointer.Interface()
		err := decoder.Decode(&valueMap)
		if err != nil {
			return nil, fmt.Errorf("Unable to extract %T due to %v", targetValuePointer.Interface(), err)
		}
	}
	return targetValuePointer.Interface(), nil
}

func (sr ServiceRouting) extractParameters(request *http.Request, response http.ResponseWriter) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	functionSignature := GetFuncSignature(sr.Handler)
	uriParameters, _ := ExtractURIParameters(sr.URI, request.RequestURI)
	for _, name := range sr.Parameters {
		if value, found := uriParameters[name]; found {
			if strings.Contains(value, ",") {
				result[name] = strings.Split(value, ",")
			} else {
				result[name] = value
			}
			continue
		}
		value := request.Form.Get(name)
		if len(value) > 0 {
			result[name] = value
		} else {
			break
		}
	}
	if HasSliceAnyElements(sr.Parameters, "@httpRequest") {
		result["@httpRequest"] = request
	}
	if HasSliceAnyElements(sr.Parameters, "@httpResponseWriter") {
		result["@httpResponseWriter"] = response
	}

	if request.ContentLength > 0 {
		for i, parameter := range sr.Parameters {
			if _, found := result[parameter]; !found {
				value, err := sr.extractParameterFromBody(parameter, functionSignature[i], request)
				if err != nil {
					return nil, fmt.Errorf("Failed to extract parameters for %v %v due to %v", sr.HTTPMethod, sr.URI, err)
				}
				result[parameter] = value
				break
			}
		}
	}
	return result, nil
}

//ServiceRouter represents routing rule
type ServiceRouter struct {
	serviceRouting []ServiceRouting
}

func (r *ServiceRouter) match(request *http.Request) []ServiceRouting {
	var result = make([]ServiceRouting, 0)
	for _, candidate := range r.serviceRouting {
		if candidate.HTTPMethod == request.Method {
			_, matched := ExtractURIParameters(candidate.URI, request.RequestURI)
			if matched {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func getContentTypeOrJSONContentType(contentType string) string {
	if contentType == textPlainContentType || contentType == jsonContentType || contentType == "" {
		return jsonContentType
	}
	return contentType
}

//WriteResponse writes response to response writer, it used encoder factory to encode passed in response to the writer, it sets back request contenttype to response.
func (r *ServiceRouter) WriteResponse(encoderFactory EncoderFactory, response interface{}, request *http.Request, responseWriter http.ResponseWriter) error {
	requestContentType := request.Header.Get("Content-Type")
	responseContentType := getContentTypeOrJSONContentType(requestContentType)
	encoder := encoderFactory.Create(responseWriter)
	responseWriter.Header().Set("Content-Type", responseContentType)
	err := encoder.Encode(response)
	if err != nil {
		return fmt.Errorf("Failed to encode response %v, due to %v", response, err)
	}
	return nil
}

//Route matches  service routing by http method , and number of parameters, then it call routing method, and sent back its response.
func (r *ServiceRouter) Route(response http.ResponseWriter, request *http.Request) error {
	candidates := r.match(request)
	if len(candidates) == 0 {
		var uriTemplates = make([]string, 0)
		for _, routing := range r.serviceRouting {
			uriTemplates = append(uriTemplates, routing.URI)
		}
		return fmt.Errorf("Failed to route request - unable to match %v with one of %v", request.RequestURI, strings.Join(uriTemplates, ","))
	}
	var finalError error

	for _, serviceRouting := range candidates {
		parameterValues, err := serviceRouting.extractParameters(request, response)
		if err != nil {
			finalError = fmt.Errorf("unable to extract parameters due to %v", err)
			continue
		}

		functionParameters, err := BuildFunctionParameters(serviceRouting.Handler, serviceRouting.Parameters, parameterValues)
		if err != nil {
			finalError = fmt.Errorf("unable to build function parameters %T due to %v", serviceRouting.Handler, err)
			continue
		}
		result := CallFunction(serviceRouting.Handler, functionParameters...)
		if len(result) > 0 {
			requestContentType := request.Header.Get("Content-Type")
			responseContentType := getContentTypeOrJSONContentType(requestContentType)
			factory := serviceRouting.getEncoderFactory(responseContentType)
			err := r.WriteResponse(factory, result[0], request, response)
			if err != nil {
				return fmt.Errorf("Failed to write response response %v, due to %v", result[0], err)
			}
			return nil
		}
		response.Header().Set("Content-Type", textPlainContentType)

	}
	if finalError != nil {
		return fmt.Errorf("Failed to route request - %v", finalError)
	}
	return nil
}

//NewServiceRouter creates a new service router, is takes list of service routing as arguments
func NewServiceRouter(serviceRouting ...ServiceRouting) *ServiceRouter {
	return &ServiceRouter{serviceRouting}
}

//RouteToService calls web service url, with passed in json request, and encodes http json response into passed response
func RouteToService(method, url string, request, response interface{}) (err error) {
	return RouteToServiceWithCustomFormat(method, url, request, response, NewJSONEncoderFactory(), NewJSONDecoderFactory())
}

//RouteToServiceWithCustomFormat calls web service url, with passed in custom format request, and encodes custom format http response into passed response
func RouteToServiceWithCustomFormat(method, url string, request, response interface{}, encoderFactory EncoderFactory, decoderFactory DecoderFactory) (err error) {
	var buffer *bytes.Buffer
	if request != nil {
		buffer = new(bytes.Buffer)
		err := encoderFactory.Create(buffer).Encode(&request)
		if err != nil {
			return fmt.Errorf("Failed to encode request: %v due to ", err)
		}
	}
	var serverResponse *http.Response
	switch strings.ToLower(method) {
	case "get":
		serverResponse, err = http.Get(url)
	case "post":
		serverResponse, err = http.Post(url, jsonContentType, buffer)
	case "delete":
		var httpRequest *http.Request
		httpRequest, err = http.NewRequest("DELETE", url, nil)
		serverResponse, err = http.DefaultClient.Do(httpRequest)
	default:
		err = fmt.Errorf("%v is not yet supproted", method)
	}
	if err != nil {
		return fmt.Errorf("Failed to get response %v %v", err, serverResponse.Header.Get("error"))
	}
	err = decoderFactory.Create(serverResponse.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("Failed to decode response to %T : %v, %v", response, err, serverResponse.Header.Get("error"))
	}
	return nil
}
