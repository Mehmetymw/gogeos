package geos

import (
	"testing"
)

// BenchmarkNewService benchmarks service creation
func BenchmarkNewService(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service, err := NewService()
		if err != nil {
			b.Fatal(err)
		}
		service.Close()
	}
}

// BenchmarkParseGeometry_WKT benchmarks WKT parsing
func BenchmarkParseGeometry_WKT(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	input := GeometryInput{WKT: "POINT(1.0 2.0)"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ParseGeometry(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseGeometry_GeoJSON benchmarks GeoJSON parsing
func BenchmarkParseGeometry_GeoJSON(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	input := GeometryInput{
		GeoJSON: map[string]interface{}{
			"type":        "Point",
			"coordinates": []interface{}{1.0, 2.0},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ParseGeometry(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWithin benchmarks spatial within operations
func BenchmarkWithin(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	pointInput := GeometryInput{WKT: "POINT(1.0 1.0)"}
	polygonInput := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}

	pointGeom, err := service.ParseGeometry(pointInput)
	if err != nil {
		b.Fatal(err)
	}

	polygonGeom, err := service.ParseGeometry(polygonInput)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Within(pointGeom, polygonGeom)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIntersects benchmarks spatial intersection operations
func BenchmarkIntersects(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	line1Input := GeometryInput{WKT: "LINESTRING(0 0, 2 2)"}
	line2Input := GeometryInput{WKT: "LINESTRING(0 2, 2 0)"}

	line1Geom, err := service.ParseGeometry(line1Input)
	if err != nil {
		b.Fatal(err)
	}

	line2Geom, err := service.ParseGeometry(line2Input)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Intersects(line1Geom, line2Geom)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDistance benchmarks distance calculations
func BenchmarkDistance(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	point1Input := GeometryInput{WKT: "POINT(0 0)"}
	point2Input := GeometryInput{WKT: "POINT(3 4)"}

	point1Geom, err := service.ParseGeometry(point1Input)
	if err != nil {
		b.Fatal(err)
	}

	point2Geom, err := service.ParseGeometry(point2Input)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Distance(point1Geom, point2Geom)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuffer benchmarks buffer operations
func BenchmarkBuffer(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	pointInput := GeometryInput{WKT: "POINT(0 0)"}
	pointGeom, err := service.ParseGeometry(pointInput)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Buffer(pointGeom, 1.0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimplify benchmarks simplification operations
func BenchmarkSimplify(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	lineInput := GeometryInput{WKT: "LINESTRING(0 0, 1 0.1, 2 0.2, 3 0.1, 4 0)"}
	lineGeom, err := service.ParseGeometry(lineInput)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Simplify(lineGeom, 0.1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToWKT benchmarks WKT conversion
func BenchmarkToWKT(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	pointInput := GeometryInput{WKT: "POINT(1.0 2.0)"}
	pointGeom, err := service.ParseGeometry(pointInput)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ToWKT(pointGeom)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnion benchmarks union operations
func BenchmarkUnion(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	poly1Input := GeometryInput{WKT: "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))"}
	poly2Input := GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}

	poly1Geom, err := service.ParseGeometry(poly1Input)
	if err != nil {
		b.Fatal(err)
	}

	poly2Geom, err := service.ParseGeometry(poly2Input)
	if err != nil {
		b.Fatal(err)
	}

	geometries := []*Geometry{poly1Geom, poly2Geom}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Union(geometries)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDifference benchmarks difference operations
func BenchmarkDifference(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	poly1Input := GeometryInput{WKT: "POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))"}
	poly2Input := GeometryInput{WKT: "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))"}

	poly1Geom, err := service.ParseGeometry(poly1Input)
	if err != nil {
		b.Fatal(err)
	}

	poly2Geom, err := service.ParseGeometry(poly2Input)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Difference(poly1Geom, poly2Geom)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateGeometry benchmarks geometry validation
func BenchmarkValidateGeometry(b *testing.B) {
	service, err := NewService()
	if err != nil {
		b.Fatal(err)
	}
	defer service.Close()

	input := GeometryInput{WKT: "POINT(1.0 2.0)"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.ValidateGeometry(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}