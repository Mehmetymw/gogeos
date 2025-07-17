package main

import (
	"fmt"
	"log"

	"github.com/mehmetymw/gogeos/geos"
)

func main() {
	// Create a new GEOS service
	service, err := geos.NewService()
	if err != nil {
		log.Fatal("Failed to create GEOS service:", err)
	}
	defer service.Close()

	// Example 1: Basic point in polygon test
	fmt.Println("=== Point in Polygon Test ===")
	pointInput := geos.GeometryInput{WKT: "POINT(1.0 1.0)"}
	polygonInput := geos.GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}

	pointGeom, err := service.ParseGeometry(pointInput)
	if err != nil {
		log.Fatal("Failed to parse point:", err)
	}

	polygonGeom, err := service.ParseGeometry(polygonInput)
	if err != nil {
		log.Fatal("Failed to parse polygon:", err)
	}

	within, err := service.Within(pointGeom, polygonGeom)
	if err != nil {
		log.Fatal("Failed to test within:", err)
	}
	fmt.Printf("Point (1,1) is within polygon: %t\n", within)

	// Example 2: Distance calculation
	fmt.Println("\n=== Distance Calculation ===")
	point1 := geos.GeometryInput{WKT: "POINT(0 0)"}
	point2 := geos.GeometryInput{WKT: "POINT(3 4)"}

	geom1, err := service.ParseGeometry(point1)
	if err != nil {
		log.Fatal("Failed to parse point1:", err)
	}

	geom2, err := service.ParseGeometry(point2)
	if err != nil {
		log.Fatal("Failed to parse point2:", err)
	}

	distance, err := service.Distance(geom1, geom2)
	if err != nil {
		log.Fatal("Failed to calculate distance:", err)
	}
	fmt.Printf("Distance between (0,0) and (3,4): %.2f\n", distance)

	// Example 3: Buffer operation
	fmt.Println("\n=== Buffer Operation ===")
	pointForBuffer := geos.GeometryInput{WKT: "POINT(0 0)"}
	pointGeom2, err := service.ParseGeometry(pointForBuffer)
	if err != nil {
		log.Fatal("Failed to parse point for buffer:", err)
	}

	buffered, err := service.Buffer(pointGeom2, 1.0)
	if err != nil {
		log.Fatal("Failed to create buffer:", err)
	}

	bufferedWKT, err := service.ToWKT(buffered)
	if err != nil {
		log.Fatal("Failed to convert buffer to WKT:", err)
	}
	fmt.Printf("Buffered point (radius 1.0): %s\n", bufferedWKT)

	// Example 4: Line intersection
	fmt.Println("\n=== Line Intersection ===")
	line1 := geos.GeometryInput{WKT: "LINESTRING(0 0, 2 2)"}
	line2 := geos.GeometryInput{WKT: "LINESTRING(0 2, 2 0)"}

	lineGeom1, err := service.ParseGeometry(line1)
	if err != nil {
		log.Fatal("Failed to parse line1:", err)
	}

	lineGeom2, err := service.ParseGeometry(line2)
	if err != nil {
		log.Fatal("Failed to parse line2:", err)
	}

	intersects, err := service.Intersects(lineGeom1, lineGeom2)
	if err != nil {
		log.Fatal("Failed to test intersection:", err)
	}
	fmt.Printf("Lines intersect: %t\n", intersects)

	// Example 5: Working with GeoJSON
	fmt.Println("\n=== GeoJSON Example ===")
	geoJSONPoint := geos.GeometryInput{
		GeoJSON: map[string]interface{}{
			"type":        "Point",
			"coordinates": []float64{1.5, 1.5},
		},
	}

	geoJSONGeom, err := service.ParseGeometry(geoJSONPoint)
	if err != nil {
		log.Fatal("Failed to parse GeoJSON:", err)
	}

	// Test if GeoJSON point is within our polygon
	withinFromGeoJSON, err := service.Within(geoJSONGeom, polygonGeom)
	if err != nil {
		log.Fatal("Failed to test GeoJSON within:", err)
	}
	fmt.Printf("GeoJSON point (1.5,1.5) is within polygon: %t\n", withinFromGeoJSON)

	// Convert back to WKT
	wkt, err := service.ToWKT(geoJSONGeom)
	if err != nil {
		log.Fatal("Failed to convert to WKT:", err)
	}
	fmt.Printf("GeoJSON point as WKT: %s\n", wkt)

	// Example 6: Geometry simplification
	fmt.Println("\n=== Geometry Simplification ===")
	complexLine := geos.GeometryInput{WKT: "LINESTRING(0 0, 0.5 0.1, 1.0 0.2, 1.5 0.1, 2.0 0.0, 2.5 0.1, 3.0 0.0)"}
	complexGeom, err := service.ParseGeometry(complexLine)
	if err != nil {
		log.Fatal("Failed to parse complex line:", err)
	}

	simplified, err := service.Simplify(complexGeom, 0.5)
	if err != nil {
		log.Fatal("Failed to simplify:", err)
	}

	simplifiedWKT, err := service.ToWKT(simplified)
	if err != nil {
		log.Fatal("Failed to convert simplified to WKT:", err)
	}
	fmt.Printf("Simplified line: %s\n", simplifiedWKT)

	fmt.Println("\n=== All examples completed successfully! ===")
}
