package tui

import (
	"reflect"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

func TestViewStateInterface(_ *testing.T) {
	// Verify all states implement ViewState
	var (
		_ ViewState = (*ListViewState)(nil)
		_ ViewState = (*DetailViewState)(nil)
		_ ViewState = (*ConfirmViewState)(nil)
	)
}

func TestListViewState_IsZeroValue(_ *testing.T) {
	// ListViewState is a zero-value struct with no fields.
	// Verify it can be created and used as a ViewState.
	state := &ListViewState{}

	// Verify it implements ViewState by assigning to the interface type.
	// This is a compile-time check; if it fails, the code won't compile.
	var _ ViewState = state
}

func TestDetailViewState_Loading(t *testing.T) {
	tests := []struct {
		name    string
		loading bool
	}{
		{name: "loading", loading: true},
		{name: "loaded", loading: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &DetailViewState{Loading: tt.loading}
			if state.Loading != tt.loading {
				t.Errorf("DetailViewState.Loading = %v, want %v", state.Loading, tt.loading)
			}
		})
	}
}

func TestConfirmViewState_Fields(t *testing.T) {
	tests := []struct {
		name      string
		action    components.ConfirmAction
		targetIDs []string
	}{
		{name: "close action", action: components.ActionClose, targetIDs: []string{"ws-1"}},
		{name: "push action", action: components.ActionPush, targetIDs: []string{"ws-2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &ConfirmViewState{
				Action:    tt.action,
				TargetIDs: tt.targetIDs,
			}

			if state.Action != tt.action {
				t.Errorf("ConfirmViewState.Action = %s, want %s", state.Action, tt.action)
			}

			if !reflect.DeepEqual(state.TargetIDs, tt.targetIDs) {
				t.Errorf("ConfirmViewState.TargetIDs = %v, want %v", state.TargetIDs, tt.targetIDs)
			}
		})
	}
}

func TestModel_IsDetailView(t *testing.T) {
	tests := []struct {
		name      string
		viewState ViewState
		want      bool
	}{
		{name: "list view", viewState: &ListViewState{}, want: false},
		{name: "detail view", viewState: &DetailViewState{}, want: true},
		{name: "confirm view", viewState: &ConfirmViewState{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{viewState: tt.viewState}
			if got := m.isDetailView(); got != tt.want {
				t.Errorf("Model.isDetailView() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_IsConfirming(t *testing.T) {
	tests := []struct {
		name      string
		viewState ViewState
		want      bool
	}{
		{name: "list view", viewState: &ListViewState{}, want: false},
		{name: "detail view", viewState: &DetailViewState{}, want: false},
		{name: "confirm view", viewState: &ConfirmViewState{}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{viewState: tt.viewState}
			if got := m.isConfirming(); got != tt.want {
				t.Errorf("Model.isConfirming() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_GetConfirmState(t *testing.T) {
	confirmState := &ConfirmViewState{Action: components.ActionPush, TargetIDs: []string{"ws-1"}}

	tests := []struct {
		name      string
		viewState ViewState
		wantNil   bool
	}{
		{name: "list view returns nil", viewState: &ListViewState{}, wantNil: true},
		{name: "detail view returns nil", viewState: &DetailViewState{}, wantNil: true},
		{name: "confirm view returns state", viewState: confirmState, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{viewState: tt.viewState}
			got := m.getConfirmState()

			if tt.wantNil && got != nil {
				t.Errorf("Model.getConfirmState() = %v, want nil", got)
			}

			if !tt.wantNil && got == nil {
				t.Error("Model.getConfirmState() = nil, want non-nil")
			}
		})
	}
}

func TestModel_GetDetailState(t *testing.T) {
	detailState := &DetailViewState{Loading: true}

	tests := []struct {
		name      string
		viewState ViewState
		wantNil   bool
	}{
		{name: "list view returns nil", viewState: &ListViewState{}, wantNil: true},
		{name: "detail view returns state", viewState: detailState, wantNil: false},
		{name: "confirm view returns nil", viewState: &ConfirmViewState{}, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{viewState: tt.viewState}
			got := m.getDetailState()

			if tt.wantNil && got != nil {
				t.Errorf("Model.getDetailState() = %v, want nil", got)
			}

			if !tt.wantNil && got == nil {
				t.Error("Model.getDetailState() = nil, want non-nil")
			}
		})
	}
}
