package bridge

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/viant/toolbox"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

//HttpBridgeConfig represent http bridge config
type HttpBridgeEndpointConfig struct {
	Port           string
	ReadTimeoutMs  int
	WriteTimeoutMs int
	MaxHeaderBytes int
}

//HttpBridgeProxyRoute represent http proxy route
type HttpBridgeProxyRoute struct {
	Pattern          string
	TargetURL        *url.URL
	ResponseModifier func(*http.Response) error
	Listener         func(request *http.Request, response *http.Response)
}

//HttpBridgeProxyConfig represent proxy config
type HttpBridgeProxyConfig struct {
	MaxIdleConnections    int
	RequestTimeoutMs      int
	KeepAliveTimeMs       int
	TLSHandshakeTimeoutMs int
	BufferPoolSize        int
	BufferSize            int
}

//HttpBridgeConfig represents HttpBridgeConfig config
type HttpBridgeConfig struct {
	Endpoint *HttpBridgeEndpointConfig
	Proxy    *HttpBridgeProxyConfig
	Routes   []*HttpBridgeProxyRoute
}

//ProxyHandlerFactory proxy handler factory
type HttpBridgeProxyHandlerFactory func(proxyConfig *HttpBridgeProxyConfig, route *HttpBridgeProxyRoute) (http.Handler, error)

//HttpBridge represents http bridge
type HttpBridge struct {
	Config   *HttpBridgeConfig
	Server   *http.Server
	Handlers map[string]http.Handler
}

//ListenAndServe start http endpoint
func (r *HttpBridge) ListenAndServe() error {
	return r.Server.ListenAndServe()
}

//ListenAndServe start http endpoint on secure port
func (r *HttpBridge) ListenAndServeTLS(certFile, keyFile string) error {
	return r.Server.ListenAndServeTLS(certFile, keyFile)
}

//NewHttpBridge creates a new instance of NewHttpBridge
func NewHttpBridge(config *HttpBridgeConfig, factory HttpBridgeProxyHandlerFactory) (*HttpBridge, error) {
	mux := http.NewServeMux()
	var handlers = make(map[string]http.Handler)
	for _, route := range config.Routes {
		handler, err := factory(config.Proxy, route)
		if err != nil {
			return nil, err
		}
		mux.Handle(route.Pattern, handler)
		handlers[route.Pattern] = handler
	}
	server := &http.Server{
		Addr:           ":" + config.Endpoint.Port,
		Handler:        mux,
		ReadTimeout:    time.Millisecond * time.Duration(config.Endpoint.ReadTimeoutMs),
		WriteTimeout:   time.Millisecond * time.Duration(config.Endpoint.WriteTimeoutMs),
		MaxHeaderBytes: config.Endpoint.MaxHeaderBytes,
	}
	return &HttpBridge{
		Server:   server,
		Config:   config,
		Handlers: handlers,
	}, nil
}

type handlerWrapper struct {
	Handler http.Handler
}

func (h *handlerWrapper) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.Handler.ServeHTTP(writer, request)
}

//NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyConfig *HttpBridgeProxyConfig, route *HttpBridgeProxyRoute) (http.Handler, error) {
	roundTripper := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(proxyConfig.RequestTimeoutMs) * time.Millisecond,
			KeepAlive: time.Duration(proxyConfig.KeepAliveTimeMs) * time.Millisecond,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(proxyConfig.TLSHandshakeTimeoutMs) * time.Millisecond,
		MaxIdleConnsPerHost: proxyConfig.MaxIdleConnections,
	}

	var director func(*http.Request)

	if route.TargetURL != nil {
		director = func(request *http.Request) {
			request.URL.Scheme = route.TargetURL.Scheme
			request.URL.Host = route.TargetURL.Host
		}
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport:      roundTripper,
		BufferPool:     toolbox.NewBytesBufferPool(proxyConfig.BufferPoolSize, proxyConfig.BufferSize),
		ModifyResponse: route.ResponseModifier,
		Director:       director,
	}
	var handler http.Handler = &handlerWrapper{reverseProxy}
	return handler, nil
}

//HTTPTrip represents recorded round trip.
type HttpTrip struct {
	responseWriter     http.ResponseWriter
	Request            *http.Request
	responseBody       *bytes.Buffer
	responseStatusCode int
}

func (w *HttpTrip) Response() *http.Response {
	return &http.Response{
		Request:    w.Request,
		StatusCode: w.responseStatusCode,
		Header:     w.responseWriter.Header(),
		Body:       ioutil.NopCloser(bytes.NewReader(w.responseBody.Bytes())),
	}
}

func (w *HttpTrip) Write(b []byte) (int, error) {
	w.responseBody.Write(b)
	return w.responseWriter.Write(b)
}

func (w *HttpTrip) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *HttpTrip) WriteHeader(status int) {
	w.responseStatusCode = status
	w.responseWriter.WriteHeader(status)
}

func (w *HttpTrip) Flush() {
	if flusher, ok := w.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *HttpTrip) CloseNotify() <-chan bool {
	if closer, ok := w.responseWriter.(http.CloseNotifier); ok {
		return closer.CloseNotify()
	}
	return make(chan bool, 1)
}

//ListeningTripHandler represents endpoint recording handler
type ListeningTripHandler struct {
	handler         http.Handler
	pool            httputil.BufferPool
	listener        func(request *http.Request, response *http.Response)
	roundTripsMutex *sync.RWMutex
}

func (h *ListeningTripHandler) Notify(roundTrip *HttpTrip) {
	if h.listener != nil {
		h.listener(roundTrip.Request, roundTrip.Response())
	}
}

//drainBody reads all of b to memory and then returns two equivalent (modified version from  httputil)
func (h ListeningTripHandler) drainBody(reader io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {
	if reader == http.NoBody {
		return http.NoBody, http.NoBody, nil
	}
	var buf = new(bytes.Buffer)
	toolbox.CopyWithBufferPool(reader, buf, h.pool)
	return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func (h ListeningTripHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	var err error
	var originalRequest = request.WithContext(request.Context())
	if request.ContentLength > 0 {
		request.Body, originalRequest.Body, err = h.drainBody(request.Body)
		if err != nil {
			log.Printf("failed to serve request :%v due to %v\n", request, err)
			return
		}
	}
	var recordedRoundTrip = &HttpTrip{
		responseWriter: responseWriter,
		Request:        originalRequest,
		responseBody:   new(bytes.Buffer),
	}
	responseWriter = http.ResponseWriter(recordedRoundTrip)
	defer h.Notify(recordedRoundTrip)
	h.handler.ServeHTTP(responseWriter, request)
}

func NewListeningHandler(handler http.Handler, bufferPoolSize, bufferSize int, listener func(request *http.Request, response *http.Response)) *ListeningTripHandler {
	var result = &ListeningTripHandler{
		handler:         handler,
		listener:        listener,
		pool:            toolbox.NewBytesBufferPool(bufferPoolSize, bufferSize),
		roundTripsMutex: &sync.RWMutex{},
	}
	return result
}

func NewProxyRecordingHandler(proxyConfig *HttpBridgeProxyConfig, route *HttpBridgeProxyRoute) (http.Handler, error) {
	handler, err := NewProxyHandler(proxyConfig, route)
	if err != nil {
		return nil, err
	}
	response := NewListeningHandler(handler, proxyConfig.BufferPoolSize, proxyConfig.BufferSize, route.Listener)
	return response, nil
}

func AsListeningTripHandler(handler http.Handler) *ListeningTripHandler {
	if result, ok := handler.(*ListeningTripHandler); ok {
		return result
	}
	return nil
}

//HttpRequest represents JSON serializable http request
type HttpRequest struct {
	Method      string      `json:",omitempty"`
	URL         string      `json:",omitempty"`
	Header      http.Header `json:",omitempty"`
	Body        string      `json:",omitempty"`
	ThinkTimeMs int         `json:",omitempty"`
}

//NewHTTPRequest create a new instance of http request
func NewHTTPRequest(method, url, body string, header http.Header) *HttpRequest {
	return &HttpRequest{
		Method: method,
		URL:    url,
		Body:   body,
		Header: header,
	}
}

//HttpResponse represents JSON serializable http response
type HttpResponse struct {
	Code   int
	Header http.Header `json:",omitempty"`
	Body   string      `json:",omitempty"`
}

func ReaderAsText(reader io.Reader) string {
	if reader == nil {
		return ""
	}
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	if isBinary(body) {
		buf := new(bytes.Buffer)
		encoder := base64.NewEncoder(base64.StdEncoding, buf)
		encoder.Write(body)
		encoder.Close()
		return fmt.Sprintf("base64:%v", string(buf.Bytes()))

	} else if len(body) > 0 {
		return fmt.Sprintf("text:%v", string(body))
	}
	return ""
}

func isBinary(input []byte) bool {
	for i, w := 0, 0; i < len(input); i += w {
		runeValue, width := utf8.DecodeRune(input[i:])
		if unicode.IsControl(runeValue) {
			return true
		}
		w = width
	}
	return false
}

func writeData(filename string, source interface{}, printStrOut bool) error {
	if toolbox.FileExists(filename) {
		os.Remove(filename)
	}

	logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v, %v", err, filename)
	}
	defer logfile.Close()

	buf, err := json.MarshalIndent(source, "", "\t")
	if err != nil {
		return err
	}
	_, err = logfile.Write(buf)

	if printStrOut {
		fmt.Printf("%v: %v\n", filename, string(buf))
	}

	return err
}

//HttpFileRecorder returns http route listener that will record request response to the passed in directory
func HttpFileRecorder(directory string, printStdOut bool) func(request *http.Request, response *http.Response) {
	tripCounter := 0

	err := toolbox.CreateDirIfNotExist(directory)
	if err != nil {
		fmt.Printf("failed to create directory%v %v\n, ", err, directory)
	}
	return func(request *http.Request, response *http.Response) {
		var body string
		if request.Body != nil {
			body = ReaderAsText(request.Body)
		}
		httpRequest := &HttpRequest{
			Method: request.Method,
			URL:    request.URL.String(),
			Header: request.Header,
			Body:   body,
		}

		err = writeData(path.Join(directory, fmt.Sprintf("%T-%v.json", *httpRequest, tripCounter)), httpRequest, printStdOut)
		if err != nil {
			fmt.Printf("failed to write request %v %v\n, ", err, request)
		}

		body = ReaderAsText(response.Body)
		request.Body = nil
		httpResponse := &HttpResponse{
			Code:   response.StatusCode,
			Header: response.Header,
			Body:   body,
		}

		err = writeData(path.Join(directory, fmt.Sprintf("%T-%v.json", *httpResponse, tripCounter)), httpResponse, printStdOut)
		if err != nil {
			fmt.Printf("failed to write response %v %v\n, ", err, response)
		}

		tripCounter++
	}

}
