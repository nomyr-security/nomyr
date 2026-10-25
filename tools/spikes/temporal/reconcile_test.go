package temporal

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type Request struct {
	OperationID string `json:"operation_id"`
}

type Result struct {
	Applied bool `json:"applied"`
}

func ReconcileBeforeRetry(ctx workflow.Context, request Request) (Result, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	var result Result
	if err := workflow.ExecuteActivity(ctx, "reconcile", request).Get(ctx, &result); err != nil {
		return Result{}, err
	}
	if result.Applied {
		return result, nil
	}
	if err := workflow.ExecuteActivity(ctx, "apply", request).Get(ctx, &result); err == nil {
		return result, nil
	}
	if err := workflow.ExecuteActivity(ctx, "reconcile", request).Get(ctx, &result); err != nil {
		return Result{}, err
	}
	if result.Applied {
		return result, nil
	}
	retryCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	err := workflow.ExecuteActivity(retryCtx, "apply", request).Get(retryCtx, &result)
	return result, err
}

func TestReconcileRunsBeforeRetryAndPayloadIsSecretFree(t *testing.T) {
	canary := fmt.Sprintf("runtime-canary-%s", strings.Repeat("c", 24))
	request := Request{OperationID: "synthetic-operation-1"}
	encoded, err := converter.GetDefaultDataConverter().ToPayload(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := rejectCanary(encoded.Data, canary); err != nil {
		t.Fatal("workflow request contains the generated canary")
	}

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	var calls []string
	environment.RegisterActivityWithOptions(func(Request) (Result, error) {
		calls = append(calls, "reconcile")
		return Result{}, nil
	}, activity.RegisterOptions{Name: "reconcile"})
	environment.RegisterActivityWithOptions(func(Request) (Result, error) {
		calls = append(calls, "apply")
		if len(calls) == 2 {
			return Result{}, temporal.NewTimeoutError(enumspb.TIMEOUT_TYPE_START_TO_CLOSE, nil)
		}
		return Result{Applied: true}, nil
	}, activity.RegisterOptions{Name: "apply"})
	environment.ExecuteWorkflow(ReconcileBeforeRetry, request)
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if want := []string{"reconcile", "apply", "reconcile", "apply"}; fmt.Sprint(calls) != fmt.Sprint(want) {
		t.Fatalf("activity sequence = %v, want %v", calls, want)
	}
}

func TestPayloadGuardRejectsRuntimeCanary(t *testing.T) {
	canary := fmt.Sprintf("runtime-canary-%s", strings.Repeat("x", 32))
	unsafePayload, err := converter.GetDefaultDataConverter().ToPayload(map[string]string{"diagnostic": canary})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unsafePayload.Data), canary) {
		t.Fatal("negative fixture did not contain its runtime canary")
	}
	if err := rejectCanary(unsafePayload.Data, canary); err == nil {
		t.Fatal("secret-bearing payload was accepted")
	}
}

func rejectCanary(payload []byte, canary string) error {
	if strings.Contains(string(payload), canary) {
		return errors.New("serialized payload contains runtime canary")
	}
	return nil
}
