package admission

import (
	"testing"

	"github.com/maxcelant/jap/internal/schema"
	"k8s.io/utils/ptr"
)

func TestDefault(t *testing.T) {
	t.Run("defaults both strategy and match", func(t *testing.T) {
		app := &schema.App{
			Routes: []schema.Route{
				{Path: "/api", Sink: "backend"},
			},
			Sinks: []schema.Sink{
				{
					Name: "backend",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080},
					},
				},
			},
		}

		err := Default(app)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if app.Routes[0].Match == nil || *app.Routes[0].Match != "exact" {
			t.Errorf("expected match to be 'exact', got %v", app.Routes[0].Match)
		}
		if app.Sinks[0].Strategy == nil || *app.Sinks[0].Strategy != "random" {
			t.Errorf("expected strategy to be 'random', got %v", app.Sinks[0].Strategy)
		}
	})
}

func TestSetStrategy(t *testing.T) {
	tests := []struct {
		name             string
		sinks            []schema.Sink
		expectedStrategy []string
	}{
		{
			name: "all upstreams weighted sets strategy to weighted",
			sinks: []schema.Sink{
				{
					Name: "backend",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080, Weight: ptr.To(10)},
						{Address: "127.0.0.1", Port: 8081, Weight: ptr.To(20)},
					},
				},
			},
			expectedStrategy: []string{"weighted"},
		},
		{
			name: "some upstreams without weight sets strategy to random",
			sinks: []schema.Sink{
				{
					Name: "backend",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080, Weight: ptr.To(10)},
						{Address: "127.0.0.1", Port: 8081},
					},
				},
			},
			expectedStrategy: []string{"random"},
		},
		{
			name: "upstream with zero weight sets strategy to random",
			sinks: []schema.Sink{
				{
					Name: "backend",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080, Weight: ptr.To(0)},
					},
				},
			},
			expectedStrategy: []string{"random"},
		},
		{
			name: "no upstreams sets strategy to random",
			sinks: []schema.Sink{
				{
					Name:      "backend",
					Upstreams: []schema.Upstream{},
				},
			},
			expectedStrategy: []string{"random"},
		},
		{
			name: "existing strategy is not overwritten",
			sinks: []schema.Sink{
				{
					Name:     "backend",
					Strategy: ptr.To("round-robin"),
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080, Weight: ptr.To(10)},
					},
				},
			},
			expectedStrategy: []string{"round-robin"},
		},
		{
			name: "multiple sinks with different scenarios",
			sinks: []schema.Sink{
				{
					Name: "weighted-sink",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 8080, Weight: ptr.To(10)},
						{Address: "127.0.0.1", Port: 8081, Weight: ptr.To(20)},
					},
				},
				{
					Name: "random-sink",
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 9000},
					},
				},
				{
					Name:     "preset-sink",
					Strategy: ptr.To("custom"),
					Upstreams: []schema.Upstream{
						{Address: "127.0.0.1", Port: 9001},
					},
				},
			},
			expectedStrategy: []string{"weighted", "random", "custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &schema.App{Sinks: tt.sinks}
			err := Default(app)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i, expected := range tt.expectedStrategy {
				if app.Sinks[i].Strategy == nil {
					t.Errorf("sink %d: expected strategy %q, got nil", i, expected)
					continue
				}
				if *app.Sinks[i].Strategy != expected {
					t.Errorf("sink %d: expected strategy %q, got %q", i, expected, *app.Sinks[i].Strategy)
				}
			}
		})
	}
}

func TestSetMatch(t *testing.T) {
	tests := []struct {
		name          string
		routes        []schema.Route
		expectedMatch []string
	}{
		{
			name: "nil match defaults to exact",
			routes: []schema.Route{
				{Path: "/api", Sink: "backend"},
			},
			expectedMatch: []string{"exact"},
		},
		{
			name: "empty string match defaults to exact",
			routes: []schema.Route{
				{Path: "/api", Match: ptr.To(""), Sink: "backend"},
			},
			expectedMatch: []string{"exact"},
		},
		{
			name: "existing match is not overwritten",
			routes: []schema.Route{
				{Path: "/api", Match: ptr.To("prefix"), Sink: "backend"},
			},
			expectedMatch: []string{"prefix"},
		},
		{
			name: "multiple routes with different scenarios",
			routes: []schema.Route{
				{Path: "/api", Sink: "backend"},
				{Path: "/static", Match: ptr.To("prefix"), Sink: "static"},
				{Path: "/health", Match: ptr.To(""), Sink: "health"},
				{Path: "/pattern", Match: ptr.To("regex"), Sink: "pattern"},
			},
			expectedMatch: []string{"exact", "prefix", "exact", "regex"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &schema.App{Routes: tt.routes}
			err := Default(app)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i, expected := range tt.expectedMatch {
				if app.Routes[i].Match == nil {
					t.Errorf("route %d: expected match %q, got nil", i, expected)
					continue
				}
				if *app.Routes[i].Match != expected {
					t.Errorf("route %d: expected match %q, got %q", i, expected, *app.Routes[i].Match)
				}
			}
		})
	}
}
