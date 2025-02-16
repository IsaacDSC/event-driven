package SDK

import (
	"context"
	"event-driven/types"
	"testing"
)

func TestConsumer(t *testing.T) {
	testCases := []struct {
		name      string
		consumers map[string]types.ConsumerFn
		expectMux bool
	}{
		{
			name: "Single handler",
			consumers: map[string]types.ConsumerFn{
				"event_example_01": func(ctx context.Context, payload map[string]any) error {
					return nil
				},
			},
			expectMux: true,
		},
		{
			name: "Multiple handlers",
			consumers: map[string]types.ConsumerFn{
				"event_example_01": func(ctx context.Context, payload map[string]any) error {
					return nil
				},
				"event_example_02": func(ctx context.Context, payload map[string]any) error {
					return nil
				},
			},
			expectMux: true,
		},
		{
			name:      "No handlers",
			consumers: map[string]types.ConsumerFn{},
			expectMux: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			consumer := NewConsumerServer("localhost:6379")
			consumer.AddHandlers(tc.consumers)
			if (consumer.mux != nil) != tc.expectMux {
				t.Fatalf("expected mux to be %v, got %v", tc.expectMux, consumer.mux != nil)
			}

			consumer.Start()
		})
	}
}
