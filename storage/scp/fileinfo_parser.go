package scp

import (
	"strings"
	"github.com/viant/toolbox/storage"
	"github.com/lunixbochs/vtclean"
	"unicode"
	"github.com/viant/toolbox"
	"fmt"
	"time"
)

const (
	fileInfoPermission = iota
	_
	fileInfoOwner
	fileInfoGroup
	fileInfoSize
	fileInfoDateMonth
	fileInfoDateDay
	fileInfoDateHour
	fileInfoDateYear
	fileInfoName
)

const (
	fileIsoInfoPermission = iota
	_
	fileIsoInfoOwner
	fileIsoInfoGroup
	fileIsoInfoSize
	fileIsoDate
	fileIsoTime
	fileIsoTimezone
	fileIsoInfoName
)

//Parser represents fileinfo parser from stdout
type Parser struct {
	IsoTimeStyle bool
}


func (p *Parser) Parse(URL, stdout string, isURLDir bool) ([]storage.Object, error) {
	var err error
	var result = make([]storage.Object, 0)
	if strings.Contains(stdout, "No such file or directory") {
		return result, nil
	}
	for _, line := range strings.Split(stdout, "\n") {
		line := strings.Replace(vtclean.Clean(line,false), "\r", "",1)
		var object storage.Object


		if p.IsoTimeStyle {
			object, err = p.extractObjectFromIsoBasedTimeCommand(URL, line, isURLDir)
		} else {
			object, err = p.extractObjectFromNonIsoBaseTimeCommand(URL, line, isURLDir)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, object)
	}
	return result, nil
}



func (p *Parser) HasNextTokenInout(nextTokenPosition int, line string) bool {
	if nextTokenPosition >= len(line) {
		return false
	}
	nextToken := []rune(string(line[nextTokenPosition:]))[0]
	return ! unicode.IsSpace(nextToken)
}

func (p *Parser) newObject(URL, name, permission, line, size string, modificationTime time.Time, isURLDirectory bool) (storage.Object, error) {
	fileMode, err := storage.NewFileMode(permission)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse line for lineinfo: %v, unable to file attributes: %v", line, err)
	}
	fileInfo := storage.NewFileInfo(name, int64(toolbox.AsInt(size)), fileMode, modificationTime, fileMode.IsDir())
	if isURLDirectory {
		URL = toolbox.URLPathJoin(URL, fileInfo.Name())
	}
	object := newStorageObject(URL, fileInfo, fileInfo)
	return object, nil
}


//extractObjectFromNonIsoBaseTimeCommand extract file storage object from line,
// it expects a file info without iso i.e  -rw-r--r--  1 awitas  1742120565   414 Jun  8 14:14:08 2017 id_rsa.pub
func (p *Parser) extractObjectFromNonIsoBaseTimeCommand(URL, line string, isURLDirectory bool) (storage.Object, error) {
	tokenIndex := 0
	if strings.TrimSpace(line) == "" {
		return nil, nil
	}
	var owner, name, permission, group, size, year, month, day, hour string
	for i, aRune := range line {
		if unicode.IsSpace(aRune) {
			if p.HasNextTokenInout(i+1, line) {
				tokenIndex++
			}
			continue
		}
		aChar := string(aRune)
		switch tokenIndex {
		case fileInfoPermission:
			permission += aChar
		case fileInfoOwner:
			owner += aChar
		case fileInfoGroup:
			group += aChar
		case fileInfoSize:
			size += aChar
		case fileInfoDateMonth:
			month += aChar
		case fileInfoDateDay:
			day += aChar
		case fileInfoDateHour:
			hour += aChar
		case fileInfoDateYear:
			year += aChar
		case fileInfoName:
			name += aChar
		}
	}


	if name == "" {
		return nil, fmt.Errorf("Failed to parse line for fileinfo: %v\n", line)
	}
	dateTime := year + " " + month + " " + day + " " + hour
	layout := toolbox.DateFormatToLayout("yyyy MMM ddd HH:mm:s")
	modificationTime, err := time.Parse(layout, dateTime)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse line for lineinfo: %v, unable to extract time: %v", line, err)
	}
	return p.newObject(URL, name, permission, line, size, modificationTime, isURLDirectory)
}


//extractObjectFromNonIsoBaseTimeCommand extract file storage object from line,
// it expects a file info with iso i.e. -rw-r--r-- 1 awitas awitas 2002 2017-11-04 22:29:33.363458941 +0000 aerospikeciads_aerospike.conf
func (p *Parser) extractObjectFromIsoBasedTimeCommand(URL, line string, isURLDirectory bool) (storage.Object, error) {
	tokeIndex := 0
	if strings.TrimSpace(line) == "" {
		return nil, nil
	}
	var owner, name, permission, group, timezone, date, modTime, size  string
	for i, aRune := range line {
		if unicode.IsSpace(aRune) {
			if p.HasNextTokenInout(i+1, line) {
				tokeIndex++
			}
			continue
		}
		aChar := string(aRune)
		switch tokeIndex {
		case fileIsoInfoPermission:
			permission += aChar
		case fileIsoInfoOwner:
			owner += aChar
		case fileIsoInfoGroup:
			group += aChar
		case fileIsoInfoSize:
			size += aChar
		case fileIsoDate:
			date += aChar
		case fileIsoTime:
			modTime += aChar
		case fileIsoTimezone:
			timezone += aChar
		case fileIsoInfoName:
			name += aChar
		}
		continue
	}

	timeLen := len(modTime)
	if timeLen > 12 {
		modTime = string(modTime[:12])
	}
	dateTime := date + " " + modTime + " " + timezone
	layout := toolbox.DateFormatToLayout("yyyy-MM-dd HH:mm:ss.SSS ZZ")
	modificationTime, err := time.Parse(layout, dateTime)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse line for lineinfo: %v, unable to extract time: %v", line, err)
	}
	return p.newObject(URL, name, permission, line, size, modificationTime, isURLDirectory)
}
