package service

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/joshakeman/gametime-take-home/data"
)

func TestProcess(t *testing.T) {
	tests := []struct {
		name        string
		payment     FakePaymentService
		completer   FakeCompleter
		wantStatus  data.Status
		wantHistory []data.Status
		wantErrs    []error
	}{
		{
			name:       "happy path",
			wantStatus: data.StatusOrderCompleted,
			wantHistory: []data.Status{
				data.StatusOrderPending, data.StatusOrderProcessing,
				data.StatusPaymentAuthorized, data.StatusOrderCompleted,
			},
		},
		{
			name:       "payment declined",
			payment:    FakePaymentService{ShouldDecline: true},
			wantStatus: data.StatusPaymentDeclined,
			wantHistory: []data.Status{
				data.StatusOrderPending, data.StatusOrderProcessing,
				data.StatusPaymentDeclined,
			},
			wantErrs: []error{ErrPaymentDeclined},
		},
		{
			name:       "completion fails, void succeeds",
			completer:  FakeCompleter{ShouldFailComplete: true},
			wantStatus: data.StatusOrderVoided,
			wantHistory: []data.Status{
				data.StatusOrderPending, data.StatusOrderProcessing,
				data.StatusPaymentAuthorized, data.StatusOrderVoided,
			},
			wantErrs: []error{ErrCompleteFailed},
		},
		{
			name:       "completion fails, void fails",
			payment:    FakePaymentService{ShouldFailVoid: true},
			completer:  FakeCompleter{ShouldFailComplete: true},
			wantStatus: data.StatusNeedsAttention,
			wantHistory: []data.Status{
				data.StatusOrderPending, data.StatusOrderProcessing,
				data.StatusPaymentAuthorized, data.StatusNeedsAttention,
			},
			wantErrs: []error{ErrCompleteFailed, ErrVoidFailed},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(data.NewStore(), tt.payment, tt.completer)
			created, err := svc.Create("user-1", 100)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			o, err := svc.Process(created.ID)

			if len(tt.wantErrs) == 0 && err != nil {
				t.Errorf("Process error = %v, want nil", err)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("Process error = %v, want it to include %v", err, want)
				}
			}
			if o.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", o.Status, tt.wantStatus)
			}

			var got []data.Status
			for i, e := range o.History {
				got = append(got, e.Status)
				if e.At.IsZero() {
					t.Errorf("History[%d] has no timestamp", i)
				}
				if i > 0 && e.At.Before(o.History[i-1].At) {
					t.Errorf("History[%d] is earlier than the entry before it", i)
				}
			}
			if !slices.Equal(got, tt.wantHistory) {
				t.Errorf("History = %v, want %v", got, tt.wantHistory)
			}

			// A finished order can't be processed again.
			if _, err := svc.Process(created.ID); !errors.Is(err, data.ErrInvalidTransition) {
				t.Errorf("second Process error = %v, want %v", err, data.ErrInvalidTransition)
			}
		})
	}
}

func TestProcessUnknownOrder(t *testing.T) {
	svc := New(data.NewStore(), FakePaymentService{}, FakeCompleter{})

	if _, err := svc.Process(999); !errors.Is(err, data.ErrNotFound) {
		t.Errorf("Process error = %v, want %v", err, data.ErrNotFound)
	}
}

// Concurrent requests for the same order must only process it once.
func TestProcessConcurrent(t *testing.T) {
	svc := New(data.NewStore(), FakePaymentService{}, FakeCompleter{})
	o, err := svc.Create("user-1", 100)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var (
		mu        sync.Mutex
		successes int
		wg        sync.WaitGroup
		start     = make(chan struct{})
	)
	for range 20 {
		wg.Go(func() {
			<-start
			if _, err := svc.Process(o.ID); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		})
	}
	close(start)
	wg.Wait()

	if successes != 1 {
		t.Errorf("successful Process calls = %d, want 1", successes)
	}
}
