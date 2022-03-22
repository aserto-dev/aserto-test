package tester

import (
	"fmt"
	"strings"
	"time"

	"github.com/aserto-dev/aserto-go/client/authorizer"
	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/x"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"

	"github.com/pkg/errors"
)

type Manager struct {
	policyID     string
	client       *authorizer.Client
	policyClient policy.PolicyClient
	authzClient  authz.AuthorizerClient
}

type TestRule struct {
	Package string `json:"package"`
	Name    string `json:"name"`
}

func (tr *TestRule) String() string {
	return fmt.Sprintf("%v.%v", tr.Package, tr.Name)
}

type TestResult struct {
	Package  string        `json:"package"`
	Name     string        `json:"name"`
	Fail     bool          `json:"fail"`
	Error    error         `json:"error"`
	Skip     bool          `json:"skip"`
	Duration time.Duration `json:"duration"`
	Output   []string      `json:"output"`
}

// Pass returns true if the test case passed.
func (tr *TestResult) Pass() bool {
	return !tr.Fail && !tr.Skip && tr.Error == nil
}

func (tr *TestResult) String() string {
	if tr.Skip {
		return fmt.Sprintf("%v.%v: %v", tr.Package, tr.Name, tr.outcome())
	}
	return fmt.Sprintf("%v.%v: %v (%v)", tr.Package, tr.Name, tr.outcome(), tr.Duration)
}

func (tr *TestResult) outcome() string {
	if tr.Pass() {
		return "PASS"
	}
	if tr.Fail {
		return "FAIL"
	}
	if tr.Skip {
		return "SKIPPED"
	}
	return "ERROR"
}

// NewManager returns new test manager instance.
func NewManager(policyID string, client *authorizer.Client) *Manager {
	return &Manager{
		policyID:     policyID,
		client:       client,
		policyClient: client.Policy,
		authzClient:  client.Authorizer,
	}
}

func (tm *Manager) Ping(c *cc.CommonCtx) bool {
	const guid string = "162d2672-aa9d-4d61-b91d-9b277cbfcdb3"

	queryResult, err := tm.authzClient.Query(c.Context, &authz.QueryRequest{
		Query: fmt.Sprintf("%s = input.guid", resultVar),
		Input: fmt.Sprintf("{%q:%q}", "guid", guid),
		PolicyContext: &api.PolicyContext{
			Id: tm.policyID,
		},
		IdentityContext: &api.IdentityContext{
			Type: api.IdentityType_IDENTITY_TYPE_NONE,
		},
		ResourceContext: &structpb.Struct{},
		Options:         &authz.QueryOptions{},
	})
	if err == nil {
		for _, r := range queryResult.Results {
			if v, ok := r.Fields[resultVar]; ok && isString(v) {
				return v.GetStringValue() == guid
			}
		}
	}
	return false
}

func (tm *Manager) Enum(c *cc.CommonCtx) ([]string, error) {
	result := []string{}

	policiesList, err := tm.policyClient.ListPolicies(c.Context, &policy.ListPoliciesRequest{})
	if err != nil {
		return result, err
	}

	for _, policyItem := range policiesList.Results {
		result = append(result, policyItem.Id)
	}

	return result, nil
}

// List tests for a given policy ID.
func (tm *Manager) List(c *cc.CommonCtx) ([]*TestRule, error) {
	result := []*TestRule{}

	policiesList, err := tm.policyClient.ListPolicies(c.Context, &policy.ListPoliciesRequest{})
	if err != nil {
		return result, err
	}

	exists := false
	for _, policyItem := range policiesList.Results {
		if policyItem.Id == tm.policyID {
			exists = true
			break
		}
	}

	if !exists {
		return result, errors.Errorf("policy id [%s] not found", tm.policyID)
	}

	policies, err := tm.policyClient.GetPolicies(c.Context, &policy.GetPoliciesRequest{Id: tm.policyID})
	if err != nil {
		return result, err
	}

	for _, policyItem := range policies.Policies {
		module, err := tm.policyClient.GetModule(c.Context, &policy.GetModuleRequest{PolicyId: tm.policyID, Id: policyItem.Id})
		if err != nil {
			return []*TestRule{}, err
		}

		for _, rule := range module.Module.Rules {
			if !strings.HasPrefix(rule, x.TestPrefix) && !strings.HasPrefix(rule, x.SkipTestPrefix) {
				continue
			}
			result = append(result, &TestRule{
				Package: "data." + module.Module.Name,
				Name:    rule,
			})
		}
	}

	return result, nil
}

// Run tests for a given policy ID.
func (tm *Manager) Run(c *cc.CommonCtx) ([]*TestResult, error) {
	result := []*TestResult{}

	testList, err := tm.List(c)
	if err != nil {
		return result, err
	}

	for _, testItem := range testList {
		// skip when todo_test_
		if strings.HasPrefix(testItem.Name, x.SkipTestPrefix) {
			result = append(result, &TestResult{
				Package: testItem.Package,
				Name:    testItem.Name,
				Skip:    true,
			})
			continue
		}

		tr, err := tm.execTest(c, testItem)
		if err != nil {
			return []*TestResult{}, err
		}

		result = append(result, tr)
	}

	return result, nil
}

const (
	resultVar            string = "x"
	metricsQueryEvalinNS string = "timer_rego_query_eval_ns"
)

// Execute test.
func (tm *Manager) execTest(c *cc.CommonCtx, t *TestRule) (*TestResult, error) {

	queryResult, err := tm.authzClient.Query(c.Context, &authz.QueryRequest{
		PolicyContext: &api.PolicyContext{
			Id: tm.policyID,
		},
		Query: fmt.Sprintf("%s = %s.%s", resultVar, t.Package, t.Name),
		Input: "",
		IdentityContext: &api.IdentityContext{
			Type: api.IdentityType_IDENTITY_TYPE_NONE,
		},
		ResourceContext: &structpb.Struct{},
		Options: &authz.QueryOptions{
			Metrics:      true,
			Instrument:   false,
			Trace:        authz.TraceLevel_TRACE_LEVEL_FAILS,
			TraceSummary: true,
		},
	})

	fail := true

	if err == nil {
		for _, r := range queryResult.Results {
			if v, ok := r.Fields[resultVar]; ok && isBool(v) {
				fail = !v.GetBoolValue()
			}
		}
	}

	queryEvalNS, ok := queryResult.Metrics.Fields[metricsQueryEvalinNS]

	duration := time.Duration(0)
	if ok {
		duration = time.Duration(queryEvalNS.GetNumberValue())
	}

	result := &TestResult{
		Package:  t.Package,                  // package name in which test is defined
		Name:     t.Name,                     // test name
		Fail:     fail,                       // outcome of x = data....
		Error:    err,                        // error of query execution
		Skip:     false,                      //
		Duration: time.Nanosecond * duration, // duration of query evaluation in nanoseconds
		Output:   queryResult.TraceSummary,   // output
	}

	return result, err
}

func isBool(v *structpb.Value) bool {
	_, ok := v.GetKind().(*structpb.Value_BoolValue)
	return ok
}

func isString(v *structpb.Value) bool {
	_, ok := v.GetKind().(*structpb.Value_StringValue)
	return ok
}
