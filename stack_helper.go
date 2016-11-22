package toolbox

import (
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
