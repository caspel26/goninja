package goninja

import (
	"context"
	"testing"
)

type fakeUser struct{ id string }

func (u fakeUser) ID() string { return u.id }

func TestWithUser_UserFromContext(t *testing.T) {
	ctx := WithUser(context.Background(), fakeUser{id: "u-1"})

	got, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("UserFromContext: ok = false, want true")
	}
	if got.ID() != "u-1" {
		t.Errorf("UserFromContext: ID = %q, want %q", got.ID(), "u-1")
	}
}

func TestUserFromContext_NotSet(t *testing.T) {
	_, ok := UserFromContext(context.Background())
	if ok {
		t.Error("UserFromContext on a bare context: ok = true, want false")
	}
}
