// Package geos provides a Go wrapper for the GEOS (Geometry Engine Open Source) library.
// This package offers thread-safe geometric operations including spatial relationships,
// geometric transformations, and topological operations.
//
// GEOS is a C++ library for performing geometric operations on planar geometries.
// This package provides a safe, idiomatic Go interface to GEOS functionality with
// automatic memory management and thread safety.
//
// Key features:
//   - Thread-safe operations through service instances
//   - Automatic memory management with finalizers
//   - Support for WKT and GeoJSON input formats
//   - Comprehensive geometric operations (intersections, unions, buffers, etc.)
//   - Spatial relationship testing (within, intersects, distance)
//
// Basic usage:
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"github.com/yourusername/gogeos/geos"
//	)
//
//	func main() {
//		// Create a new GEOS service
//		service, err := geos.NewService()
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer service.Close()
//
//		// Parse geometries
//		point := geos.GeometryInput{WKT: "POINT(1.0 1.0)"}
//		polygon := geos.GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
//
//		pointGeom, err := service.ParseGeometry(point)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		polygonGeom, err := service.ParseGeometry(polygon)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Test spatial relationship
//		within, err := service.Within(pointGeom, polygonGeom)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Printf("Point is within polygon: %t\n", within)
//	}
//
// Requirements:
//   - GEOS C library must be installed and accessible via pkg-config
//   - CGO must be enabled for compilation
package geos

/*
#cgo pkg-config: geos
#include <geos_c.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Service provides GEOS-based geometric operations with thread safety.
// It wraps the GEOS C library context and ensures all operations are thread-safe
// using read-write mutexes. Each Service instance manages its own GEOS context
// and should be properly closed when no longer needed to prevent memory leaks.
//
// Example usage:
//
//	service, err := geos.NewService()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer service.Close()
//
//	// Use service for geometric operations...
type Service struct {
	context C.GEOSContextHandle_t
	mutex   sync.RWMutex
}

// NewService creates a new GEOS service with proper initialization.
// It initializes the GEOS context and sets up automatic cleanup using finalizers.
// The returned service is thread-safe and ready for geometric operations.
//
// Returns:
//   - *Service: A configured GEOS service instance
//   - error: An error if GEOS context initialization fails
//
// Example:
//
//	service, err := geos.NewService()
//	if err != nil {
//		return fmt.Errorf("failed to create GEOS service: %w", err)
//	}
//	defer service.Close()
func NewService() (*Service, error) {
	// Initialize GEOS context
	ctx := C.GEOS_init_r()
	if ctx == nil {
		return nil, errors.New("failed to initialize GEOS context")
	}

	service := &Service{
		context: ctx,
	}

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(service, (*Service).Close)

	return service, nil
}

// Close cleans up GEOS resources safely.
// This method should be called when the service is no longer needed to prevent
// memory leaks. It's safe to call multiple times and is automatically called
// by the finalizer if forgotten.
//
// Example:
//
//	service, err := geos.NewService()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer service.Close() // Ensure cleanup
func (s *Service) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.context != nil {
		C.GEOS_finish_r(s.context)
		s.context = nil
	}
	runtime.SetFinalizer(s, nil)
}

// Geometry represents a spatial geometry with automatic cleanup.
// It wraps a GEOS geometry object and maintains a reference to the service
// that created it to ensure proper cleanup. Geometry objects are automatically
// cleaned up when garbage collected, but can also be explicitly destroyed.
//
// Note: Geometry objects are not thread-safe on their own, but operations
// on them through the Service are thread-safe.
type Geometry struct {
	geom    *C.struct_GEOSGeom_t
	service *Service
}

// GeometryInput represents input geometry data that can be either WKT or GeoJSON format.
// Only one of WKT or GeoJSON should be provided. The SRID field is optional and
// currently not used in processing but reserved for future spatial reference system support.
//
// Supported GeoJSON types: Point, LineString, Polygon
// Supported WKT types: All standard OGC WKT geometry types
//
// Example WKT input:
//
//	input := GeometryInput{
//		WKT: "POINT(1.0 2.0)",
//	}
//
// Example GeoJSON input:
//
//	input := GeometryInput{
//		GeoJSON: map[string]interface{}{
//			"type": "Point",
//			"coordinates": []float64{1.0, 2.0},
//		},
//	}
type GeometryInput struct {
	WKT     string                 `json:"wkt,omitempty"`
	GeoJSON map[string]interface{} `json:"geojson,omitempty"`
	SRID    int                    `json:"srid,omitempty"`
}

// newGeometry creates a new geometry with cleanup
func (s *Service) newGeometry(geom *C.struct_GEOSGeom_t) *Geometry {
	if geom == nil {
		return nil
	}

	g := &Geometry{
		geom:    geom,
		service: s,
	}

	// Set finalizer for automatic cleanup
	runtime.SetFinalizer(g, (*Geometry).destroy)
	return g
}

// destroy cleans up geometry resources
func (g *Geometry) destroy() {
	if g.geom != nil && g.service != nil && g.service.context != nil {
		g.service.mutex.RLock()
		if g.service.context != nil {
			C.GEOSGeom_destroy_r(g.service.context, g.geom)
		}
		g.service.mutex.RUnlock()
		g.geom = nil
	}
	runtime.SetFinalizer(g, nil)
}

// ParseGeometry parses WKT or GeoJSON input into a GEOS geometry object.
// It supports both WKT strings and GeoJSON objects, with validation for
// proper format and geometric validity.
//
// Parameters:
//   - input: GeometryInput containing either WKT string or GeoJSON object
//
// Returns:
//   - *Geometry: A parsed and validated geometry object
//   - error: An error if parsing fails or geometry is invalid
//
// Supported formats:
//   - WKT: Well-Known Text format (e.g., "POINT(1.0 2.0)")
//   - GeoJSON: Point, LineString, and Polygon geometries
//
// Example:
//
//	input := GeometryInput{WKT: "POINT(1.0 2.0)"}
//	geom, err := service.ParseGeometry(input)
//	if err != nil {
//		return fmt.Errorf("failed to parse geometry: %w", err)
//	}
func (s *Service) ParseGeometry(input GeometryInput) (*Geometry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return nil, errors.New("GEOS context is not initialized")
	}

	var wkt string

	if input.WKT != "" {
		wkt = input.WKT
	} else if input.GeoJSON != nil {
		// Convert GeoJSON to WKT (improved implementation)
		geoType, ok := input.GeoJSON["type"].(string)
		if !ok {
			return nil, errors.New("invalid GeoJSON: missing type")
		}

		coords, ok := input.GeoJSON["coordinates"]
		if !ok {
			return nil, errors.New("invalid GeoJSON: missing coordinates")
		}

		var err error
		wkt, err = s.geoJSONToWKT(geoType, coords)
		if err != nil {
			return nil, fmt.Errorf("failed to convert GeoJSON to WKT: %v", err)
		}
	} else {
		return nil, errors.New("no geometry provided: either WKT or GeoJSON is required")
	}

	// Validate WKT format before parsing
	if len(wkt) == 0 {
		return nil, errors.New("empty WKT string")
	}

	// Create C string safely
	cWKT := C.CString(wkt)
	defer C.free(unsafe.Pointer(cWKT))

	// Parse geometry with error checking
	geom := C.GEOSGeomFromWKT_r(s.context, cWKT)
	if geom == nil {
		return nil, fmt.Errorf("failed to parse WKT geometry: %s", wkt)
	}

	// Validate the parsed geometry
	if C.GEOSisValid_r(s.context, geom) == 0 {
		C.GEOSGeom_destroy_r(s.context, geom)
		return nil, fmt.Errorf("invalid geometry: %s", wkt)
	}

	return s.newGeometry(geom), nil
}

// geoJSONToWKT converts GeoJSON coordinates to WKT (improved implementation)
func (s *Service) geoJSONToWKT(geoType string, coords interface{}) (string, error) {
	switch geoType {
	case "Point":
		if coordArray, ok := coords.([]interface{}); ok && len(coordArray) >= 2 {
			x, okX := coordArray[0].(float64)
			y, okY := coordArray[1].(float64)
			if okX && okY {
				return fmt.Sprintf("POINT(%f %f)", x, y), nil
			}
		}
		return "", errors.New("invalid Point coordinates")

	case "Polygon":
		if rings, ok := coords.([]interface{}); ok && len(rings) > 0 {
			if ring, ok := rings[0].([]interface{}); ok && len(ring) >= 4 {
				wkt := "POLYGON(("
				for i, coord := range ring {
					if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
						x, okX := coordArray[0].(float64)
						y, okY := coordArray[1].(float64)
						if okX && okY {
							if i > 0 {
								wkt += ", "
							}
							wkt += fmt.Sprintf("%f %f", x, y)
						}
					}
				}
				wkt += "))"
				return wkt, nil
			}
		}
		return "", errors.New("invalid Polygon coordinates")

	case "LineString":
		if coords, ok := coords.([]interface{}); ok && len(coords) >= 2 {
			wkt := "LINESTRING("
			for i, coord := range coords {
				if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
					x, okX := coordArray[0].(float64)
					y, okY := coordArray[1].(float64)
					if okX && okY {
						if i > 0 {
							wkt += ", "
						}
						wkt += fmt.Sprintf("%f %f", x, y)
					}
				}
			}
			wkt += ")"
			return wkt, nil
		}
		return "", errors.New("invalid LineString coordinates")
	}

	return "", fmt.Errorf("unsupported GeoJSON type: %s", geoType)
}

// ToWKT converts a geometry object to its Well-Known Text (WKT) representation.
// This is useful for serializing geometries for storage or transmission.
//
// Parameters:
//   - geom: The geometry object to convert
//
// Returns:
//   - string: The WKT representation of the geometry
//   - error: An error if conversion fails
//
// Example:
//
//	input := GeometryInput{WKT: "POINT(1.0 2.0)"}
//	geom, err := service.ParseGeometry(input)
//	if err != nil {
//		log.Fatal(err)
//	}
//	wkt, err := service.ToWKT(geom)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(wkt) // Output: POINT (1.0000000000000000 2.0000000000000000)
func (s *Service) ToWKT(geom *Geometry) (string, error) {
	if geom == nil || geom.geom == nil {
		return "", errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return "", errors.New("GEOS context is not initialized")
	}

	cWKT := C.GEOSGeomToWKT_r(s.context, geom.geom)
	if cWKT == nil {
		return "", errors.New("failed to convert geometry to WKT")
	}
	defer C.free(unsafe.Pointer(cWKT))

	return C.GoString(cWKT), nil
}

// Within tests whether geometry A is completely within geometry B.
// This is a spatial relationship test that returns true if every point of A
// is inside B and the interiors of A and B have at least one point in common.
//
// Parameters:
//   - a: The geometry to test if it's within B
//   - b: The geometry to test against
//
// Returns:
//   - bool: True if A is within B, false otherwise
//   - error: An error if the operation fails
//
// Example:
//
//	point := GeometryInput{WKT: "POINT(1.0 1.0)"}
//	polygon := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
//
//	pointGeom, _ := service.ParseGeometry(point)
//	polygonGeom, _ := service.ParseGeometry(polygon)
//
//	isWithin, err := service.Within(pointGeom, polygonGeom)
//	// isWithin will be true since the point is inside the polygon
func (s *Service) Within(a, b *Geometry) (bool, error) {
	if a == nil || b == nil || a.geom == nil || b.geom == nil {
		return false, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return false, errors.New("GEOS context is not initialized")
	}

	result := C.GEOSWithin_r(s.context, a.geom, b.geom)
	if result == 2 { // Error case
		return false, errors.New("GEOS within operation failed")
	}

	return result == 1, nil
}

// Intersects tests whether two geometries spatially intersect.
// Returns true if the geometries have any points in common, including
// touching boundaries or overlapping areas.
//
// Parameters:
//   - a: First geometry to test
//   - b: Second geometry to test
//
// Returns:
//   - bool: True if the geometries intersect, false otherwise
//   - error: An error if the operation fails
//
// Example:
//
//	line1 := GeometryInput{WKT: "LINESTRING(0 0, 2 2)"}
//	line2 := GeometryInput{WKT: "LINESTRING(0 2, 2 0)"}
//
//	geom1, _ := service.ParseGeometry(line1)
//	geom2, _ := service.ParseGeometry(line2)
//
//	intersects, err := service.Intersects(geom1, geom2)
//	// intersects will be true since the lines cross each other
func (s *Service) Intersects(a, b *Geometry) (bool, error) {
	if a == nil || b == nil || a.geom == nil || b.geom == nil {
		return false, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return false, errors.New("GEOS context is not initialized")
	}

	result := C.GEOSIntersects_r(s.context, a.geom, b.geom)
	if result == 2 { // Error case
		return false, errors.New("GEOS intersects operation failed")
	}

	return result == 1, nil
}

// Distance calculates the minimum distance between two geometries.
// For intersecting geometries, the distance is 0. For non-intersecting
// geometries, it returns the shortest distance between any two points
// of the geometries.
//
// Parameters:
//   - a: First geometry
//   - b: Second geometry
//
// Returns:
//   - float64: The minimum distance between the geometries
//   - error: An error if the operation fails
//
// Example:
//
//	point1 := GeometryInput{WKT: "POINT(0 0)"}
//	point2 := GeometryInput{WKT: "POINT(3 4)"}
//
//	geom1, _ := service.ParseGeometry(point1)
//	geom2, _ := service.ParseGeometry(point2)
//
//	distance, err := service.Distance(geom1, geom2)
//	// distance will be 5.0 (Euclidean distance)
func (s *Service) Distance(a, b *Geometry) (float64, error) {
	if a == nil || b == nil || a.geom == nil || b.geom == nil {
		return 0, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return 0, errors.New("GEOS context is not initialized")
	}

	var distance C.double
	result := C.GEOSDistance_r(s.context, a.geom, b.geom, &distance)
	if result == 0 {
		return 0, errors.New("failed to calculate distance")
	}

	return float64(distance), nil
}

// Buffer creates a buffer zone around a geometry at the specified distance.
// The buffer operation creates a new geometry that includes all points within
// the specified distance from the original geometry.
//
// Parameters:
//   - geom: The geometry to buffer
//   - radius: The buffer distance (positive for expansion, negative for contraction)
//
// Returns:
//   - *Geometry: A new geometry representing the buffer zone
//   - error: An error if the operation fails
//
// Example:
//
//	point := GeometryInput{WKT: "POINT(0 0)"}
//	geom, _ := service.ParseGeometry(point)
//
//	buffered, err := service.Buffer(geom, 1.0)
//	// buffered will be a circular polygon with radius 1.0 around the point
func (s *Service) Buffer(geom *Geometry, radius float64) (*Geometry, error) {
	if geom == nil || geom.geom == nil {
		return nil, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return nil, errors.New("GEOS context is not initialized")
	}

	buffered := C.GEOSBuffer_r(s.context, geom.geom, C.double(radius), 8)
	if buffered == nil {
		return nil, errors.New("failed to create buffer")
	}

	return s.newGeometry(buffered), nil
}

// Simplify simplifies a geometry using the Douglas-Peucker algorithm.
// This reduces the number of vertices in the geometry while preserving
// its overall shape within the specified tolerance.
//
// Parameters:
//   - geom: The geometry to simplify
//   - tolerance: The maximum distance a vertex can be from the simplified geometry
//
// Returns:
//   - *Geometry: A new simplified geometry
//   - error: An error if the operation fails
//
// Example:
//
//	complex := GeometryInput{WKT: "LINESTRING(0 0, 1 0.1, 2 0.2, 3 0.1, 4 0)"}
//	geom, _ := service.ParseGeometry(complex)
//
//	simplified, err := service.Simplify(geom, 0.5)
//	// simplified will have fewer vertices while maintaining the general shape
func (s *Service) Simplify(geom *Geometry, tolerance float64) (*Geometry, error) {
	if geom == nil || geom.geom == nil {
		return nil, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return nil, errors.New("GEOS context is not initialized")
	}

	simplified := C.GEOSSimplify_r(s.context, geom.geom, C.double(tolerance))
	if simplified == nil {
		return nil, errors.New("failed to simplify geometry")
	}

	return s.newGeometry(simplified), nil
}

// Union creates a union of multiple geometries.
// The union operation combines all input geometries into a single geometry
// that represents the set of all points that are in any of the input geometries.
//
// Parameters:
//   - geometries: A slice of geometries to union together
//
// Returns:
//   - *Geometry: A new geometry representing the union of all input geometries
//   - error: An error if the operation fails
//
// Example:
//
//	poly1 := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
//	poly2 := GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}
//
//	geom1, _ := service.ParseGeometry(poly1)
//	geom2, _ := service.ParseGeometry(poly2)
//
//	union, err := service.Union([]*Geometry{geom1, geom2})
//	// union will be a single polygon covering the combined area
func (s *Service) Union(geometries []*Geometry) (*Geometry, error) {
	if len(geometries) == 0 {
		return nil, errors.New("no geometries provided")
	}

	if len(geometries) == 1 {
		return geometries[0], nil
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return nil, errors.New("GEOS context is not initialized")
	}

	result := geometries[0]
	for i := 1; i < len(geometries); i++ {
		if geometries[i] == nil || geometries[i].geom == nil {
			continue
		}

		union := C.GEOSUnion_r(s.context, result.geom, geometries[i].geom)
		if union == nil {
			return nil, errors.New("failed to create union")
		}

		result = s.newGeometry(union)
	}

	return result, nil
}

// Difference creates the geometric difference between two geometries.
// The difference operation returns a geometry that represents the part of
// geometry A that is not in geometry B (A - B).
//
// Parameters:
//   - a: The geometry to subtract from
//   - b: The geometry to subtract
//
// Returns:
//   - *Geometry: A new geometry representing A - B
//   - error: An error if the operation fails
//
// Example:
//
//	poly1 := GeometryInput{WKT: "POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))"}
//	poly2 := GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}
//	
//	geom1, _ := service.ParseGeometry(poly1)
//	geom2, _ := service.ParseGeometry(poly2)
//	
//	difference, err := service.Difference(geom1, geom2)
//	// difference will be poly1 with a hole where poly2 was
func (s *Service) Difference(a, b *Geometry) (*Geometry, error) {
	if a == nil || b == nil || a.geom == nil || b.geom == nil {
		return nil, errors.New("invalid geometry")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.context == nil {
		return nil, errors.New("GEOS context is not initialized")
	}

	diff := C.GEOSDifference_r(s.context, a.geom, b.geom)
	if diff == nil {
		return nil, errors.New("failed to create difference")
	}

	return s.newGeometry(diff), nil
}

// ValidateGeometry validates input geometry format without full parsing.
// This is a lightweight validation that checks the basic structure and format
// of WKT or GeoJSON input without creating actual geometry objects.
//
// Parameters:
//   - input: The geometry input to validate
//
// Returns:
//   - error: An error if the input format is invalid, nil if valid
//
// Example:
//
//	input := GeometryInput{WKT: "POINT(1.0 2.0)"}
//	err := service.ValidateGeometry(input)
//	if err != nil {
//		log.Printf("Invalid geometry format: %v", err)
//	}
func (s *Service) ValidateGeometry(input GeometryInput) error {
	if input.WKT == "" && input.GeoJSON == nil {
		return errors.New("no geometry provided: either WKT or GeoJSON is required")
	}

	if input.WKT != "" {
		// Basic WKT validation
		if len(input.WKT) == 0 {
			return errors.New("empty WKT string")
		}
		// Check for basic WKT keywords
		wkt := input.WKT
		validTypes := []string{"POINT", "LINESTRING", "POLYGON", "MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION"}
		isValid := false
		for _, validType := range validTypes {
			if len(wkt) >= len(validType) && wkt[:len(validType)] == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("invalid WKT: must start with a valid geometry type")
		}
	}

	if input.GeoJSON != nil {
		// Basic GeoJSON validation
		if _, ok := input.GeoJSON["type"]; !ok {
			return errors.New("invalid GeoJSON: missing type field")
		}
		if _, ok := input.GeoJSON["coordinates"]; !ok {
			return errors.New("invalid GeoJSON: missing coordinates field")
		}
	}

	return nil
}
