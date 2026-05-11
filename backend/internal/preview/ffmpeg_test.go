package preview

import (
	"math"
	"testing"
)

func TestNewDefaultsToThreeSecondTeaserSegments(t *testing.T) {
	gen := New(Config{})
	if gen.cfg.DurationSeconds != 3 {
		t.Fatalf("DurationSeconds = %d, want 3", gen.cfg.DurationSeconds)
	}
}

func TestMediumVideoPreviewPlanUsesFourThreeSecondSegments(t *testing.T) {
	plan := buildTeaserPlan(Config{DurationSeconds: 3, Segments: 3}, 300)
	if len(plan.starts) != 4 {
		t.Fatalf("segments = %d, want 4", len(plan.starts))
	}
	if plan.eachSec != 3 {
		t.Fatalf("eachSec = %.2f, want 3", plan.eachSec)
	}
	want := []float64{15, 95, 175, 255}
	for i := range want {
		if math.Abs(plan.starts[i]-want[i]) > 0.001 {
			t.Fatalf("start[%d] = %.2f, want %.2f", i, plan.starts[i], want[i])
		}
	}
}

func TestLongVideoPreviewPlanUsesFourThreeSecondSegments(t *testing.T) {
	plan := buildTeaserPlan(Config{DurationSeconds: 15, Segments: 3}, 601)
	if len(plan.starts) != 4 {
		t.Fatalf("segments = %d, want 4", len(plan.starts))
	}
	if plan.eachSec != 3 {
		t.Fatalf("eachSec = %.2f, want 3", plan.eachSec)
	}
	want := []float64{120.2, 240.4, 360.6, 480.8}
	for i := range want {
		if math.Abs(plan.starts[i]-want[i]) > 0.001 {
			t.Fatalf("start[%d] = %.2f, want %.2f", i, plan.starts[i], want[i])
		}
	}
}

func TestShortVideoPreviewPlanUsesUpToThreeThreeSecondSegments(t *testing.T) {
	plan := buildTeaserPlan(Config{DurationSeconds: 15, Segments: 3}, 20)
	if len(plan.starts) != 3 {
		t.Fatalf("segments = %d, want 3", len(plan.starts))
	}
	if plan.eachSec != 3 {
		t.Fatalf("eachSec = %.2f, want 3", plan.eachSec)
	}
	want := []float64{2, 9.5, 17}
	for i := range want {
		if math.Abs(plan.starts[i]-want[i]) > 0.001 {
			t.Fatalf("start[%d] = %.2f, want %.2f", i, plan.starts[i], want[i])
		}
	}
}

func TestShortVideoPreviewPlanDropsSegmentsThatDoNotFit(t *testing.T) {
	plan := buildTeaserPlan(Config{DurationSeconds: 15, Segments: 3}, 8)
	if len(plan.starts) != 2 {
		t.Fatalf("segments = %d, want 2", len(plan.starts))
	}
	if plan.eachSec != 3 {
		t.Fatalf("eachSec = %.2f, want 3", plan.eachSec)
	}
	want := []float64{0.8, 5}
	for i := range want {
		if math.Abs(plan.starts[i]-want[i]) > 0.001 {
			t.Fatalf("start[%d] = %.2f, want %.2f", i, plan.starts[i], want[i])
		}
	}
}

func TestShortVideoPreviewPlanReturnsNoSegmentsWhenOneSegmentCannotFit(t *testing.T) {
	plan := buildTeaserPlan(Config{DurationSeconds: 15, Segments: 3}, 2.5)
	if len(plan.starts) != 0 {
		t.Fatalf("segments = %d, want 0", len(plan.starts))
	}
	if plan.eachSec != 3 {
		t.Fatalf("eachSec = %.2f, want 3", plan.eachSec)
	}
}
