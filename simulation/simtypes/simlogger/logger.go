// This package is a hack. It is meant to solve temporary logging pains in simulation.
// Every issue this package fixes needs a longer term fix in our stack, to not get into full node logs.
package simlogger

import (
	"strings"

	"github.com/tendermint/tendermint/libs/log"
)

type simLogger struct {
	logger log.Logger
}

func (s *simLogger) Debug(msg string, keyvals ...interface{}) {
	// Suppress this log
	if strings.Contains(msg, "committed KVStore") {
		return
	}
	s.logger.Debug(msg, keyvals...)
}

func (s *simLogger) Info(msg string, keyvals ...interface{}) {
	s.logger.Info(msg, keyvals)
}

func (s *simLogger) Error(msg string, keyvals ...interface{}) {
	s.logger.Error(msg, keyvals)
}

func (s *simLogger) With(keyvals ...interface{}) log.Logger {
	return s.logger.With(keyvals...)
}

func NewSimLogger(logger log.Logger) log.Logger {
	return &simLogger{logger}
}
