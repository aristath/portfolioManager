package work

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	wt := &WorkType{
		ID: "test:work",
	}

	r.Register(wt)

	assert.Equal(t, 1, r.Count())
	assert.True(t, r.Has("test:work"))
}

func TestRegistry_RegisterOverwrites(t *testing.T) {
	r := NewRegistry()

	wt1 := &WorkType{
		ID: "test:work",
	}
	wt2 := &WorkType{
		ID: "test:work",
	}

	r.Register(wt1)
	r.Register(wt2)

	assert.Equal(t, 1, r.Count())
	got := r.Get("test:work")
	assert.NotNil(t, got)
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	wt := &WorkType{
		ID:           "security:sync",
		MarketTiming: AfterMarketClose,
	}
	r.Register(wt)

	t.Run("returns registered work type", func(t *testing.T) {
		got := r.Get("security:sync")
		require.NotNil(t, got)
		assert.Equal(t, "security:sync", got.ID)
		assert.Equal(t, AfterMarketClose, got.MarketTiming)
	})

	t.Run("returns nil for unknown ID", func(t *testing.T) {
		got := r.Get("unknown:work")
		assert.Nil(t, got)
	})
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()

	r.Register(&WorkType{ID: "test:work"})

	assert.True(t, r.Has("test:work"))
	assert.False(t, r.Has("unknown:work"))
}

func TestRegistry_Remove(t *testing.T) {
	r := NewRegistry()

	r.Register(&WorkType{ID: "test:work"})
	assert.True(t, r.Has("test:work"))

	r.Remove("test:work")
	assert.False(t, r.Has("test:work"))
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_IDs(t *testing.T) {
	r := NewRegistry()

	r.Register(&WorkType{ID: "planner:weights"})
	r.Register(&WorkType{ID: "security:sync"})
	r.Register(&WorkType{ID: "maintenance:backup"})

	ids := r.IDs()

	// IDs should be sorted alphabetically
	assert.Equal(t, []string{"maintenance:backup", "planner:weights", "security:sync"}, ids)
}

func TestRegistry_ByPriority(t *testing.T) {
	t.Skip("Priority ordering removed in favor of FIFO registration order")
}

func TestRegistry_ByPriority_ReturnsACopy(t *testing.T) {
	t.Skip("Priority ordering removed in favor of FIFO registration order - see TestRegistry_All_ReturnsCopy")
}

func TestRegistry_GetDependencies(t *testing.T) {
	r := NewRegistry()

	r.Register(&WorkType{ID: "planner:weights"})
	r.Register(&WorkType{ID: "planner:context", DependsOn: []string{"planner:weights"}})
	r.Register(&WorkType{ID: "planner:plan", DependsOn: []string{"planner:context"}})

	t.Run("returns dependencies", func(t *testing.T) {
		deps := r.GetDependencies("planner:context")

		require.Len(t, deps, 1)
		assert.Equal(t, "planner:weights", deps[0].ID)
	})

	t.Run("returns nil for work type with no dependencies", func(t *testing.T) {
		deps := r.GetDependencies("planner:weights")

		assert.Len(t, deps, 0)
	})

	t.Run("returns nil for unknown work type", func(t *testing.T) {
		deps := r.GetDependencies("unknown:work")

		assert.Nil(t, deps)
	})

	t.Run("filters out missing dependencies", func(t *testing.T) {
		r.Register(&WorkType{ID: "test:orphan", DependsOn: []string{"missing:dep"}})

		deps := r.GetDependencies("test:orphan")

		assert.Len(t, deps, 0)
	})
}

func TestRegistry_GetDependents(t *testing.T) {
	r := NewRegistry()

	r.Register(&WorkType{ID: "planner:weights"})
	r.Register(&WorkType{ID: "planner:context", DependsOn: []string{"planner:weights"}})
	r.Register(&WorkType{ID: "planner:plan", DependsOn: []string{"planner:context"}})

	t.Run("returns dependents", func(t *testing.T) {
		dependents := r.GetDependents("planner:weights")

		require.Len(t, dependents, 1)
		assert.Equal(t, "planner:context", dependents[0].ID)
	})

	t.Run("returns empty for work type with no dependents", func(t *testing.T) {
		dependents := r.GetDependents("planner:plan")

		assert.Len(t, dependents, 0)
	})

	t.Run("returns empty for unknown work type", func(t *testing.T) {
		dependents := r.GetDependents("unknown:work")

		assert.Len(t, dependents, 0)
	})
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	// Pre-register some work types
	for i := 0; i < 10; i++ {
		r.Register(&WorkType{ID: "initial:" + string(rune('a'+i))})
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = r.Get("initial:a")
				_ = r.Has("initial:b")
				_ = r.Count()
				_ = r.IDs()
				_ = r.All()
			}
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.Register(&WorkType{ID: "concurrent:" + string(rune('a'+id))})
				r.Remove("concurrent:" + string(rune('a'+id)))
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestRegistry_FullWorkflowExample(t *testing.T) {
	r := NewRegistry()

	// Register a planner chain with dependencies
	r.Register(&WorkType{
		ID: "planner:weights",
		FindSubjects: func() []string {
			return []string{""} // Global work
		},
		Execute: func(ctx context.Context, subject string, progress *ProgressReporter) error {
			return nil
		},
	})

	r.Register(&WorkType{
		ID:        "planner:context",
		DependsOn: []string{"planner:weights"},
		FindSubjects: func() []string {
			return []string{""}
		},
		Execute: func(ctx context.Context, subject string, progress *ProgressReporter) error {
			return nil
		},
	})

	r.Register(&WorkType{
		ID:        "planner:plan",
		DependsOn: []string{"planner:context"},
		FindSubjects: func() []string {
			return []string{""}
		},
		Execute: func(ctx context.Context, subject string, progress *ProgressReporter) error {
			return nil
		},
	})

	// Verify the chain
	assert.Equal(t, 3, r.Count())

	// Check dependency chain
	contextDeps := r.GetDependencies("planner:context")
	require.Len(t, contextDeps, 1)
	assert.Equal(t, "planner:weights", contextDeps[0].ID)

	planDeps := r.GetDependencies("planner:plan")
	require.Len(t, planDeps, 1)
	assert.Equal(t, "planner:context", planDeps[0].ID)

	// Check reverse dependencies
	weightsDependents := r.GetDependents("planner:weights")
	require.Len(t, weightsDependents, 1)
	assert.Equal(t, "planner:context", weightsDependents[0].ID)
}

// Phase 1 Tests: Registration Order (FIFO Queue Implementation)

func TestRegistry_All_RegistrationOrder(t *testing.T) {
	registry := NewRegistry()

	// Register in specific order
	registry.Register(&WorkType{ID: "first"})
	registry.Register(&WorkType{ID: "second"})
	registry.Register(&WorkType{ID: "third"})

	all := registry.All()

	assert.Equal(t, 3, len(all))
	assert.Equal(t, "first", all[0].ID)
	assert.Equal(t, "second", all[1].ID)
	assert.Equal(t, "third", all[2].ID)
}

func TestRegistry_All_ReturnsCopy(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&WorkType{ID: "test"})

	all1 := registry.All()
	all2 := registry.All()

	// Should be different slices (copies) - verify by modifying one
	assert.Equal(t, len(all1), len(all2))
	all1[0] = nil
	assert.NotNil(t, all2[0], "Modifying one slice should not affect the other")
}
