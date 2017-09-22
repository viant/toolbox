package bridge

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
	"github.com/viant/toolbox"
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
	director := func(request *http.Request) {
		request.URL.Scheme = route.TargetURL.Scheme
		request.URL.Host = route.TargetURL.Host
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport:      roundTripper,
		BufferPool:     toolbox.NewBytesBufferPool(proxyConfig.BufferPoolSize, proxyConfig.BufferSize),
		ModifyResponse: route.ResponseModifier,
		Director:       director,
	}
	return reverseProxy, nil
}

//HttpTrip represents recorded round trip.
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
			log.Printf("Faled to serve request :%v due to %v\n", request, err)
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
