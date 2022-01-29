package logruzio

import (
	"encoding/json"
	"github.com/logzio/logzio-go"
	"github.com/sirupsen/logrus"
	"strings"
)

// HookOpts represents Logrus Logzio hook options
type HookOpts struct {
	Sender           *logzio.LogzioSender
	AdditionalFields logrus.Fields
}

// Hook represents a Logrus Logzio hook
type Hook struct {
	hookOpts *HookOpts
}

// New creates a default Logzio hook.
// What it does is taking `token` and `appName` and attaching them to the log data.
// In addition, it sets a connection to the Logzio's Logstash endpoint.
// If the connection fails, it returns an error.
//
// To set more advanced configurations, initialize the hook in the following way:
//
// hook := &Hook{&HookOpts{
//		Sender: logzioSender,
//		AdditionalFields: logrus.Fields{...},
// }
func New(host, token, appName string, fields logrus.Fields) (*Hook, error) {
	opts := &HookOpts{AdditionalFields: fields}
	opts.AdditionalFields["type"] = appName
	l, err := logzio.New(
		token,
		logzio.SetDebug(nil),
		logzio.SetUrl(host),
		logzio.SetInMemoryQueue(true),
		logzio.SetinMemoryCapacity(1024*1024),
	)
	if err != nil {
		return nil, err
	}
	opts.Sender = l
	return &Hook{opts}, nil
}

// Fire writes `entry` to Logzio
func (h *Hook) Fire(entry *logrus.Entry) error {
	r := map[string]interface{}{
		"message":  entry.Message,
		"severity": strings.ToUpper(entry.Level.String()),
	}
	// Add in context fields.
	for k, v := range h.hookOpts.AdditionalFields {
		// We don't override fields that are already set
		if _, ok := entry.Data[k]; !ok {
			r[k] = v
		}
	}

	dataBytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	err = h.hookOpts.Sender.Send(dataBytes)
	if err != nil {
		return err
	}

	return nil
}

// Levels returns logging levels
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
