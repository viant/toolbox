package toolbox

import (
	"path"
	"runtime"
	"strings"
)

// CallerInfo return filename, function or file line from the stack
func CallerInfo(callerIndex int) (string, string, int) {
	var callerPointer = make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(callerIndex, callerPointer)
	callerInfo := runtime.FuncForPC(callerPointer[0])
	file, line := callerInfo.FileLine(callerPointer[0])
	callerName := callerInfo.Name()
	dotPosition := strings.LastIndex(callerName, ".")
	return file, callerName[dotPosition+1:], line
}

//CallerDirectory returns directory of caller source code directory
func CallerDirectory(callerIndex int) string {
	file, _, _ := CallerInfo(callerIndex)
	parent, _ := path.Split(file)
	return parent
}

func hasMatch(target string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.HasSuffix(target, candidate) {
			return true
		}
	}
	return false
}

//DiscoverCaller returns the first matched caller info
func DiscoverCaller(offset, maxDepth int, ignoreFiles ...string) (string, string, int) {
	var callerPointer = make([]uintptr, maxDepth) // at least 1 entry needed
	var caller *runtime.Func
	var filename string
	var line int
	for i := offset; i < maxDepth; i++ {
		runtime.Callers(i, callerPointer)
		caller = runtime.FuncForPC(callerPointer[0])
		filename, line = caller.FileLine(callerPointer[0])
		if hasMatch(filename, ignoreFiles...) {
			continue
		}
		break
	}
	callerName := caller.Name()
	dotPosition := strings.LastIndex(callerName, ".")
	return filename, callerName[dotPosition+1:], line
}
