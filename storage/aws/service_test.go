package aws_test

import (
	"github.com/stretchr/testify/assert"
	aws "github.com/viant/toolbox/storage/aws"
	"testing"
)

func TestService_List(t *testing.T) {

	config := &aws.Config{
		Region:  "us-east-1",
		Key:     "***",
		Secrect: "**",
		Token:   "",
	}

	service := aws.NewService(config)
	assert.NotNil(t, service)
	//result, err := service.List("s3://bucket/file.gz")
	//assert.Nil(t, err)
	//
	//for _, o := range result {
	//	fmt.Printf("R: %v\n", o.URL())
	//
	//}
	//
	//byteReader, err := service.Download(result[0])
	//assert.Nil(t, err)
	//reader, err := gzip.NewReader(byteReader)
	//assert.Nil(t, err)
	//logBytes, err := ioutil.ReadAll(reader)
	//lines := strings.Split(string(logBytes), "\n")
	//
	//fmt.Printf("%v", strings.Join(lines [0:10], "\n"))
}
