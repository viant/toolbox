package toolbox

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

var jsonContentType = "application/json"
var textPlainContentType = "text/plain"

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

var httpMethods = map[string]bool{
	MethodDelete:  true,
	MethodGet:     true,
	MethodPatch:   true,
	MethodPost:    true,
	MethodPut:     true,
	MethodHead:    true,
	MethodTrace:   true,
	MethodOptions: true,
}

//DefaultEncoderFactory  - NewJSONEncoderFactory
var DefaultEncoderFactory = NewJSONEncoderFactory()

//DefaultDecoderFactory - NewJSONDecoderFactory
var DefaultDecoderFactory = NewJSONDecoderFactory()

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
	return DefaultDecoderFactory
}

func (sr ServiceRouting) getEncoderFactory(contentType string) EncoderFactory {
	if sr.ContentTypeDecoders != nil {
		if factory, found := sr.ContentTypeEncoders[contentType]; found {
			return factory
		}
	}
	return DefaultEncoderFactory
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
	request.ParseForm()
	functionSignature := GetFuncSignature(sr.Handler)
	uriParameters, _ := ExtractURIParameters(sr.URI, request.RequestURI)
	for _, name := range sr.Parameters {
		value, found := uriParameters[name]
		if found {
			if strings.Contains(value, ",") {
				result[name] = strings.Split(value, ",")
			} else {
				result[name] = value
			}
			continue
		}

		value = request.Form.Get(name)
		if len(value) > 0 {
			result[name] = value
		} else {
			continue
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
	if strings.Contains(contentType, textPlainContentType) || strings.Contains(contentType, jsonContentType) || contentType == "" {
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
	if _, found := httpMethods[strings.ToUpper(method)]; !found {
		return errors.New("Unsupported method:" + method)
	}
	var buffer *bytes.Buffer
	if request != nil {
		buffer = new(bytes.Buffer)
		err := encoderFactory.Create(buffer).Encode(&request)
		if err != nil {
			return fmt.Errorf("Failed to encode request: %v due to ", err)
		}
	}
	var serverResponse *http.Response
	var httpRequest *http.Request
	httpMethod := strings.ToUpper(method)
	if request != nil {
		httpRequest, err = http.NewRequest(httpMethod, url, buffer)
		if err != nil {
			return err
		}
		httpRequest.Header.Set("Content-Type", jsonContentType)
	} else {
		httpRequest, err = http.NewRequest(httpMethod, url, nil)
		if err != nil {
			return err
		}
	}

	serverResponse, err = http.DefaultClient.Do(httpRequest)
	if err != nil && serverResponse != nil {
		return fmt.Errorf("Failed to get response %v %v", err, serverResponse.Header.Get("error"))
	}

	if response != nil  {
		if serverResponse == nil || serverResponse.Body == nil {
			return fmt.Errorf("Failed to recieve response %v", err)
		}
		body, err := ioutil.ReadAll(serverResponse.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response %v", err)
		}
		err = decoderFactory.Create(strings.NewReader(string(body))).Decode(response)
		if err != nil {
			return fmt.Errorf("Failed to decode response to %T: body: %v: %v, %v", response, string(body), err, serverResponse.Header.Get("error"))
		}
	}
	return nil
}
