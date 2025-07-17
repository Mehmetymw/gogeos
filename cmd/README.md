# GoGEOS Examples

This directory contains example command-line programs demonstrating various features of the GoGEOS library.

## Running the Examples

Before running the examples, ensure you have:

1. GEOS library installed on your system
2. Go 1.18 or higher
3. CGO enabled (default)

```bash
# Run basic usage example
go run ./cmd/basic_usage

# Run advanced operations example
go run ./cmd/advanced_operations

# Or use the Makefile
make run-basic
make run-advanced

# Build examples as binaries
make examples
./bin/basic_usage
./bin/advanced_operations
```

## Example Programs

### `basic_usage/main.go`
Demonstrates fundamental operations including:
- Point in polygon testing
- Distance calculations
- Buffer operations
- Line intersections
- Working with GeoJSON input
- Geometry simplification

### `advanced_operations/main.go`
Shows more complex operations including:
- Union of multiple polygons
- Difference operations (creating holes)
- Complex buffer operations with different sizes
- Geometry validation
- Complex GeoJSON operations
- Simplification with different tolerances

## Key Concepts Demonstrated

### 1. Service Management
```go
service, err := geos.NewService()
if err != nil {
    log.Fatal(err)
}
defer service.Close() // Always close to prevent memory leaks
```

### 2. Geometry Input Formats
```go
// WKT format
wktInput := geos.GeometryInput{WKT: "POINT(1.0 2.0)"}

// GeoJSON format
geoJSONInput := geos.GeometryInput{
    GeoJSON: map[string]interface{}{
        "type": "Point",
        "coordinates": []float64{1.0, 2.0},
    },
}
```

### 3. Error Handling
All operations return errors that should be checked:
```go
geom, err := service.ParseGeometry(input)
if err != nil {
    log.Fatal("Failed to parse geometry:", err)
}
```

### 4. Memory Management
- Geometries are automatically cleaned up by finalizers
- Always call `service.Close()` when done
- The library handles GEOS C library memory management

## Expected Output

When you run the examples, you should see detailed output showing:
- Test results for spatial relationships
- Distance calculations
- WKT representations of transformed geometries
- Validation results for various geometry inputs
- Simplified geometries at different tolerance levels

## Troubleshooting

If you encounter issues:

1. **GEOS not found**: Ensure GEOS development headers are installed
2. **CGO errors**: Make sure CGO is enabled (`CGO_ENABLED=1`)
3. **pkg-config errors**: Verify pkg-config can find GEOS (`pkg-config --cflags --libs geos`)

## Next Steps

After running these examples, you can:
- Experiment with different geometry types
- Try different buffer sizes and tolerances
- Combine multiple operations
- Integrate with your own spatial data processing workflows