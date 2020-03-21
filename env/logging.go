package env

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	if IsLocal() {
		log.SetFormatter(&log.TextFormatter{
			ForceColors: true,
		})
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetFormatter(new(stackdriverFormatter))
	}
}

type stackdriverFormatter struct {
	log.JSONFormatter
}

func (f *stackdriverFormatter) Format(entry *log.Entry) ([]byte, error) {
	switch entry.Level {
	// EMERGENCY (800) One or more systems are unusable.
	case log.PanicLevel:
		entry.Data["severity"] = 800

	// ALERT (700) A person must take an action immediately.
	case log.FatalLevel:
		entry.Data["severity"] = 700

	// CRITICAL (600) Critical events cause more severe problems or outages.
	// No equivalent in logrus.

	// ERROR (500) Error events are likely to cause problems.
	case log.ErrorLevel:
		entry.Data["severity"] = 500

	// WARNING (400) Warning events might cause problems.
	case log.WarnLevel:
		entry.Data["severity"] = 400

	// NOTICE (300) Normal but significant events, such as start up, shut down, or a configuration change.
	// No equivalent in logrus.

	// INFO (200) Routine information, such as ongoing status or performance.
	case log.InfoLevel:
		entry.Data["severity"] = 200

	// DEBUG (100) Debug or trace information.
	case log.DebugLevel:
		entry.Data["severity"] = 100

	// DEFAULT (0) The log entry has no assigned severity level.
	case log.TraceLevel:
		entry.Data["severity"] = 0
	}

	return f.JSONFormatter.Format(entry)
}
