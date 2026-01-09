package admission

import (
	"fmt"

	"github.com/maxcelant/sinkplot/internal/schema"
	"k8s.io/utils/ptr"
)

type defaulter struct {
	app *schema.App
}

// Default directly modifies the passed in object
func Default(app *schema.App) error {
	d := defaulter{app}
	if err := d.setStrategy(); err != nil {
		return fmt.Errorf("failed to set a default strategy: %w", err)
	}
	if err := d.setMatch(); err != nil {
		return fmt.Errorf("failed to set a default strategy: %w", err)
	}
	return nil
}

// setStrategy sets a default sink strategy based on the sink fields and whether the
// weight is set for all the upstreams
func (d *defaulter) setStrategy() error {
	// Find all the sink objects that have an unset strategy field
	var missing []struct {
		index    int
		weighted bool
	}
	for i, s := range d.app.Sinks {
		if s.Strategy == nil {
			// Check if all upstreams have a weight set
			allWeighted := len(s.Upstreams) > 0
			for _, u := range s.Upstreams {
				if u.Weight == nil || *u.Weight == 0 {
					allWeighted = false
					break
				}
			}
			missing = append(missing, struct {
				index    int
				weighted bool
			}{index: i, weighted: allWeighted})
		}
	}
	// default to weighted or random strategy based on upstream weights
	for _, m := range missing {
		if m.weighted {
			d.app.Sinks[m.index].Strategy = ptr.To("weighted")
		} else {
			d.app.Sinks[m.index].Strategy = ptr.To("random")
		}
	}
	return nil
}

// setMatch sets any unset route matcher to exact
func (d *defaulter) setMatch() error {
	for i, route := range d.app.Routes {
		if route.Match == nil || *route.Match == "" {
			d.app.Routes[i].Match = ptr.To("exact")
		}
	}
	return nil
}
