package scp_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/storage/scp"
	"net/url"
	"testing"
)

// -rw-r--r--  1 awitas  wheel  6668 Oct 25 11:41:44 2017 /build/dcm/dcm-server/target/classes/dcm.properties

func Test_ExtractFileInfo(t *testing.T) {

	var parserURL, _ = url.Parse("scp://127.0.0.1/")

	{
		var parser = scp.Parser{IsoTimeStyle: false}
		objects, err := parser.Parse(parserURL, "-rw-r--r--  1 awitas  1742120565   414 Jun  8 14:14:08 2017 f.pub", true)
		if assert.Nil(t, err) {
			var object = objects[0]
			assert.Equal(t, "scp://127.0.0.1/f.pub", object.URL())
			assert.Equal(t, int64(1496931248), object.FileInfo().ModTime().Unix())
			assert.Equal(t, int64(414), object.FileInfo().Size())
			assert.Equal(t, true, object.IsContent())
		}
	}
	{
		parserURL, _ = url.Parse("scp://127.0.0.1:22/")
		var parser = scp.Parser{IsoTimeStyle: true}
		var objects, err = parser.Parse(parserURL, "-rw-r--r-- 1 awitas awitas 2002 2017-11-04 22:29:33.363458941 +0000 aerospikeciads_aerospike.conf", true)
		if assert.Nil(t, err) {
			var object = objects[0]
			assert.Equal(t, "scp://127.0.0.1:22/aerospikeciads_aerospike.conf", object.URL())
			assert.Equal(t, int64(1509834573), object.FileInfo().ModTime().Unix())
			assert.Equal(t, int64(2002), object.FileInfo().Size())
			assert.Equal(t, true, object.IsContent())
		}
	}

	{
		parserURL, _ = url.Parse("scp://127.0.0.1:22/")
		var parser = scp.Parser{IsoTimeStyle: true}
		var objects, err = parser.Parse(parserURL, `-rw-------  1 myusername  MYGROUP\\Domain Users  1679 Mar  8 15:27:22 2019 /Users/myusername/.ssh/git_id_rsa`, true)
		if assert.Nil(t, err) {
			var object = objects[0]
			assert.Equal(t, "scp://127.0.0.1:22/Users/myusername/.ssh/git_id_rsa", object.URL())
			assert.Equal(t, int64(1552058842), object.FileInfo().ModTime().Unix())
			assert.Equal(t, int64(1679), object.FileInfo().Size())
			assert.Equal(t, true, object.IsContent())
		}

	}
	{
		parserURL, _ = url.Parse("scp://127.0.0.1:22/")
		var parser = scp.Parser{IsoTimeStyle: true}
		var objects, err = parser.Parse(parserURL, `rwxr-xr-x@ 1 "github.com/viant/toolbox/storage/scp"  WORKGROUP\\Domain Users  108143621 Apr 19 15:55:57 2019 /Users/ojoseph/git/test-app/../go-cm/dist/linux/go-cm`, true)
		if assert.Nil(t, err) {
			var object = objects[0]
			assert.Equal(t, "scp://127.0.0.1:22/Users/ojoseph/git/test-app/../go-cm/dist/linux/go-cm", object.URL())
			assert.Equal(t, int64(1555689357), object.FileInfo().ModTime().Unix())
			assert.Equal(t, int64(108143621), object.FileInfo().Size())
			assert.Equal(t, true, object.IsContent())
		}

	}

}
