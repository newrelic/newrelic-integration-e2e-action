package logger

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

// CommandLogger is an object to log output from commands
type CommandLogger interface {
	// Open returns an io.Writer the command should write its stdout and stderr to.
	// Implementations may assume callers do not interact with the underlying writers between calls to Open and Close.
	Open(name string) io.Writer
	// Close signals implementations that no more data will be written.
	// Implementations may write additional trailers when Close is called.
	Close()
}

// LogrusLogger is an implementation of CommandLogger that uses a logrus logger.
type LogrusLogger struct {
	logrus *logrus.Logger
	pipe   *io.PipeWriter
}

func NewLogrusLogger(logger *logrus.Logger) LogrusLogger {
	return LogrusLogger{logrus: logger}
}

func (ll LogrusLogger) Open(name string) io.Writer {
	ll.pipe = ll.logrus.WithField("command", name).Writer()
	return ll.pipe
}

func (ll LogrusLogger) Close() {
	if ll.pipe != nil {
		_ = ll.pipe.Close()
	}
}

// GHALogger implements CommandLogger on top of an io.Writer.
// GHALogger will wrap command output in groups using the GHA syntax:
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#grouping-log-lines
type GHALogger struct {
	io.Writer
}

func NewGHALogger(writer io.Writer) GHALogger {
	return GHALogger{writer}
}

func (gha GHALogger) Open(name string) io.Writer {
	_, _ = fmt.Fprintf(gha, "::group::%s\n", name)
	return gha
}

func (gha GHALogger) Close() {
	_, _ = fmt.Fprintln(gha, "::endgroup::")
}
