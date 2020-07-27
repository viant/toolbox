package sampler

import (
	"math/rand"
	"sync"
	"time"
)

//Service represents a sampler service
type Service struct {
	rand            *rand.Rand
	acceptThreshold int32
	mux             sync.Mutex
	PCT             float64
}

//Accept accept sample meeting accept gaol PCT
func (s *Service) Accept() bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	n := s.rand.Int31n(100000)
	return n < s.acceptThreshold
}


//New creates a pct sampler
func New(acceptPCT float64) *Service {
	source := rand.NewSource(time.Now().UnixNano())
	return &Service{
		PCT:             acceptPCT,
		rand:            rand.New(source),
		acceptThreshold: int32(acceptPCT * 1000.0),
	}
}
