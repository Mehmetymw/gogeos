package geos

import (
	"strings"
	"testing"
)

// TestNewService tests the creation of a new GEOS service
func TestNewService(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	if service.context == nil {
		t.Fatal("Service context should not be nil")
	}
}

// TestServiceClose tests the cleanup of GEOS service
func TestServiceClose(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}

	// Close the service
	service.Close()

	// Verify context is cleaned up
	if service.context != nil {
		t.Error("Service context should be nil after Close()")
	}

	// Multiple closes should be safe
	service.Close()
}

// TestParseGeometry_WKT tests parsing WKT geometries
func TestParseGeometry_WKT(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		wkt      string
		expected bool
	}{
		{"Point", "POINT(1.0 2.0)", true},
		{"LineString", "LINESTRING(0 0, 1 1, 2 2)", true},
		{"Polygon", "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", true},
		{"MultiPoint", "MULTIPOINT((0 0), (1 1))", true},
		{"MultiLineString", "MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", true},
		{"MultiPolygon", "MULTIPOLYGON(((0 0, 1 0, 1 1, 0 1, 0 0)), ((2 2, 3 2, 3 3, 2 3, 2 2)))", true},
		{"Empty WKT", "", false},
		{"Invalid WKT", "INVALID_GEOMETRY", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := GeometryInput{WKT: tc.wkt}
			geom, err := service.ParseGeometry(input)

			if tc.expected {
				if err != nil {
					t.Errorf("Expected successful parsing of %s, got error: %v", tc.wkt, err)
				}
				if geom == nil {
					t.Error("Expected non-nil geometry")
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.wkt)
				}
				if geom != nil {
					t.Error("Expected nil geometry for invalid input")
				}
			}
		})
	}
}

// TestParseGeometry_GeoJSON tests parsing GeoJSON geometries
func TestParseGeometry_GeoJSON(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geoJSON  map[string]interface{}
		expected bool
	}{
		{
			"Point",
			map[string]interface{}{
				"type":        "Point",
				"coordinates": []interface{}{1.0, 2.0},
			},
			true,
		},
		{
			"LineString",
			map[string]interface{}{
				"type":        "LineString",
				"coordinates": []interface{}{[]interface{}{0.0, 0.0}, []interface{}{1.0, 1.0}},
			},
			true,
		},
		{
			"Polygon",
			map[string]interface{}{
				"type": "Polygon",
				"coordinates": []interface{}{
					[]interface{}{
						[]interface{}{0.0, 0.0},
						[]interface{}{1.0, 0.0},
						[]interface{}{1.0, 1.0},
						[]interface{}{0.0, 1.0},
						[]interface{}{0.0, 0.0},
					},
				},
			},
			true,
		},
		{
			"Invalid - Missing type",
			map[string]interface{}{
				"coordinates": []interface{}{1.0, 2.0},
			},
			false,
		},
		{
			"Invalid - Missing coordinates",
			map[string]interface{}{
				"type": "Point",
			},
			false,
		},
		{
			"Invalid - Unsupported type",
			map[string]interface{}{
				"type":        "MultiPoint",
				"coordinates": []interface{}{[]interface{}{1.0, 2.0}},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := GeometryInput{GeoJSON: tc.geoJSON}
			geom, err := service.ParseGeometry(input)

			if tc.expected {
				if err != nil {
					t.Errorf("Expected successful parsing of GeoJSON, got error: %v", err)
				}
				if geom == nil {
					t.Error("Expected non-nil geometry")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid GeoJSON")
				}
				if geom != nil {
					t.Error("Expected nil geometry for invalid input")
				}
			}
		})
	}
}

// TestParseGeometry_EmptyInput tests parsing with empty input
func TestParseGeometry_EmptyInput(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	input := GeometryInput{}
	geom, err := service.ParseGeometry(input)

	if err == nil {
		t.Error("Expected error for empty input")
	}
	if geom != nil {
		t.Error("Expected nil geometry for empty input")
	}
}

// TestToWKT tests converting geometry to WKT
func TestToWKT(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name        string
		inputWKT    string
		expectError bool
	}{
		{"Point", "POINT(1.0 2.0)", false},
		{"LineString", "LINESTRING(0 0, 1 1)", false},
		{"Polygon", "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := GeometryInput{WKT: tc.inputWKT}
			geom, err := service.ParseGeometry(input)
			if err != nil {
				t.Fatalf("Failed to parse geometry: %v", err)
			}

			wkt, err := service.ToWKT(geom)
			if tc.expectError {
				if err == nil {
					t.Error("Expected error converting to WKT")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error converting to WKT: %v", err)
				}
				if wkt == "" {
					t.Error("Expected non-empty WKT string")
				}
			}
		})
	}
}

// TestToWKT_NilGeometry tests converting nil geometry to WKT
func TestToWKT_NilGeometry(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	wkt, err := service.ToWKT(nil)
	if err == nil {
		t.Error("Expected error for nil geometry")
	}
	if wkt != "" {
		t.Error("Expected empty WKT string for nil geometry")
	}
}

// TestWithin tests spatial within relationship
func TestWithin(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geomA    string
		geomB    string
		expected bool
	}{
		{
			"Point within polygon",
			"POINT(1.0 1.0)",
			"POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			true,
		},
		{
			"Point outside polygon",
			"POINT(3.0 3.0)",
			"POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			false,
		},
		{
			"Point on boundary",
			"POINT(0.0 0.0)",
			"POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputA := GeometryInput{WKT: tc.geomA}
			inputB := GeometryInput{WKT: tc.geomB}

			geomA, err := service.ParseGeometry(inputA)
			if err != nil {
				t.Fatalf("Failed to parse geometry A: %v", err)
			}

			geomB, err := service.ParseGeometry(inputB)
			if err != nil {
				t.Fatalf("Failed to parse geometry B: %v", err)
			}

			result, err := service.Within(geomA, geomB)
			if err != nil {
				t.Fatalf("Failed to test within: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected %t, got %t for %s within %s", tc.expected, result, tc.geomA, tc.geomB)
			}
		})
	}
}

// TestWithin_NilGeometry tests within with nil geometries
func TestWithin_NilGeometry(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	point := GeometryInput{WKT: "POINT(1.0 1.0)"}
	geom, err := service.ParseGeometry(point)
	if err != nil {
		t.Fatalf("Failed to parse geometry: %v", err)
	}

	// Test with nil first geometry
	_, err = service.Within(nil, geom)
	if err == nil {
		t.Error("Expected error for nil first geometry")
	}

	// Test with nil second geometry
	_, err = service.Within(geom, nil)
	if err == nil {
		t.Error("Expected error for nil second geometry")
	}

	// Test with both nil
	_, err = service.Within(nil, nil)
	if err == nil {
		t.Error("Expected error for both nil geometries")
	}
}

// TestIntersects tests spatial intersection
func TestIntersects(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geomA    string
		geomB    string
		expected bool
	}{
		{
			"Intersecting lines",
			"LINESTRING(0 0, 2 2)",
			"LINESTRING(0 2, 2 0)",
			true,
		},
		{
			"Non-intersecting lines",
			"LINESTRING(0 0, 1 1)",
			"LINESTRING(2 2, 3 3)",
			false,
		},
		{
			"Overlapping polygons",
			"POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			"POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))",
			true,
		},
		{
			"Touching polygons",
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"POLYGON((1 0, 2 0, 2 1, 1 1, 1 0))",
			true,
		},
		{
			"Non-touching polygons",
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"POLYGON((2 0, 3 0, 3 1, 2 1, 2 0))",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputA := GeometryInput{WKT: tc.geomA}
			inputB := GeometryInput{WKT: tc.geomB}

			geomA, err := service.ParseGeometry(inputA)
			if err != nil {
				t.Fatalf("Failed to parse geometry A: %v", err)
			}

			geomB, err := service.ParseGeometry(inputB)
			if err != nil {
				t.Fatalf("Failed to parse geometry B: %v", err)
			}

			result, err := service.Intersects(geomA, geomB)
			if err != nil {
				t.Fatalf("Failed to test intersects: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected %t, got %t for %s intersects %s", tc.expected, result, tc.geomA, tc.geomB)
			}
		})
	}
}

// TestDistance tests distance calculation
func TestDistance(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geomA    string
		geomB    string
		expected float64
		tolerance float64
	}{
		{
			"Distance between points",
			"POINT(0 0)",
			"POINT(3 4)",
			5.0,
			0.001,
		},
		{
			"Distance between identical points",
			"POINT(1 1)",
			"POINT(1 1)",
			0.0,
			0.001,
		},
		{
			"Distance between intersecting geometries",
			"POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			"POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))",
			0.0,
			0.001,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputA := GeometryInput{WKT: tc.geomA}
			inputB := GeometryInput{WKT: tc.geomB}

			geomA, err := service.ParseGeometry(inputA)
			if err != nil {
				t.Fatalf("Failed to parse geometry A: %v", err)
			}

			geomB, err := service.ParseGeometry(inputB)
			if err != nil {
				t.Fatalf("Failed to parse geometry B: %v", err)
			}

			result, err := service.Distance(geomA, geomB)
			if err != nil {
				t.Fatalf("Failed to calculate distance: %v", err)
			}

			if result < tc.expected-tc.tolerance || result > tc.expected+tc.tolerance {
				t.Errorf("Expected distance %.3f±%.3f, got %.3f", tc.expected, tc.tolerance, result)
			}
		})
	}
}

// TestBuffer tests buffer operations
func TestBuffer(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geom     string
		radius   float64
		expected bool
	}{
		{"Point buffer", "POINT(0 0)", 1.0, true},
		{"Line buffer", "LINESTRING(0 0, 1 1)", 0.5, true},
		{"Polygon buffer", "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", 0.1, true},
		{"Negative buffer", "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))", -0.1, true},
		{"Zero buffer", "POINT(0 0)", 0.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := GeometryInput{WKT: tc.geom}
			geom, err := service.ParseGeometry(input)
			if err != nil {
				t.Fatalf("Failed to parse geometry: %v", err)
			}

			buffered, err := service.Buffer(geom, tc.radius)
			if tc.expected {
				if err != nil {
					t.Errorf("Unexpected error creating buffer: %v", err)
				}
				if buffered == nil {
					t.Error("Expected non-nil buffered geometry")
				}
			} else {
				if err == nil {
					t.Error("Expected error creating buffer")
				}
			}
		})
	}
}

// TestBuffer_NilGeometry tests buffer with nil geometry
func TestBuffer_NilGeometry(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	_, err = service.Buffer(nil, 1.0)
	if err == nil {
		t.Error("Expected error for nil geometry")
	}
}

// TestSimplify tests geometry simplification
func TestSimplify(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name      string
		geom      string
		tolerance float64
		expected  bool
	}{
		{"Simple line", "LINESTRING(0 0, 1 0.1, 2 0.2, 3 0.1, 4 0)", 0.5, true},
		{"Complex polygon", "POLYGON((0 0, 0.5 0.1, 1 0, 1.5 0.1, 2 0, 2 1, 0 1, 0 0))", 0.2, true},
		{"Zero tolerance", "LINESTRING(0 0, 1 1, 2 2)", 0.0, true},
		{"High tolerance", "LINESTRING(0 0, 1 1, 2 2)", 10.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := GeometryInput{WKT: tc.geom}
			geom, err := service.ParseGeometry(input)
			if err != nil {
				t.Fatalf("Failed to parse geometry: %v", err)
			}

			simplified, err := service.Simplify(geom, tc.tolerance)
			if tc.expected {
				if err != nil {
					t.Errorf("Unexpected error simplifying geometry: %v", err)
				}
				if simplified == nil {
					t.Error("Expected non-nil simplified geometry")
				}
			} else {
				if err == nil {
					t.Error("Expected error simplifying geometry")
				}
			}
		})
	}
}

// TestUnion tests union operations
func TestUnion(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	// Test union of two polygons
	t.Run("Two polygons", func(t *testing.T) {
		poly1 := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
		poly2 := GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}

		geom1, err := service.ParseGeometry(poly1)
		if err != nil {
			t.Fatalf("Failed to parse polygon 1: %v", err)
		}

		geom2, err := service.ParseGeometry(poly2)
		if err != nil {
			t.Fatalf("Failed to parse polygon 2: %v", err)
		}

		union, err := service.Union([]*Geometry{geom1, geom2})
		if err != nil {
			t.Fatalf("Failed to create union: %v", err)
		}

		if union == nil {
			t.Error("Expected non-nil union geometry")
		}
	})

	// Test union of single geometry
	t.Run("Single geometry", func(t *testing.T) {
		poly := GeometryInput{WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"}
		geom, err := service.ParseGeometry(poly)
		if err != nil {
			t.Fatalf("Failed to parse polygon: %v", err)
		}

		union, err := service.Union([]*Geometry{geom})
		if err != nil {
			t.Fatalf("Failed to create union: %v", err)
		}

		if union != geom {
			t.Error("Expected union of single geometry to return the same geometry")
		}
	})

	// Test union of empty slice
	t.Run("Empty slice", func(t *testing.T) {
		_, err := service.Union([]*Geometry{})
		if err == nil {
			t.Error("Expected error for empty geometry slice")
		}
	})
}

// TestDifference tests difference operations
func TestDifference(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		geomA    string
		geomB    string
		expected bool
	}{
		{
			"Polygon difference",
			"POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))",
			"POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))",
			true,
		},
		{
			"Non-overlapping polygons",
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"POLYGON((2 2, 3 2, 3 3, 2 3, 2 2))",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputA := GeometryInput{WKT: tc.geomA}
			inputB := GeometryInput{WKT: tc.geomB}

			geomA, err := service.ParseGeometry(inputA)
			if err != nil {
				t.Fatalf("Failed to parse geometry A: %v", err)
			}

			geomB, err := service.ParseGeometry(inputB)
			if err != nil {
				t.Fatalf("Failed to parse geometry B: %v", err)
			}

			difference, err := service.Difference(geomA, geomB)
			if tc.expected {
				if err != nil {
					t.Errorf("Unexpected error creating difference: %v", err)
				}
				if difference == nil {
					t.Error("Expected non-nil difference geometry")
				}
			} else {
				if err == nil {
					t.Error("Expected error creating difference")
				}
			}
		})
	}
}

// TestValidateGeometry tests geometry validation
func TestValidateGeometry(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	testCases := []struct {
		name     string
		input    GeometryInput
		expected bool
	}{
		{"Valid WKT Point", GeometryInput{WKT: "POINT(1 2)"}, true},
		{"Valid WKT LineString", GeometryInput{WKT: "LINESTRING(0 0, 1 1)"}, true},
		{"Valid WKT Polygon", GeometryInput{WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"}, true},
		{"Invalid WKT", GeometryInput{WKT: "INVALID_GEOMETRY"}, false},
		{"Empty WKT", GeometryInput{WKT: ""}, false},
		{
			"Valid GeoJSON Point",
			GeometryInput{GeoJSON: map[string]interface{}{"type": "Point", "coordinates": []interface{}{1.0, 2.0}}},
			true,
		},
		{
			"Invalid GeoJSON - missing type",
			GeometryInput{GeoJSON: map[string]interface{}{"coordinates": []interface{}{1.0, 2.0}}},
			false,
		},
		{
			"Invalid GeoJSON - missing coordinates",
			GeometryInput{GeoJSON: map[string]interface{}{"type": "Point"}},
			false,
		},
		{"Empty input", GeometryInput{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateGeometry(tc.input)
			if tc.expected {
				if err != nil {
					t.Errorf("Expected valid geometry, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected validation error for invalid geometry")
				}
			}
		})
	}
}

// TestConcurrency tests thread safety
func TestConcurrency(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}
	defer service.Close()

	// Test concurrent operations
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Perform various operations
			point := GeometryInput{WKT: "POINT(1 1)"}
			polygon := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}

			pointGeom, err := service.ParseGeometry(point)
			if err != nil {
				errors <- err
				return
			}

			polygonGeom, err := service.ParseGeometry(polygon)
			if err != nil {
				errors <- err
				return
			}

			_, err = service.Within(pointGeom, polygonGeom)
			if err != nil {
				errors <- err
				return
			}

			_, err = service.Distance(pointGeom, polygonGeom)
			if err != nil {
				errors <- err
				return
			}

			_, err = service.Buffer(pointGeom, 1.0)
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

// TestServiceClosedContext tests operations on closed service
func TestServiceClosedContext(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create GEOS service: %v", err)
	}

	// Close the service
	service.Close()

	// Try to use the service after closing
	input := GeometryInput{WKT: "POINT(1 1)"}
	_, err = service.ParseGeometry(input)
	if err == nil {
		t.Error("Expected error when using closed service")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Expected 'not initialized' error, got: %v", err)
	}
}