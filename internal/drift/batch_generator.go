package drift

import (
	"context"
	"log/slog"
)

// EventEmitter provides event emission for drift test generation lifecycle.
type EventEmitter interface {
	EmitDriftTestGenerated(ctx context.Context, runID string, d Drift, gt *GeneratedTest)
	EmitDriftTestGenerationFailed(ctx context.Context, runID string, d Drift, err error)
}

// BatchGenerator generates tests for multiple drifts concurrently.
type BatchGenerator struct {
	generator *Generator
	max       int
	emitter   EventEmitter
}

// NewBatchGenerator creates a batch generator with a concurrency limit.
func NewBatchGenerator(generator *Generator, max int, emitter EventEmitter) *BatchGenerator {
	if max <= 0 {
		max = 5
	}
	return &BatchGenerator{
		generator: generator,
		max:       max,
		emitter:   emitter,
	}
}

// GenerateBatch generates tests for a list of drifts, limited to max concurrently.
func (bg *BatchGenerator) GenerateBatch(ctx context.Context, runID string, drifts []Drift) {
	if len(drifts) == 0 {
		return
	}
	sem := make(chan struct{}, bg.max)
	for _, d := range drifts {
		sem <- struct{}{} // acquire slot
		go func(d Drift) {
			defer func() { <-sem }() // release slot
			gt, err := bg.generator.GenerateForDrift(ctx, d)
			if err != nil {
				slog.Warn("auto-generate drift test failed",
					"drift_id", d.ID,
					"repository", d.Repository,
					"file_path", d.FilePath,
					"error", err,
				)
				if bg.emitter != nil {
					bg.emitter.EmitDriftTestGenerationFailed(ctx, runID, d, err)
				}
				return
			}
			slog.Info("auto-generated test for drift",
				"drift_id", d.ID,
				"test_id", gt.ID,
				"test_name", gt.TestName,
			)
			if bg.emitter != nil {
				bg.emitter.EmitDriftTestGenerated(ctx, runID, d, gt)
			}
		}(d)
	}
}
