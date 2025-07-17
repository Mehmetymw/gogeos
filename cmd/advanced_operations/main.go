package main

import (
	"fmt"
	"log"

	"github.com/yourusername/gogeos/geos"
)

func main() {
	// Create a new GEOS service
	service, err := geos.NewService()
	if err != nil {
		log.Fatal("Failed to create GEOS service:", err)
	}
	defer service.Close()

	// Example 1: Union of multiple polygons
	fmt.Println("=== Union of Multiple Polygons ===")
	poly1 := geos.GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
	poly2 := geos.GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}
	poly3 := geos.GeometryInput{WKT: "POLYGON((2 2, 4 2, 4 4, 2 4, 2 2))"}

	geom1, err := service.ParseGeometry(poly1)
	if err != nil {
		log.Fatal("Failed to parse poly1:", err)
	}

	geom2, err := service.ParseGeometry(poly2)
	if err != nil {
		log.Fatal("Failed to parse poly2:", err)
	}

	geom3, err := service.ParseGeometry(poly3)
	if err != nil {
		log.Fatal("Failed to parse poly3:", err)
	}

	union, err := service.Union([]*geos.Geometry{geom1, geom2, geom3})
	if err != nil {
		log.Fatal("Failed to create union:", err)
	}

	unionWKT, err := service.ToWKT(union)
	if err != nil {
		log.Fatal("Failed to convert union to WKT:", err)
	}
	fmt.Printf("Union of 3 polygons: %s\n", unionWKT)

	// Example 2: Difference operation - creating holes
	fmt.Println("\n=== Difference Operation ===")
	largePolygon := geos.GeometryInput{WKT: "POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))"}
	smallPolygon := geos.GeometryInput{WKT: "POLYGON((2 2, 8 2, 8 8, 2 8, 2 2))"}

	largeGeom, err := service.ParseGeometry(largePolygon)
	if err != nil {
		log.Fatal("Failed to parse large polygon:", err)
	}

	smallGeom, err := service.ParseGeometry(smallPolygon)
	if err != nil {
		log.Fatal("Failed to parse small polygon:", err)
	}

	difference, err := service.Difference(largeGeom, smallGeom)
	if err != nil {
		log.Fatal("Failed to create difference:", err)
	}

	differenceWKT, err := service.ToWKT(difference)
	if err != nil {
		log.Fatal("Failed to convert difference to WKT:", err)
	}
	fmt.Printf("Difference (large - small): %s\n", differenceWKT)

	// Example 3: Complex buffer operations
	fmt.Println("\n=== Complex Buffer Operations ===")
	lineString := geos.GeometryInput{WKT: "LINESTRING(0 0, 5 0, 5 5, 10 5)"}
	lineGeom, err := service.ParseGeometry(lineString)
	if err != nil {
		log.Fatal("Failed to parse line:", err)
	}

	// Create different buffer sizes
	bufferSizes := []float64{0.5, 1.0, 2.0}
	for _, size := range bufferSizes {
		buffered, err := service.Buffer(lineGeom, size)
		if err != nil {
			log.Fatal("Failed to create buffer:", err)
		}

		bufferedWKT, err := service.ToWKT(buffered)
		if err != nil {
			log.Fatal("Failed to convert buffer to WKT:", err)
		}
		fmt.Printf("Buffer size %.1f: %s\n", size, bufferedWKT)
	}

	// Example 4: Geometry validation
	fmt.Println("\n=== Geometry Validation ===")
	validGeometries := []geos.GeometryInput{
		{WKT: "POINT(1 2)"},
		{WKT: "LINESTRING(0 0, 1 1)"},
		{WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"},
	}

	invalidGeometries := []geos.GeometryInput{
		{WKT: "POINT(1)"},                 // Missing coordinate
		{WKT: "LINESTRING(0 0)"},          // Single point in linestring
		{WKT: "POLYGON((0 0, 1 0, 1 1))"}, // Unclosed polygon
		{WKT: "INVALID_GEOMETRY"},         // Invalid geometry type
	}

	fmt.Println("Valid geometries:")
	for i, geom := range validGeometries {
		err := service.ValidateGeometry(geom)
		if err != nil {
			fmt.Printf("  %d. %s - INVALID: %v\n", i+1, geom.WKT, err)
		} else {
			fmt.Printf("  %d. %s - VALID\n", i+1, geom.WKT)
		}
	}

	fmt.Println("Invalid geometries:")
	for i, geom := range invalidGeometries {
		err := service.ValidateGeometry(geom)
		if err != nil {
			fmt.Printf("  %d. %s - INVALID: %v\n", i+1, geom.WKT, err)
		} else {
			fmt.Printf("  %d. %s - VALID\n", i+1, geom.WKT)
		}
	}

	// Example 5: Working with complex GeoJSON
	fmt.Println("\n=== Complex GeoJSON Operations ===")
	geoJSONPolygon := geos.GeometryInput{
		GeoJSON: map[string]interface{}{
			"type": "Polygon",
			"coordinates": [][][]float64{
				{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}}, // Outer ring
			},
		},
	}

	geoJSONLineString := geos.GeometryInput{
		GeoJSON: map[string]interface{}{
			"type": "LineString",
			"coordinates": [][]float64{
				{1, 1}, {2, 2}, {3, 1}, {4, 2},
			},
		},
	}

	geoJSONPolyGeom, err := service.ParseGeometry(geoJSONPolygon)
	if err != nil {
		log.Fatal("Failed to parse GeoJSON polygon:", err)
	}

	geoJSONLineGeom, err := service.ParseGeometry(geoJSONLineString)
	if err != nil {
		log.Fatal("Failed to parse GeoJSON linestring:", err)
	}

	// Test if line intersects polygon
	intersects, err := service.Intersects(geoJSONLineGeom, geoJSONPolyGeom)
	if err != nil {
		log.Fatal("Failed to test intersection:", err)
	}
	fmt.Printf("GeoJSON line intersects polygon: %t\n", intersects)

	// Calculate distance from line to polygon
	distance, err := service.Distance(geoJSONLineGeom, geoJSONPolyGeom)
	if err != nil {
		log.Fatal("Failed to calculate distance:", err)
	}
	fmt.Printf("Distance from line to polygon: %.6f\n", distance)

	// Example 6: Simplification with different tolerances
	fmt.Println("\n=== Simplification with Different Tolerances ===")
	detailedLine := geos.GeometryInput{WKT: "LINESTRING(0 0, 0.1 0.1, 0.2 0.05, 0.3 0.15, 0.4 0.08, 0.5 0.12, 0.6 0.06, 0.7 0.14, 0.8 0.09, 0.9 0.11, 1.0 0.1)"}
	detailedGeom, err := service.ParseGeometry(detailedLine)
	if err != nil {
		log.Fatal("Failed to parse detailed line:", err)
	}

	tolerances := []float64{0.01, 0.05, 0.1, 0.2}
	for _, tolerance := range tolerances {
		simplified, err := service.Simplify(detailedGeom, tolerance)
		if err != nil {
			log.Fatal("Failed to simplify:", err)
		}

		simplifiedWKT, err := service.ToWKT(simplified)
		if err != nil {
			log.Fatal("Failed to convert simplified to WKT:", err)
		}
		fmt.Printf("Tolerance %.2f: %s\n", tolerance, simplifiedWKT)
	}

	fmt.Println("\n=== All advanced examples completed successfully! ===")
}
