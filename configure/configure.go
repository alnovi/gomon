package configure

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

func LoadFromEnv(ctx context.Context, config any) error {
	if err := envconfig.Process(ctx, config); err != nil {
		return fmt.Errorf("failed to envconfig.Process: %w", err)
	}
	return nil
}
