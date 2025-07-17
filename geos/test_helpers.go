package geos

import (
	"testing"
)

// TestHelper provides utility functions for testing
type TestHelper struct {
	service *Service
	t       *testing.T
}

// NewTestHelper creates a new test helper with a GEOS service
func NewTestHelper(t *testing.T) *TestHelper {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	return &TestHelper{
		service: service,
		t:       t,
	}
}

// Close cleans up the test helper
func (th *TestHelper) Close() {
	th.service.Close()
}

// ParseWKT parses WKT geometry with automatic error handling
func (th *TestHelper) ParseWKT(wkt string) *Geometry {
	input := GeometryInput{WKT: wkt}
	geom, err := th.service.ParseGeometry(input)
	if err != nil {
		th.t.Fatalf("Failed to parse WKT %s: %v", wkt, err)
	}
	return geom
}

// ParseGeoJSON parses GeoJSON geometry with automatic error handling
func (th *TestHelper) ParseGeoJSON(geoJSON map[string]interface{}) *Geometry {
	input := GeometryInput{GeoJSON: geoJSON}
	geom, err := th.service.ParseGeometry(input)
	if err != nil {
		th.t.Fatalf("Failed to parse GeoJSON: %v", err)
	}
	return geom
}

// AssertWithin asserts that geometry A is within geometry B
func (th *TestHelper) AssertWithin(a, b *Geometry, expected bool) {
	result, err := th.service.Within(a, b)
	if err != nil {
		th.t.Fatalf("Failed to test within: %v", err)
	}
	if result != expected {
		th.t.Errorf("Expected within result %t, got %t", expected, result)
	}
}

// AssertIntersects asserts that two geometries intersect
func (th *TestHelper) AssertIntersects(a, b *Geometry, expected bool) {
	result, err := th.service.Intersects(a, b)
	if err != nil {
		th.t.Fatalf("Failed to test intersects: %v", err)
	}
	if result != expected {
		th.t.Errorf("Expected intersects result %t, got %t", expected, result)
	}
}

// AssertDistance asserts that the distance between two geometries is within tolerance
func (th *TestHelper) AssertDistance(a, b *Geometry, expected, tolerance float64) {
	result, err := th.service.Distance(a, b)
	if err != nil {
		th.t.Fatalf("Failed to calculate distance: %v", err)
	}
	if result < expected-tolerance || result > expected+tolerance {
		th.t.Errorf("Expected distance %.3f±%.3f, got %.3f", expected, tolerance, result)
	}
}

// AssertBuffer creates a buffer and validates it's not nil
func (th *TestHelper) AssertBuffer(geom *Geometry, radius float64) *Geometry {
	result, err := th.service.Buffer(geom, radius)
	if err != nil {
		th.t.Fatalf("Failed to create buffer: %v", err)
	}
	if result == nil {
		th.t.Fatal("Expected non-nil buffer result")
	}
	return result
}

// AssertSimplify simplifies a geometry and validates it's not nil
func (th *TestHelper) AssertSimplify(geom *Geometry, tolerance float64) *Geometry {
	result, err := th.service.Simplify(geom, tolerance)
	if err != nil {
		th.t.Fatalf("Failed to simplify geometry: %v", err)
	}
	if result == nil {
		th.t.Fatal("Expected non-nil simplify result")
	}
	return result
}

// AssertUnion creates a union of geometries and validates it's not nil
func (th *TestHelper) AssertUnion(geometries []*Geometry) *Geometry {
	result, err := th.service.Union(geometries)
	if err != nil {
		th.t.Fatalf("Failed to create union: %v", err)
	}
	if result == nil {
		th.t.Fatal("Expected non-nil union result")
	}
	return result
}

// AssertDifference creates a difference between geometries and validates it's not nil
func (th *TestHelper) AssertDifference(a, b *Geometry) *Geometry {
	result, err := th.service.Difference(a, b)
	if err != nil {
		th.t.Fatalf("Failed to create difference: %v", err)
	}
	if result == nil {
		th.t.Fatal("Expected non-nil difference result")
	}
	return result
}

// AssertToWKT converts geometry to WKT and validates it's not empty
func (th *TestHelper) AssertToWKT(geom *Geometry) string {
	result, err := th.service.ToWKT(geom)
	if err != nil {
		th.t.Fatalf("Failed to convert to WKT: %v", err)
	}
	if result == "" {
		th.t.Fatal("Expected non-empty WKT result")
	}
	return result
}

// AssertValidateGeometry validates geometry input
func (th *TestHelper) AssertValidateGeometry(input GeometryInput, expected bool) {
	err := th.service.ValidateGeometry(input)
	if expected && err != nil {
		th.t.Errorf("Expected valid geometry, got error: %v", err)
	} else if !expected && err == nil {
		th.t.Error("Expected invalid geometry, got nil error")
	}
}

// Common test geometries
func (th *TestHelper) PointGeometry() *Geometry {
	return th.ParseWKT("POINT(1.0 1.0)")
}

func (th *TestHelper) LineGeometry() *Geometry {
	return th.ParseWKT("LINESTRING(0 0, 1 1, 2 2)")
}

func (th *TestHelper) PolygonGeometry() *Geometry {
	return th.ParseWKT("POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))")
}

func (th *TestHelper) ComplexLineGeometry() *Geometry {
	return th.ParseWKT("LINESTRING(0 0, 0.5 0.1, 1.0 0.2, 1.5 0.1, 2.0 0)")
}

func (th *TestHelper) OverlappingPolygonGeometry() *Geometry {
	return th.ParseWKT("POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))")
}

func (th *TestHelper) PointGeoJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":        "Point",
		"coordinates": []interface{}{1.0, 2.0},
	}
}

func (th *TestHelper) LineGeoJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "LineString",
		"coordinates": []interface{}{
			[]interface{}{0.0, 0.0},
			[]interface{}{1.0, 1.0},
			[]interface{}{2.0, 2.0},
		},
	}
}

func (th *TestHelper) PolygonGeoJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "Polygon",
		"coordinates": []interface{}{
			[]interface{}{
				[]interface{}{0.0, 0.0},
				[]interface{}{2.0, 0.0},
				[]interface{}{2.0, 2.0},
				[]interface{}{0.0, 2.0},
				[]interface{}{0.0, 0.0},
			},
		},
	}
}