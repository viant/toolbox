package secret

import (
	"testing"
	"time"
)

func TestSecretKey_Secret(t *testing.T) {
	ReadUserAndPassword(time.Second * 10)

}
