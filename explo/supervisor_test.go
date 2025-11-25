package explo_test

import (
	"context"
	"testing"
	"time"

	"github.com/morebec/mx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupervisor_Init(t *testing.T) {
	t.Run("GIVEN multiple application subsystems added, WHEN init THEN all of them should be init", func(t *testing.T) {
		initCounter := 0
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-1",
				initFunc: func() error {
					initCounter++
					return nil
				},
			}).
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-2",
				initFunc: func() error {
					initCounter++
					return nil
				},
			})

		err := supervisor.Init()
		require.NoError(t, err)
		assert.Equal(t, 2, initCounter)
	})

	t.Run("GIVEN an application subsystem that fails to init, WHEN init THEN error should be returned", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-1",
				initFunc: func() error {
					return assert.AnError
				},
			})

		err := supervisor.Init()
		require.Error(t, err)
	})
}

func TestSupervisor_Run(t *testing.T) {
	t.Run("GIVEN multiple application subsystems, WHEN run THEN all of them should be run", func(t *testing.T) {
		runCounter := 0
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-1",
				runFunc: func(ctx context.Context) error {
					runCounter++
					return nil
				},
			}).
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-2",
				runFunc: func(ctx context.Context) error {
					runCounter++
					return nil
				},
			})

		err := supervisor.Init()
		require.NoError(t, err)

		err = supervisor.Run(t.Context())
		require.NoError(t, err)

		assert.Equal(t, 2, runCounter)
	})

	t.Run("GIVEN an application subsystem that fails to run, WHEN run THEN error should be returned", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app",
				runFunc: func(ctx context.Context) error {
					return assert.AnError
				},
			}).
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "test-app-2",
				runFunc: func(ctx context.Context) error {
					return nil
				},
			})

		err := supervisor.Init()
		require.NoError(t, err)

		err = supervisor.Run(t.Context())
		require.Error(t, err)
	})

	t.Run("GIVEN supervisedApp run is idempotent while running, WHEN run called twice during execution THEN second call returns immediately", func(t *testing.T) {
		blockChan := make(chan struct{})
		runCounter := 0
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.WithApplicationSubsystem(&mockApplicationSubsystem{
			name: "test-app",
			runFunc: func(ctx context.Context) error {
				runCounter++
				<-blockChan // Block until we signal
				return nil
			},
		})

		err := supervisor.Init()
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		go supervisor.Run(ctx)
		time.Sleep(50 * time.Millisecond) // Let the first run start

		// The supervisor's Run just delegates to the supervised app, so calling Run again
		// will also try to run it, but the supervised app will return immediately if already running
		err = supervisor.Run(ctx)
		assert.NoError(t, err) // No error because apps that are already running return nil

		close(blockChan)
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 1, runCounter) // Still only ran once
	})
}

func TestSupervisor_Status(t *testing.T) {
	t.Run("GIVEN applications with different states, WHEN status called THEN return correct status for each", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "failing-app",
				runFunc: func(ctx context.Context) error {
					return assert.AnError
				},
			}).
			WithApplicationSubsystem(&mockApplicationSubsystem{
				name: "successful-app",
				runFunc: func(ctx context.Context) error {
					return nil
				},
			})

		err := supervisor.Init()
		require.NoError(t, err)

		err = supervisor.Run(t.Context())
		require.Error(t, err)

		statuses := supervisor.Status()
		require.Len(t, statuses, 2)

		failingStatus := findStatus(statuses, "failing-app")
		require.NotNil(t, failingStatus)
		assert.Equal(t, mx.ApplicationSubsystemStateFailed, failingStatus.State)
		assert.NotNil(t, failingStatus.Error)

		successStatus := findStatus(statuses, "successful-app")
		require.NotNil(t, successStatus)
		assert.Equal(t, mx.ApplicationSubsystemStateStopped, successStatus.State)
		assert.Nil(t, successStatus.Error)
	})
}

func TestSupervisor_Stop(t *testing.T) {
	t.Run("GIVEN a running application, WHEN stop called THEN application should stop", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		blockChan := make(chan struct{})
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.WithApplicationSubsystem(&mockApplicationSubsystem{
			name: "blocking-app",
			runFunc: func(appCtx context.Context) error {
				<-blockChan
				return nil
			},
		})

		err := supervisor.Init()
		require.NoError(t, err)

		go supervisor.Run(ctx)
		time.Sleep(100 * time.Millisecond)

		err = supervisor.Stop("blocking-app")
		require.NoError(t, err)

		close(blockChan)
		time.Sleep(100 * time.Millisecond)

		statuses := supervisor.Status()
		require.Len(t, statuses, 1)
		assert.Equal(t, mx.ApplicationSubsystemStateStopped, statuses[0].State)
	})

	t.Run("GIVEN stop is idempotent, WHEN stop called multiple times THEN should not error", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.WithApplicationSubsystem(&mockApplicationSubsystem{name: "test-app"})

		err := supervisor.Init()
		require.NoError(t, err)

		err = supervisor.Stop("test-app")
		require.NoError(t, err)

		err = supervisor.Stop("test-app")
		require.NoError(t, err)
	})

	t.Run("GIVEN application not found, WHEN stop called THEN error returned", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		err := supervisor.Stop("non-existent")
		require.Error(t, err)
	})
}

func TestSupervisor_Start(t *testing.T) {
	t.Run("GIVEN a stopped application, WHEN start called THEN application should start", func(t *testing.T) {
		runCounter := 0
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.WithApplicationSubsystem(&mockApplicationSubsystem{
			name: "test-app",
			runFunc: func(ctx context.Context) error {
				runCounter++
				<-ctx.Done()
				return nil
			},
		})

		err := supervisor.Init()
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		err = supervisor.Start("test-app", ctx)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		statuses := supervisor.Status()
		require.Len(t, statuses, 1)
		assert.Equal(t, mx.ApplicationSubsystemStateRunning, statuses[0].State)
	})

	t.Run("GIVEN start is idempotent, WHEN start called multiple times THEN should not error", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		supervisor.WithApplicationSubsystem(&mockApplicationSubsystem{
			name: "test-app",
			runFunc: func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
		})

		err := supervisor.Init()
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		err = supervisor.Start("test-app", ctx)
		require.NoError(t, err)

		err = supervisor.Start("test-app", ctx)
		require.NoError(t, err)
	})

	t.Run("GIVEN application not found, WHEN start called THEN error returned", func(t *testing.T) {
		supervisor := &mx.SupervisorApplicationSubsystem{}
		err := supervisor.Start("non-existent", t.Context())
		require.Error(t, err)
	})
}

// Helper function to find a status by name
func findStatus(statuses []mx.ApplicationSubsystemStatus, name string) *mx.ApplicationSubsystemStatus {
	for i := range statuses {
		if statuses[i].Name == name {
			return &statuses[i]
		}
	}
	return nil
}
