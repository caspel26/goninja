package goninja

import "testing"

// actionsOptionFake is a minimal SetActions-satisfying type, just enough
// to prove Actions' plumbing without needing a full generated resource.
type actionsOptionFake struct {
	id      string
	actions []Action
}

func (f *actionsOptionFake) SetActions(actions ...Action) {
	f.actions = actions
}

func TestActions_BuildsAndAttachesActions(t *testing.T) {
	f := &actionsOptionFake{id: "fake-1"}

	opt := Actions(func(r *actionsOptionFake, suffix string) []Action {
		return []Action{{Name: "action-for-" + r.id + "-" + suffix}}
	}, "extra")
	opt(f)

	if len(f.actions) != 1 || f.actions[0].Name != "action-for-fake-1-extra" {
		t.Errorf("actions = %+v, want one action built from r and arg", f.actions)
	}
}

func TestActions_EmptyBuildResultClearsActions(t *testing.T) {
	f := &actionsOptionFake{actions: []Action{{Name: "stale"}}}

	Actions(func(r *actionsOptionFake, _ string) []Action { return nil }, "")(f)

	if f.actions != nil {
		t.Errorf("actions = %+v, want nil after a build func returning none", f.actions)
	}
}

func TestActions_NoExtraArgumentViaEmptyStruct(t *testing.T) {
	f := &actionsOptionFake{id: "fake-2"}

	// A build func that genuinely needs nothing beyond r still works —
	// arg's type is whatever the caller's build func wants, including
	// struct{}{} when there's nothing to pass.
	Actions(func(r *actionsOptionFake, _ struct{}) []Action {
		return []Action{{Name: "action-for-" + r.id}}
	}, struct{}{})(f)

	if len(f.actions) != 1 || f.actions[0].Name != "action-for-fake-2" {
		t.Errorf("actions = %+v, want one action built from r alone", f.actions)
	}
}

// TestActions_AssignableToGeneratedOptionType proves the doc comment's
// claim: the func(R) Actions returns is an unnamed function value, so
// it's directly assignable to a generated <Model>Option (a named type
// with the same underlying signature) without an explicit conversion —
// exactly how New<Model>Resource(db, goninja.Actions(build, arg)) type-
// checks in generated code.
func TestActions_AssignableToGeneratedOptionType(t *testing.T) {
	type fakeOption func(*actionsOptionFake)

	var opt fakeOption = Actions(func(r *actionsOptionFake, _ struct{}) []Action {
		return []Action{{Name: "assignable"}}
	}, struct{}{})

	f := &actionsOptionFake{}
	opt(f)
	if len(f.actions) != 1 || f.actions[0].Name != "assignable" {
		t.Errorf("actions = %+v, want one action after applying the assigned option", f.actions)
	}
}
