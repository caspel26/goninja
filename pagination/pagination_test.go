package pagination

import (
	"net/url"
	"testing"

	"github.com/caspel26/goninja"
)

func TestParseLimitOffset(t *testing.T) {
	tests := []struct {
		name       string
		query      url.Values
		wantLimit  int
		wantOffset int
		wantErr    bool
	}{
		{
			name:       "defaults when unset",
			query:      url.Values{},
			wantLimit:  DefaultLimit,
			wantOffset: 0,
		},
		{
			name:       "explicit values",
			query:      url.Values{"limit": {"5"}, "offset": {"10"}},
			wantLimit:  5,
			wantOffset: 10,
		},
		{
			name:       "limit capped at MaxLimit",
			query:      url.Values{"limit": {"1000"}},
			wantLimit:  MaxLimit,
			wantOffset: 0,
		},
		{
			name:    "non-integer limit is a BadRequest",
			query:   url.Values{"limit": {"abc"}},
			wantErr: true,
		},
		{
			name:    "negative limit is a BadRequest",
			query:   url.Values{"limit": {"-1"}},
			wantErr: true,
		},
		{
			name:    "non-integer offset is a BadRequest",
			query:   url.Values{"offset": {"abc"}},
			wantErr: true,
		},
		{
			name:    "negative offset is a BadRequest",
			query:   url.Values{"offset": {"-1"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset, err := ParseLimitOffset(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseLimitOffset: err = nil, want a BadRequest")
				}
				if _, ok := err.(goninja.BadRequest); !ok {
					t.Errorf("ParseLimitOffset: err type = %T, want goninja.BadRequest", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseLimitOffset: unexpected error: %v", err)
			}
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}
