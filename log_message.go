package toolbox

//LogMessage represent log message
type LogMessage struct {
	MessageType string
	Message     interface{}
}

//LogMessages represents log messages
type LogMessages struct {
	Messages []LogMessage
}
