package geos

import (
	"testing"
)

// TestIntegration_RealWorldScenario tests a real-world GIS scenario
func TestIntegration_RealWorldScenario(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Scenario: Finding buildings within a service area buffer
	// 1. Create a service center point
	serviceCenter := helper.ParseWKT("POINT(100 100)")

	// 2. Create a service area (buffer around the center)
	serviceArea := helper.AssertBuffer(serviceCenter, 50.0)

	// 3. Create some buildings
	building1 := helper.ParseWKT("POLYGON((90 90, 95 90, 95 95, 90 95, 90 90))")
	building2 := helper.ParseWKT("POLYGON((110 110, 115 110, 115 115, 110 115, 110 110))")
	building3 := helper.ParseWKT("POLYGON((200 200, 205 200, 205 205, 200 205, 200 200))")

	// 4. Test which buildings are within the service area
	helper.AssertWithin(building1, serviceArea, true)  // Should be within
	helper.AssertWithin(building2, serviceArea, true)  // Should be within
	helper.AssertWithin(building3, serviceArea, false) // Should be outside

	// 5. Calculate distances from service center to buildings
	helper.AssertDistance(serviceCenter, building1, 14.142, 0.01)  // ~sqrt(200)
	helper.AssertDistance(serviceCenter, building2, 14.142, 0.01)  // ~sqrt(200)
	helper.AssertDistance(serviceCenter, building3, 141.421, 0.01) // ~sqrt(20000)

	// 6. Create a union of all buildings within service area
	buildingsInArea := []*Geometry{building1, building2}
	unionBuildings := helper.AssertUnion(buildingsInArea)

	// 7. Calculate the area not covered by buildings (difference)
	uncoveredArea := helper.AssertDifference(serviceArea, unionBuildings)

	// 8. Validate all geometries are properly formed
	serviceAreaWKT := helper.AssertToWKT(serviceArea)
	uncoveredAreaWKT := helper.AssertToWKT(uncoveredArea)

	if serviceAreaWKT == "" || uncoveredAreaWKT == "" {
		t.Error("Expected non-empty WKT strings")
	}
}

// TestIntegration_SpatialAnalysis tests complex spatial analysis operations
func TestIntegration_SpatialAnalysis(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Create a complex polygon representing a city boundary
	cityBoundary := helper.ParseWKT("POLYGON((0 0, 100 0, 100 100, 0 100, 0 0))")

	// Create roads as linestrings
	road1 := helper.ParseWKT("LINESTRING(0 50, 100 50)")
	road2 := helper.ParseWKT("LINESTRING(50 0, 50 100)")
	road3 := helper.ParseWKT("LINESTRING(0 0, 100 100)")

	// Create buffers around roads to represent road corridors
	roadBuffer1 := helper.AssertBuffer(road1, 5.0)
	roadBuffer2 := helper.AssertBuffer(road2, 5.0)
	roadBuffer3 := helper.AssertBuffer(road3, 3.0)

	// Test road intersections
	helper.AssertIntersects(road1, road2, true)
	helper.AssertIntersects(road1, road3, true)
	helper.AssertIntersects(road2, road3, true)

	// Create union of all road corridors
	roadCorridors := []*Geometry{roadBuffer1, roadBuffer2, roadBuffer3}
	allRoads := helper.AssertUnion(roadCorridors)

	// Calculate developable area (city boundary minus roads)
	developableArea := helper.AssertDifference(cityBoundary, allRoads)

	// Test that developable area is within city boundary
	helper.AssertWithin(developableArea, cityBoundary, true)

	// Test that roads intersect with city boundary
	helper.AssertIntersects(allRoads, cityBoundary, true)

	// Verify all results are valid geometries
	developableWKT := helper.AssertToWKT(developableArea)
	roadsWKT := helper.AssertToWKT(allRoads)

	if len(developableWKT) == 0 || len(roadsWKT) == 0 {
		t.Error("Expected non-empty WKT strings for all geometries")
	}
}

// TestIntegration_GeometrySimplification tests geometry simplification workflow
func TestIntegration_GeometrySimplification(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Create a complex line with many vertices
	complexLine := helper.ParseWKT("LINESTRING(0 0, 0.1 0.01, 0.2 0.02, 0.3 0.01, 0.4 0.03, 0.5 0.02, 0.6 0.01, 0.7 0.04, 0.8 0.02, 0.9 0.01, 1.0 0.0)")

	// Test different simplification tolerances
	tolerances := []float64{0.005, 0.01, 0.02, 0.05}
	
	for _, tolerance := range tolerances {
		simplified := helper.AssertSimplify(complexLine, tolerance)
		
		// Verify simplified geometry is valid
		simplifiedWKT := helper.AssertToWKT(simplified)
		if len(simplifiedWKT) == 0 {
			t.Errorf("Expected non-empty WKT for tolerance %f", tolerance)
		}
		
		// Verify simplified geometry intersects with original
		helper.AssertIntersects(complexLine, simplified, true)
	}
}

// TestIntegration_GeoJSONWorkflow tests complete GeoJSON workflow
func TestIntegration_GeoJSONWorkflow(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Create geometries from GeoJSON
	pointGeom := helper.ParseGeoJSON(helper.PointGeoJSON())
	lineGeom := helper.ParseGeoJSON(helper.LineGeoJSON())
	polygonGeom := helper.ParseGeoJSON(helper.PolygonGeoJSON())

	// Test spatial relationships
	helper.AssertWithin(pointGeom, polygonGeom, true)
	helper.AssertIntersects(lineGeom, polygonGeom, true)

	// Convert back to WKT
	pointWKT := helper.AssertToWKT(pointGeom)
	lineWKT := helper.AssertToWKT(lineGeom)
	polygonWKT := helper.AssertToWKT(polygonGeom)

	if len(pointWKT) == 0 || len(lineWKT) == 0 || len(polygonWKT) == 0 {
		t.Error("Expected non-empty WKT strings from GeoJSON conversion")
	}

	// Test buffer operations on GeoJSON-derived geometries
	pointBuffer := helper.AssertBuffer(pointGeom, 1.0)
	lineBuffer := helper.AssertBuffer(lineGeom, 0.5)

	// Test union of buffered geometries
	bufferedGeoms := []*Geometry{pointBuffer, lineBuffer}
	unionBuffered := helper.AssertUnion(bufferedGeoms)

	// Verify union intersects with original polygon
	helper.AssertIntersects(unionBuffered, polygonGeom, true)
}

// TestIntegration_ErrorRecovery tests error handling and recovery
func TestIntegration_ErrorRecovery(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Test with invalid geometries
	invalidInputs := []GeometryInput{
		{WKT: "POINT(1)"},                    // Missing coordinate
		{WKT: "LINESTRING(0 0)"},             // Single point
		{WKT: "POLYGON((0 0, 1 0, 1 1))"},    // Unclosed polygon
		{WKT: "INVALID_GEOMETRY"},            // Invalid type
		{},                                   // Empty input
	}

	for _, input := range invalidInputs {
		helper.AssertValidateGeometry(input, false)
	}

	// Test that service continues to work after errors
	validGeom := helper.ParseWKT("POINT(1 1)")
	validWKT := helper.AssertToWKT(validGeom)
	
	if len(validWKT) == 0 {
		t.Error("Service should continue working after handling invalid inputs")
	}
}

// TestIntegration_Performance tests performance with complex operations
func TestIntegration_Performance(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Create a complex polygon with many vertices
	complexPolygon := helper.ParseWKT("POLYGON((0 0, 10 0, 20 5, 30 0, 40 0, 50 5, 60 0, 70 0, 80 5, 90 0, 100 0, 100 10, 95 20, 100 30, 100 40, 95 50, 100 60, 100 70, 95 80, 100 90, 100 100, 90 100, 80 95, 70 100, 60 100, 50 95, 40 100, 30 100, 20 95, 10 100, 0 100, 0 90, 5 80, 0 70, 0 60, 5 50, 0 40, 0 30, 5 20, 0 10, 0 0))")

	// Test multiple operations on the complex polygon
	buffer1 := helper.AssertBuffer(complexPolygon, 5.0)
	buffer2 := helper.AssertBuffer(complexPolygon, -2.0)
	
	simplified := helper.AssertSimplify(complexPolygon, 3.0)
	
	// Test union and difference operations
	union := helper.AssertUnion([]*Geometry{buffer1, simplified})
	difference := helper.AssertDifference(buffer1, buffer2)
	
	// Verify all results are valid
	results := []*Geometry{buffer1, buffer2, simplified, union, difference}
	for i, geom := range results {
		wkt := helper.AssertToWKT(geom)
		if len(wkt) == 0 {
			t.Errorf("Expected non-empty WKT for result %d", i)
		}
	}
}

// TestIntegration_MemoryManagement tests memory management with many operations
func TestIntegration_MemoryManagement(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Create many geometries and perform operations
	const numOperations = 100
	
	baseGeom := helper.ParseWKT("POINT(0 0)")
	
	for i := 0; i < numOperations; i++ {
		// Create a buffer
		buffered := helper.AssertBuffer(baseGeom, float64(i+1))
		
		// Convert to WKT
		wkt := helper.AssertToWKT(buffered)
		
		// Verify it's valid
		if len(wkt) == 0 {
			t.Errorf("Expected non-empty WKT for operation %d", i)
		}
		
		// Create a new geometry from the WKT
		newGeom := helper.ParseWKT(wkt)
		
		// Test distance (should be 0 for identical geometries)
		helper.AssertDistance(buffered, newGeom, 0.0, 0.001)
	}
}

// TestIntegration_ConcurrentOperations tests concurrent access patterns
func TestIntegration_ConcurrentOperations(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Close()

	// Test concurrent geometry creation and operations
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Each goroutine performs a series of operations
			point := helper.ParseWKT("POINT(1 1)")
			polygon := helper.ParseWKT("POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))")
			
			// Test spatial relationships
			helper.AssertWithin(point, polygon, true)
			helper.AssertIntersects(point, polygon, true)
			
			// Test geometric operations
			pointBuffer := helper.AssertBuffer(point, 0.5)
			polygonSimplified := helper.AssertSimplify(polygon, 0.1)
			
			// Test union
			union := helper.AssertUnion([]*Geometry{pointBuffer, polygonSimplified})
			
			// Convert results to WKT
			unionWKT := helper.AssertToWKT(union)
			
			if len(unionWKT) == 0 {
				errors <- &testError{message: "Expected non-empty WKT"}
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

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}