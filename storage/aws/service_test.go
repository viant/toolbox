package aws_test

import (
	"compress/gzip"
	"fmt"
	"github.com/stretchr/testify/assert"
	aws "github.com/viant/toolbox/storage/aws"
	"io/ioutil"
	"strings"
	"testing"
)

func TestService_List(t *testing.T) {

	fmt.Print("Has service\n")
	config := &aws.Config{
		Region: "us-east-1",
		Key:    "AKIAIR2KFFVSWQDEWVAA",
		Secret: "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",
		Token:  "",
	}

	service := aws.NewService(config)
	assert.NotNil(t, service)

	fmt.Print("Has service\n")
	result, err := service.List("s3://r.ci.com/adlog/2017/06/12/07/")
	assert.Nil(t, err)

	for _, o := range result {
		fmt.Printf("R: %v\n", o.URL())

	}

	byteReader, err := service.Download(result[0])
	assert.Nil(t, err)
	reader, err := gzip.NewReader(byteReader)
	assert.Nil(t, err)
	logBytes, err := ioutil.ReadAll(reader)
	lines := strings.Split(string(logBytes), "\n")

	fmt.Printf("%v", strings.Join(lines[0:10], "\n"))

	assert.True(t, false)

}
