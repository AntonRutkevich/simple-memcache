package resp

import (
	"fmt"
	"github.com/antonrutkevich/simple-memcache/pkg/domain/core"
)

type commandNotSupported struct {
	command string
}

func errCommandNotSupported(command string) *commandNotSupported {
	return &commandNotSupported{command: command}
}

func (e *commandNotSupported) Error() string {
	return fmt.Sprintf("%s not supported", e.command)
}

func (e *commandNotSupported) ClientErrorCode() core.ClientErrCode {
	return "NOT_SUPPORTED"
}

type protocolError struct {
	message string
}

func errProtocolError(message string, args ...interface{}) *protocolError {
	return &protocolError{message: fmt.Sprintf(message, args...)}
}

func (e *protocolError) Error() string {
	return fmt.Sprintf("%s not supported", e.message)
}

func (e *protocolError) ClientErrorCode() core.ClientErrCode {
	return "PROTOCOL_ERROR"
}
