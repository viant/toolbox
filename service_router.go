package toolbox

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	jsonContentType       = "application/json"
	yamlContentTypeSuffix = "/yaml"
	textPlainContentType  = "text/plain"
	contentTypeHeader     = "Content-Type"
)

const (
	//MethodGet HTTP GET meothd
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

//HandlerInvoker method is responsible  of passing required parameters to router handler.
type HandlerInvoker func(serviceRouting *ServiceRouting, request *http.Request, response http.ResponseWriter, parameters map[string]interface{}) error

//DefaultEncoderFactory  - NewJSONEncoderFactory
var DefaultEncoderFactory = NewJSONEncoderFactory()

//DefaultDecoderFactory - NewJSONDecoderFactory
var DefaultDecoderFactory = NewJSONDecoderFactory()

//YamlDefaultEncoderFactory  - NewYamlEncoderFactory
var YamlDefaultEncoderFactory = NewYamlEncoderFactory()

//YamlDefaultDecoderFactory - NewYamlDecoderFactory
var YamlDefaultDecoderFactory = NewFlexYamlDecoderFactory()

//ServiceRouting represents a simple web services routing rule, which is matched with http request
type ServiceRouting struct {
	URI                 string      //matching uri
	Handler             interface{} //has to be func
	HTTPMethod          string
	Parameters          []string
	ContentTypeEncoders map[string]EncoderFactory //content type encoder factory
	ContentTypeDecoders map[string]DecoderFactory //content type decoder factory
	HandlerInvoker      HandlerInvoker            //optional function that will be used instead of reflection to invoke a handler.
}

func (sr ServiceRouting) getDecoderFactory(contentType string) DecoderFactory {
	if sr.ContentTypeDecoders != nil {
		if factory, found := sr.ContentTypeDecoders[contentType]; found {
			return factory
		}
	}
	if strings.HasSuffix(contentType, yamlContentTypeSuffix) {
		return YamlDefaultDecoderFactory
	}
	return DefaultDecoderFactory
}

func (sr ServiceRouting) getEncoderFactory(contentType string) EncoderFactory {
	if sr.ContentTypeDecoders != nil {
		if factory, found := sr.ContentTypeEncoders[contentType]; found {
			return factory
		}
	}
	if strings.HasSuffix(contentType, yamlContentTypeSuffix) {
		return YamlDefaultEncoderFactory
	}
	return DefaultEncoderFactory
}

func (sr ServiceRouting) extractParameterFromBody(parameterName string, targetType reflect.Type, request *http.Request) (interface{}, error) {
	targetValuePointer := reflect.New(targetType)
	contentType := getContentTypeOrJSONContentType(request.Header.Get(contentTypeHeader))
	decoderFactory := sr.getDecoderFactory(contentType)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	decoder := decoderFactory.Create(bytes.NewReader(body))

	if !strings.Contains(parameterName, ":") {
		err := decoder.Decode(targetValuePointer.Interface())
		if err != nil {
			return nil, fmt.Errorf("unable to extract %T due to: %v, body: !%s!", targetValuePointer.Interface(), err, body)
		}
	} else {
		var valueMap = make(map[string]interface{})
		pair := strings.SplitN(parameterName, ":", 2)
		valueMap[pair[1]] = targetValuePointer.Interface()
		err := decoder.Decode(&valueMap)
		if err != nil {
			return nil, fmt.Errorf("unable to extract %T due to %v", targetValuePointer.Interface(), err)
		}
	}
	return targetValuePointer.Interface(), nil
}

func (sr ServiceRouting) extractParameters(request *http.Request, response http.ResponseWriter) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	_ = request.ParseForm()
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
					return nil, fmt.Errorf("failed to extract parameters for %v %v due to %v", sr.HTTPMethod, sr.URI, err)
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
	serviceRouting []*ServiceRouting
}

func (r *ServiceRouter) match(request *http.Request) []*ServiceRouting {
	var result = make([]*ServiceRouting, 0)
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

//Route matches  service routing by http method , and number of parameters, then it call routing method, and sent back its response.
func (r *ServiceRouter) Route(response http.ResponseWriter, request *http.Request) error {
	candidates := r.match(request)
	if len(candidates) == 0 {
		var uriTemplates = make([]string, 0)
		for _, routing := range r.serviceRouting {
			uriTemplates = append(uriTemplates, routing.URI)
		}
		return fmt.Errorf("failed to route request - unable to match %v with one of %v", request.RequestURI, strings.Join(uriTemplates, ","))
	}
	var finalError error

	for _, serviceRouting := range candidates {

		parameterValues, err := serviceRouting.extractParameters(request, response)
		if err != nil {
			finalError = fmt.Errorf("unable to extract parameters due to %v", err)
			continue
		}

		if serviceRouting.HandlerInvoker != nil {
			err := serviceRouting.HandlerInvoker(serviceRouting, request, response, parameterValues)
			if err != nil {
				finalError = fmt.Errorf("unable to extract parameters due to %v", err)
			}
			continue
		}

		functionParameters, err := BuildFunctionParameters(serviceRouting.Handler, serviceRouting.Parameters, parameterValues)
		if err != nil {
			finalError = fmt.Errorf("unable to build function parameters %T due to %v", serviceRouting.Handler, err)
			continue
		}

		result := CallFunction(serviceRouting.Handler, functionParameters...)
		if len(result) > 0 {
			err = WriteServiceRoutingResponse(response, request, serviceRouting, result[0])
			if err != nil {
				return fmt.Errorf("failed to write response response %v, due to %v", result[0], err)
			}
			return nil
		}
		response.Header().Set(contentTypeHeader, textPlainContentType)
	}
	if finalError != nil {
		return fmt.Errorf("failed to route request - %v", finalError)
	}
	return nil
}

//WriteServiceRoutingResponse writes service router response
func WriteServiceRoutingResponse(response http.ResponseWriter, request *http.Request, serviceRouting *ServiceRouting, result interface{}) error {
	if result == nil {
		result = struct{}{}
	}
	statusCodeAccessor, ok := result.(StatucCodeAccessor)
	if ok {
		statusCode := statusCodeAccessor.GetStatusCode()
		if statusCode > 0 && statusCode != http.StatusOK {
			response.WriteHeader(statusCode)
			return nil
		}
	}
	contentTypeAccessor, ok := result.(ContentTypeAccessor)
	var responseContentType string
	if ok {
		responseContentType = contentTypeAccessor.GetContentType()
	}
	if responseContentType == "" {
		requestContentType := request.Header.Get(contentTypeHeader)
		responseContentType = getContentTypeOrJSONContentType(requestContentType)
	}
	encoderFactory := serviceRouting.getEncoderFactory(responseContentType)
	encoder := encoderFactory.Create(response)
	response.Header().Set(contentTypeHeader, responseContentType)
	err := encoder.Encode(result)
	if err != nil {
		return fmt.Errorf("failed to encode response %v, due to %v", response, err)
	}
	return nil
	if err != nil {
		return fmt.Errorf("failed to write response response %v, due to %v", result, err)
	}
	return nil
}

//WriteResponse writes response to response writer, it used encoder factory to encode passed in response to the writer, it sets back request contenttype to response.
func (r *ServiceRouter) WriteResponse(encoderFactory EncoderFactory, response interface{}, request *http.Request, responseWriter http.ResponseWriter) error {
	requestContentType := request.Header.Get(contentTypeHeader)
	responseContentType := getContentTypeOrJSONContentType(requestContentType)
	encoder := encoderFactory.Create(responseWriter)
	responseWriter.Header().Set(contentTypeHeader, responseContentType)
	err := encoder.Encode(response)
	if err != nil {
		return fmt.Errorf("failed to encode response %v, due to %v", response, err)
	}
	return nil
}

//NewServiceRouter creates a new service router, is takes list of service routing as arguments
func NewServiceRouter(serviceRouting ...ServiceRouting) *ServiceRouter {
	var routings = make([]*ServiceRouting, 0)
	for i := range serviceRouting {
		routings = append(routings, &serviceRouting[i])
	}
	return &ServiceRouter{routings}
}

//RouteToService calls web service url, with passed in json request, and encodes http json response into passed response
func RouteToService(method, url string, request, response interface{}, options ...*HttpOptions) (err error) {
	client, err := NewToolboxHTTPClient(options...)
	if err != nil {
		return err
	}
	return client.Request(method, url, request, response, NewJSONEncoderFactory(), NewJSONDecoderFactory())
}

type HttpOptions struct {
	Key   string
	Value interface{}
}

func NewHttpClient(options ...*HttpOptions) (*http.Client, error) {
	if len(options) == 0 {
		return http.DefaultClient, nil
	}

	var (
		// Default values matching DefaultHttpClient
		RequestTimeoutMs        = 30 * time.Second
		KeepAliveTimeMs         = 30 * time.Second
		TLSHandshakeTimeoutMs   = 10 * time.Second
		ExpectContinueTimeout   = 1 * time.Second
		IdleConnTimeout         = 90 * time.Second
		DualStack               = true
		MaxIdleConnsPerHost     = http.DefaultMaxIdleConnsPerHost
		MaxIdleConns            = 100
		FollowRedirects         = true
		ResponseHeaderTimeoutMs time.Duration
		TimeoutMs               time.Duration
	)

	for _, option := range options {
		switch option.Key {
		case "RequestTimeoutMs":
			RequestTimeoutMs = time.Duration(AsInt(option.Value)) * time.Millisecond
		case "TimeoutMs":
			TimeoutMs = time.Duration(AsInt(option.Value)) * time.Millisecond
		case "KeepAliveTimeMs":
			KeepAliveTimeMs = time.Duration(AsInt(option.Value)) * time.Millisecond
		case "TLSHandshakeTimeoutMs":
			KeepAliveTimeMs = time.Duration(AsInt(option.Value)) * time.Millisecond
		case "ResponseHeaderTimeoutMs":
			ResponseHeaderTimeoutMs = time.Duration(AsInt(option.Value)) * time.Millisecond
		case "MaxIdleConns":
			MaxIdleConns = AsInt(option.Value)
		case "MaxIdleConnsPerHost":
			MaxIdleConnsPerHost = AsInt(option.Value)
		case "DualStack":
			DualStack = AsBoolean(option.Value)
		case "FollowRedirects":
			FollowRedirects = AsBoolean(option.Value)
		default:
			return nil, fmt.Errorf("Invalid option: %v", option.Key)

		}
	}
	roundTripper := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   RequestTimeoutMs,
			KeepAlive: KeepAliveTimeMs,
			DualStack: DualStack,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		ExpectContinueTimeout: ExpectContinueTimeout,
		IdleConnTimeout:       IdleConnTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeoutMs,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
		ResponseHeaderTimeout: ResponseHeaderTimeoutMs,
	}

	client := &http.Client{
		Transport: &roundTripper,
		Timeout:   TimeoutMs,
	}

	if !FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client, nil
}

// ToolboxHTTPClient contains preconfigured http client
type ToolboxHTTPClient struct {
	httpClient *http.Client
}

// NewToolboxHTTPClient instantiate new client with provided options
func NewToolboxHTTPClient(options ...*HttpOptions) (*ToolboxHTTPClient, error) {
	client, err := NewHttpClient(options...)
	if err != nil {
		return nil, err
	}
	return &ToolboxHTTPClient{client}, nil
}

// Request sends http request using the existing client
func (c *ToolboxHTTPClient) Request(method, url string, request, response interface{}, encoderFactory EncoderFactory, decoderFactory DecoderFactory) (err error) {
	if _, found := httpMethods[strings.ToUpper(method)]; !found {
		return errors.New("unsupported method:" + method)
	}
	var buffer *bytes.Buffer

	if request != nil {
		buffer = new(bytes.Buffer)
		if IsString(request) {
			buffer.Write([]byte(AsString(request)))
		} else {
			err := encoderFactory.Create(buffer).Encode(&request)
			if err != nil {
				return fmt.Errorf("failed to encode request: %v due to ", err)
			}
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
		httpRequest.Header.Set(contentTypeHeader, jsonContentType)
	} else {
		httpRequest, err = http.NewRequest(httpMethod, url, nil)
		if err != nil {
			return err
		}
	}
	serverResponse, err = c.httpClient.Do(httpRequest)
	if serverResponse != nil {
		// must close we have serverResponse to avoid fd leak
		defer serverResponse.Body.Close()
	}
	if err != nil && serverResponse != nil {
		return fmt.Errorf("failed to get response %v %v", err, serverResponse.Header.Get("error"))
	}

	if response != nil {
		updateResponse(serverResponse, response)
		if serverResponse == nil {
			return fmt.Errorf("failed to receive response %v", err)
		}
		var errorPrefix = fmt.Sprintf("failed to process response: %v, ", serverResponse.StatusCode)
		body, err := ioutil.ReadAll(serverResponse.Body)
		if err != nil {
			return fmt.Errorf("%v unable read body %v", errorPrefix, err)
		}
		if len(body) == 0 {
			return fmt.Errorf("%v response body was empty", errorPrefix)
		}

		if serverResponse.StatusCode == http.StatusNotFound {
			updateResponse(serverResponse, response)
			return nil
		}

		if int(serverResponse.StatusCode/100)*100 == http.StatusInternalServerError {
			return errors.New(string(body))
		}
		err = decoderFactory.Create(strings.NewReader(string(body))).Decode(response)
		if err != nil {
			return fmt.Errorf("%v. unable decode response as %T: body: %v: %v", errorPrefix, response, string(body), err)
		}
		updateResponse(serverResponse, response)
	}
	return nil
}

//StatucCodeMutator client side reponse optional interface
type StatucCodeMutator interface {
	SetStatusCode(code int)
}

//StatucCodeAccessor server side response accessor
type StatucCodeAccessor interface {
	GetStatusCode() int
}

//ContentTypeMutator client side reponse optional interface
type ContentTypeMutator interface {
	SetContentType(contentType string)
}

//ContentTypeAccessor server side response accessor
type ContentTypeAccessor interface {
	GetContentType() string
}

//updateResponse update response with content type and status code if applicable
func updateResponse(httpResponse *http.Response, response interface{}) {
	if response == nil {
		return
	}
	statusCodeMutator, ok := response.(StatucCodeMutator)
	if ok {
		statusCodeMutator.SetStatusCode(httpResponse.StatusCode)
	}
	contentTypeMutator, ok := response.(ContentTypeMutator)
	if ok {
		contentTypeMutator.SetContentType(httpResponse.Header.Get(contentTypeHeader))
	}
}
