package explo_test

import (
	"context"
	"testing"

	"github.com/morebec/mx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingApplicationSubsystem_Init(t *testing.T) {
	const applicationName = "some-application"

	type logEntry struct {
		Msg       string `json:"msg"`
		Level     string `json:"level"`
		Subsystem string `json:"subsystem"`
	}

	type expect struct {
		err         require.ErrorAssertionFunc
		logMessages []logEntry
	}
	tests := []struct {
		name           string
		givenInitError error
		then           expect
	}{
		{
			name:           "GIVEN Init succeeds THEN logs success messages and returns no error",
			givenInitError: nil,
			then: expect{
				err: require.NoError,
				logMessages: []logEntry{
					{Msg: "initializing application subsystem...", Level: "INFO", Subsystem: applicationName},
					{Msg: "initialized application subsystem successfully", Level: "INFO", Subsystem: applicationName},
				},
			},
		},
		{
			name:           "GIVEN Init fails THEN logs error messages and returns error",
			givenInitError: assert.AnError,
			then: expect{
				err: require.Error,
				logMessages: []logEntry{
					{Msg: "initializing application subsystem...", Level: "INFO", Subsystem: applicationName},
					{Msg: "failed to initialize application subsystem: " + assert.AnError.Error(), Level: "ERROR", Subsystem: applicationName},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logWriter := newTestLogWriter[logEntry]()
			logger := newTestLogger(logWriter)
			s := mx.NewLoggingApplicationSubsystem(
				mockApplicationSubsystem{
					name:     applicationName,
					initFunc: func() error { return tt.givenInitError },
				},
				logger,
			)

			err := s.Initialize()
			tt.then.err(t, err)
			logWriter.assertLogEntries(t, tt.then.logMessages)
		})
	}
}

func TestLoggingApplicationSubsystem_Run(t *testing.T) {
	const applicationName = "some-application"

	type logEntry struct {
		Msg       string `json:"msg"`
		Level     string `json:"level"`
		Subsystem string `json:"subsystem"`
	}

	type expect struct {
		err         require.ErrorAssertionFunc
		logMessages []logEntry
	}
	tests := []struct {
		name          string
		givenRunError error
		then          expect
	}{
		{
			name:          "GIVEN Run succeeds THEN logs success messages and returns no error",
			givenRunError: nil,
			then: expect{
				err: require.NoError,
				logMessages: []logEntry{
					{Msg: "running application subsystem...", Level: "INFO", Subsystem: applicationName},
					{Msg: "application subsystem run completed", Level: "INFO", Subsystem: applicationName},
				},
			},
		},
		{
			name:          "GIVEN Run fails THEN logs error messages and returns error",
			givenRunError: assert.AnError,
			then: expect{
				err: require.Error,
				logMessages: []logEntry{
					{Msg: "running application subsystem...", Level: "INFO", Subsystem: applicationName},
					{Msg: "application subsystem failed: " + assert.AnError.Error(), Level: "ERROR", Subsystem: applicationName},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logWriter := newTestLogWriter[logEntry]()
			logger := newTestLogger(logWriter)
			s := mx.NewLoggingApplicationSubsystem(
				mockApplicationSubsystem{
					name:    applicationName,
					runFunc: func(ctx context.Context) error { return tt.givenRunError },
				},
				logger,
			)

			err := s.Run(t.Context())
			tt.then.err(t, err)
			logWriter.assertLogEntries(t, tt.then.logMessages)
		})
	}
}

type mockApplicationSubsystem struct {
	name     string
	initFunc func() error
	runFunc  func(context.Context) error
}

func (m mockApplicationSubsystem) Name() string {
	return m.name
}

func (m mockApplicationSubsystem) Initialize() error {
	if m.initFunc == nil {
		return nil
	}
	return m.initFunc()
}

func (m mockApplicationSubsystem) Teardown(context.Context) error {
	return nil
}

func (m mockApplicationSubsystem) Run(ctx context.Context) error {
	if m.runFunc == nil {
		return nil
	}
	return m.runFunc(ctx)
}
