// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package storetest

import (
	"context"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"sync"
)

// Ensure, that StorerMock does implement store.Storer.
// If this is not the case, regenerate this file with moq.
var _ store.Storer = &StorerMock{}

// StorerMock is a mock implementation of store.Storer.
//
//	func TestSomethingThatUsesStorer(t *testing.T) {
//
//		// make and configure a mocked store.Storer
//		mockedStorer := &StorerMock{
//			CheckAllBundleContentsAreApprovedFunc: func(ctx context.Context, bundleID string) (bool, error) {
//				panic("mock out the CheckAllBundleContentsAreApproved method")
//			},
//			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			CloseFunc: func(ctx context.Context) error {
//				panic("mock out the Close method")
//			},
//			ListBundlesFunc: func(ctx context.Context, offset int, limit int) ([]*models.Bundle, int, error) {
//				panic("mock out the ListBundles method")
//			},
//		}
//
//		// use mockedStorer in code that requires store.Storer
//		// and then make assertions.
//
//	}
type StorerMock struct {
	// CheckAllBundleContentsAreApprovedFunc mocks the CheckAllBundleContentsAreApproved method.
	CheckAllBundleContentsAreApprovedFunc func(ctx context.Context, bundleID string) (bool, error)

	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// ListBundlesFunc mocks the ListBundles method.
	ListBundlesFunc func(ctx context.Context, offset int, limit int) ([]*models.Bundle, int, error)

	// calls tracks calls to the methods.
	calls struct {
		// CheckAllBundleContentsAreApproved holds details about calls to the CheckAllBundleContentsAreApproved method.
		CheckAllBundleContentsAreApproved []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// BundleID is the bundleID argument value.
			BundleID string
		}
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// Close holds details about calls to the Close method.
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// ListBundles holds details about calls to the ListBundles method.
		ListBundles []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Offset is the offset argument value.
			Offset int
			// Limit is the limit argument value.
			Limit int
		}
	}
	lockCheckAllBundleContentsAreApproved sync.RWMutex
	lockChecker                           sync.RWMutex
	lockClose                             sync.RWMutex
	lockListBundles                       sync.RWMutex
}

// CheckAllBundleContentsAreApproved calls CheckAllBundleContentsAreApprovedFunc.
func (mock *StorerMock) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	if mock.CheckAllBundleContentsAreApprovedFunc == nil {
		panic("StorerMock.CheckAllBundleContentsAreApprovedFunc: method is nil but Storer.CheckAllBundleContentsAreApproved was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		BundleID string
	}{
		Ctx:      ctx,
		BundleID: bundleID,
	}
	mock.lockCheckAllBundleContentsAreApproved.Lock()
	mock.calls.CheckAllBundleContentsAreApproved = append(mock.calls.CheckAllBundleContentsAreApproved, callInfo)
	mock.lockCheckAllBundleContentsAreApproved.Unlock()
	return mock.CheckAllBundleContentsAreApprovedFunc(ctx, bundleID)
}

// CheckAllBundleContentsAreApprovedCalls gets all the calls that were made to CheckAllBundleContentsAreApproved.
// Check the length with:
//
//	len(mockedStorer.CheckAllBundleContentsAreApprovedCalls())
func (mock *StorerMock) CheckAllBundleContentsAreApprovedCalls() []struct {
	Ctx      context.Context
	BundleID string
} {
	var calls []struct {
		Ctx      context.Context
		BundleID string
	}
	mock.lockCheckAllBundleContentsAreApproved.RLock()
	calls = mock.calls.CheckAllBundleContentsAreApproved
	mock.lockCheckAllBundleContentsAreApproved.RUnlock()
	return calls
}

// Checker calls CheckerFunc.
func (mock *StorerMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("StorerMock.CheckerFunc: method is nil but Storer.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//
//	len(mockedStorer.CheckerCalls())
func (mock *StorerMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *StorerMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("StorerMock.CloseFunc: method is nil but Storer.Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//
//	len(mockedStorer.CloseCalls())
func (mock *StorerMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// ListBundles calls ListBundlesFunc.
func (mock *StorerMock) ListBundles(ctx context.Context, offset int, limit int) ([]*models.Bundle, int, error) {
	if mock.ListBundlesFunc == nil {
		panic("StorerMock.ListBundlesFunc: method is nil but Storer.ListBundles was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Offset int
		Limit  int
	}{
		Ctx:    ctx,
		Offset: offset,
		Limit:  limit,
	}
	mock.lockListBundles.Lock()
	mock.calls.ListBundles = append(mock.calls.ListBundles, callInfo)
	mock.lockListBundles.Unlock()
	return mock.ListBundlesFunc(ctx, offset, limit)
}

// ListBundlesCalls gets all the calls that were made to ListBundles.
// Check the length with:
//
//	len(mockedStorer.ListBundlesCalls())
func (mock *StorerMock) ListBundlesCalls() []struct {
	Ctx    context.Context
	Offset int
	Limit  int
} {
	var calls []struct {
		Ctx    context.Context
		Offset int
		Limit  int
	}
	mock.lockListBundles.RLock()
	calls = mock.calls.ListBundles
	mock.lockListBundles.RUnlock()
	return calls
}
