package goninja

import "testing"

func TestResourceConfig_RouteEnabled(t *testing.T) {
	tests := []struct {
		name  string
		cfg   ResourceConfig
		route Route
		want  bool
	}{
		{"empty Routes enables everything", ResourceConfig{}, RouteDelete, true},
		{"listed route enabled", ResourceConfig{Routes: []Route{RouteList, RouteRetrieve}}, RouteList, true},
		{"unlisted route disabled", ResourceConfig{Routes: []Route{RouteList, RouteRetrieve}}, RouteCreate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.RouteEnabled(tt.route); got != tt.want {
				t.Errorf("RouteEnabled(%q) = %v, want %v", tt.route, got, tt.want)
			}
		})
	}
}

func TestResourceConfig_PathOr(t *testing.T) {
	if got := (ResourceConfig{}).PathOr("/books"); got != "/books" {
		t.Errorf("PathOr with unset Path = %q, want %q", got, "/books")
	}
	if got := (ResourceConfig{Path: "/v1/books"}).PathOr("/books"); got != "/v1/books" {
		t.Errorf("PathOr with set Path = %q, want %q", got, "/v1/books")
	}
}
