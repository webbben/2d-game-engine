package entity

import (
	"testing"

	"github.com/webbben/2d-game-engine/model"
)

// createIntersectionResult is a helper to create an IntersectionResult for testing
func createIntersectionResult(intersects bool, dx, dy float64) model.IntersectionResult {
	return model.IntersectionResult{
		Intersects: intersects,
		Dx:         dx,
		Dy:         dy,
	}
}

// createCollisionResult is a helper to create a CollisionResult with consistent corner intersections
func createCollisionResult(tl, tr, bl, br model.IntersectionResult) model.CollisionResult {
	return model.CollisionResult{
		TopLeft:     tl,
		TopRight:    tr,
		BottomLeft:  bl,
		BottomRight: br,
	}
}

func TestCalculateCollisionAdjustment(t *testing.T) {
	tests := []struct {
		name                   string
		dx, dy                 float64
		cr                     model.CollisionResult
		expectedCx, expectedCy float64
	}{
		{
			name: "left full block",
			dx:   -10,
			dy:   0,
			cr: createCollisionResult(
				model.IntersectionResult{Intersects: true, Dx: 5, Dy: 0},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 3, Dy: 0},
				model.IntersectionResult{},
			),
			expectedCx: 5, // max(5, 3)
			expectedCy: 0,
		},
		{
			name: "left partial block",
			dx:   -10,
			dy:   0,
			cr: createCollisionResult(
				model.IntersectionResult{Intersects: true, Dx: 2, Dy: 0},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 4, Dy: 0},
				model.IntersectionResult{},
			),
			expectedCx: 4, // max(2, 4)
			expectedCy: 0,
		},
		{
			name: "left no collision (full open)",
			dx:   -10,
			dy:   0,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			expectedCx: 0,
			expectedCy: 0,
		},
		{
			name: "right full block",
			dx:   10,
			dy:   0,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 5, Dy: 0},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 3, Dy: 0},
			),
			expectedCx: -5, // -max(5, 3)
			expectedCy: 0,
		},
		{
			name: "up full block",
			dx:   0,
			dy:   -10,
			cr: createCollisionResult(
				model.IntersectionResult{Intersects: true, Dx: 0, Dy: 5},
				model.IntersectionResult{Intersects: true, Dx: 0, Dy: 3},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			expectedCx: 0,
			expectedCy: 5, // max(5, 3)
		},
		{
			name: "down full block",
			dx:   0,
			dy:   10,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 0, Dy: 5},
				model.IntersectionResult{Intersects: true, Dx: 0, Dy: 3},
			),
			expectedCx: 0,
			expectedCy: -5, // -max(5, 3)
		},
		{
			name: "moving diagonal up-right: walk into outward corner slightly above the point",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 5, Dy: 8},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			// we expect to slide up
			expectedCx: -5,
			expectedCy: 0,
		},
		{
			name: "moving diagonal up-right: walk into outward corner slightly below the point",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 7, Dy: 5},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			// we expect to slide right
			expectedCx: 0,
			expectedCy: 5,
		},
		{
			name: "moving diagonal up-right: walk into a flat wall above you",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				// I don't think the dx here matters, so I put random values for it
				// but, in this case, Dy should be the same for both I think.
				model.IntersectionResult{Intersects: true, Dx: 20, Dy: 8},
				model.IntersectionResult{Intersects: true, Dx: 7, Dy: 8},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			// we expect to slide right
			expectedCx: 0,
			expectedCy: 8,
		},
		{
			name: "moving diagonal up-right: walk into a flat wall to the right",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 7, Dy: 8},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 7, Dy: 20},
			),
			// we expect to slide up
			expectedCx: -7,
			expectedCy: 0,
		},
		{
			name: "moving diagonal up-right: walking around (but clipping) a corner on the top left",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				model.IntersectionResult{Intersects: true, Dx: 3, Dy: 5},
				model.IntersectionResult{},
				model.IntersectionResult{},
				model.IntersectionResult{},
			),
			// keep moving right until the corner has been completely passed
			expectedCx: 0,
			expectedCy: 5,
		},
		{
			name: "moving diagonal up-right: walking into an inward corner",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				// tl and tr have same Dy, tr and br have same Dx
				model.IntersectionResult{Intersects: true, Dx: 3, Dy: 5},
				model.IntersectionResult{Intersects: true, Dx: 8, Dy: 5},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 8, Dy: 9},
			),
			// move into the inward corner as much as possible
			expectedCx: -8, // tr and br's Dx is negated
			expectedCy: 5,  // tl and tr's Dy is negated (positive Cy since movement Dy is negative)
		},
		{
			name: "moving diagonal up-right: walking into an inward corner, while already standing directly against corner",
			dx:   10,
			dy:   -10,
			cr: createCollisionResult(
				// tl and tr have same Dy, tr and br have same Dx
				model.IntersectionResult{Intersects: true, Dx: 3, Dy: 10},
				model.IntersectionResult{Intersects: true, Dx: 10, Dy: 10},
				model.IntersectionResult{},
				model.IntersectionResult{Intersects: true, Dx: 10, Dy: 9},
			),
			// don't move at all; all movement should be cancelled out
			expectedCx: -10, // tr and br's Dx is negated
			expectedCy: 10,  // tl and tr's Dy is negated (positive Cy since movement Dy is negative)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cx, cy := calculateCollisionAdjustment(tt.dx, tt.dy, tt.cr)
			if cx != tt.expectedCx {
				t.Errorf("cx = %v, want %v", cx, tt.expectedCx)
			}
			if cy != tt.expectedCy {
				t.Errorf("cy = %v, want %v", cy, tt.expectedCy)
			}
		})
	}
}
