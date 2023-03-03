package log

import (
	"fmt"
	"io"
)

// The DefaultLogger does not output any information on standard output.
//
// It only shows error output on standard error.
type DefaultLogger struct {
	ErrorOutput io.Writer
}

func (l *DefaultLogger) LogRequest(request RequestInfo) {
}

func (l DefaultLogger) LogResponse(response ResponseInfo) {
}

func (l DefaultLogger) LogDebug(message string) {
}

func (l DefaultLogger) LogError(message string) {
	fmt.Fprint(l.ErrorOutput, message)
}
