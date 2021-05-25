package service_test

import (
	"context"
	"testing"

	"github.com/yarbelk/distributedservice/data"
	"github.com/yarbelk/distributedservice/proto"
	"github.com/yarbelk/distributedservice/service"
)

// MockStorer because we are not trying to test the Storer; do that in that package.
// we want to control what the storer returns so we can test the service and how
// it interacts with the supporting libraries/services like memberlist and
// consistent hasing
type MockStorer struct {
	customerStateCalled, writeLogCalled bool

	customerState      data.CustomerState
	customerStateError error
	log                *proto.CustomerEventLog
}

func (m *MockStorer) GetCustomerState(id uint64) (data.CustomerState, error) {
	m.customerStateCalled = true
	return m.customerState, m.customerStateError
}

func (m *MockStorer) WriteLog(id uint64, el *proto.CustomerEventLog) error {
	m.writeLogCalled = true
	m.log = el
	return nil
}

// These are testing stubs: basically to show _some_ of what I'm thinking
// I put them together from quick notes i made in-lieu of formal TDD due to
// time constraints

func TestSimpleLookup(t *testing.T) {
	c := service.Customer{Storage: &MockStorer{customerState: data.CustomerState{LastAction: "fake", CurrentSequence: 0}}}

	var tests = []struct {
		name     string
		expected *proto.CustomerState
		given    *proto.Customer
	}{
		{"Get exiting customer", &proto.CustomerState{Id: 1, LastAction: "fake"}, &proto.Customer{Id: 1}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.CustomerState(context.Background(), tt.given)
			if err != nil {
				t.Fatalf("don't want error here %s", err)
			}
			if actual.Id != tt.expected.Id || actual.LastAction != tt.expected.LastAction {
				t.Errorf("given(%s): expected %s, actual %s", tt.given, tt.expected, actual)
			}
		})
	}
}

func TestHashFilterServer(t *testing.T) {
	// SEtup the mock storage
	// setup a memberlist with mock transport etc (hashicorp is awesome and gives you all this stuff for testing)

	t.Run("Test filters on write to correct node", func(t *testing.T) {
		// blah
	})
	t.Run("Test filters get right error on wrong node called", func(t *testing.T) {
		// blah
	})
	t.Run("Test filters get right error on wrong node called", func(t *testing.T) {
		// blah
	})
}

func TestHashFilterClient(t *testing.T) {
	// Mock the service responses as per above.
	// you can also run them using testing.M and manage it as an integration test locally since
	// memberlist and the peices all support that
	// setup a memberlist with mock transport etc (hashicorp is awesome and gives you all this stuff for testing)

	t.Run("Test filters set metadata for routing", func(t *testing.T) {
		// blah
	})
	t.Run("Test filters client calls right host", func(t *testing.T) {
		// blah
	})
	t.Run("Test filters cliet gets sane user friendly errors", func(t *testing.T) {
		// blah
	})
}
