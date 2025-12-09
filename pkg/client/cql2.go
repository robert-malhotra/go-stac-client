// Package client provides a CQL2-JSON filter query builder for STAC API searches.
//
// This implementation wraps the github.com/planetlabs/go-ogc/filter package
// to provide a convenient, fluent API for building CQL2-JSON filter expressions.
//
// Example usage:
//
//	f := client.NewFilterBuilder().
//	    And(client.Lt(client.Property("eo:cloud_cover"), client.Number(10))).
//	    And(client.SIntersects(
//	        client.Property("geometry"),
//	        client.BBox(-122.5, 37.5, -122.0, 38.0),
//	    )).
//	    Build()
package client

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/planetlabs/go-ogc/filter"
)

// -----------------------------------------------------------------------------
// Re-exports from go-ogc/filter for convenience
// -----------------------------------------------------------------------------

// Filter is the top-level CQL2 filter type from go-ogc.
type Filter = filter.Filter

// Expression types
type (
	BooleanExpression  = filter.BooleanExpression
	ScalarExpression   = filter.ScalarExpression
	SpatialExpression  = filter.SpatialExpression
	TemporalExpression = filter.TemporalExpression
	NumericExpression  = filter.NumericExpression
)

// -----------------------------------------------------------------------------
// Property References
// -----------------------------------------------------------------------------

// Property creates a property reference expression.
// Common STAC properties include:
//   - "datetime" - acquisition datetime
//   - "eo:cloud_cover" - cloud cover percentage
//   - "geometry" - item geometry
//   - "id" - item ID
//   - "collection" - collection ID
func Property(name string) *filter.Property {
	return &filter.Property{Name: name}
}

// -----------------------------------------------------------------------------
// Literal Values
// -----------------------------------------------------------------------------

// String creates a string literal.
func String(s string) *filter.String {
	return &filter.String{Value: s}
}

// Number creates a numeric literal.
func Number(n float64) *filter.Number {
	return &filter.Number{Value: n}
}

// Boolean creates a boolean literal.
func Boolean(b bool) *filter.Boolean {
	return &filter.Boolean{Value: b}
}

// -----------------------------------------------------------------------------
// Comparison Operators
// -----------------------------------------------------------------------------

// Eq creates an equality comparison (=).
func Eq(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.Equals,
		Left:  left,
		Right: right,
	}
}

// Neq creates an inequality comparison (<>).
func Neq(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.NotEquals,
		Left:  left,
		Right: right,
	}
}

// Lt creates a less-than comparison (<).
func Lt(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.LessThan,
		Left:  left,
		Right: right,
	}
}

// Lte creates a less-than-or-equal comparison (<=).
func Lte(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.LessThanOrEquals,
		Left:  left,
		Right: right,
	}
}

// Gt creates a greater-than comparison (>).
func Gt(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.GreaterThan,
		Left:  left,
		Right: right,
	}
}

// Gte creates a greater-than-or-equal comparison (>=).
func Gte(left, right filter.ScalarExpression) *filter.Comparison {
	return &filter.Comparison{
		Name:  filter.GreaterThanOrEquals,
		Left:  left,
		Right: right,
	}
}

// Like creates a pattern matching expression.
// Use % for multi-character wildcard and _ for single character wildcard.
func Like(value filter.CharacterExpression, pattern filter.PatternExpression) *filter.Like {
	return &filter.Like{
		Value:   value,
		Pattern: pattern,
	}
}

// Between creates a range comparison (value BETWEEN low AND high).
func Between(value, low, high filter.NumericExpression) *filter.Between {
	return &filter.Between{
		Value: value,
		Low:   low,
		High:  high,
	}
}

// In creates a membership test (value IN list).
func In(item filter.ScalarExpression, list ...filter.ScalarExpression) *filter.In {
	return &filter.In{
		Item: item,
		List: list,
	}
}

// IsNull creates a null check (value IS NULL).
func IsNull(value filter.Expression) *filter.IsNull {
	return &filter.IsNull{
		Value: value,
	}
}

// -----------------------------------------------------------------------------
// Logical Operators
// -----------------------------------------------------------------------------

// And creates a logical AND of multiple expressions.
func And(exprs ...filter.BooleanExpression) *filter.And {
	return &filter.And{Args: exprs}
}

// Or creates a logical OR of multiple expressions.
func Or(exprs ...filter.BooleanExpression) *filter.Or {
	return &filter.Or{Args: exprs}
}

// Not creates a logical NOT of an expression.
func Not(expr filter.BooleanExpression) *filter.Not {
	return &filter.Not{Arg: expr}
}

// -----------------------------------------------------------------------------
// Spatial Types & Operators
// -----------------------------------------------------------------------------

// Geometry converts an orb.Geometry to a filter.Geometry for use in spatial operations.
// This is the primary way to use orb geometries with the CQL2 filter builder.
//
// Example:
//
//	pt := orb.Point{-122.4194, 37.7749}
//	filter := SIntersects(Property("geometry"), Geometry(pt))
func Geometry(g orb.Geometry) *filter.Geometry {
	gj := geojson.NewGeometry(g)
	return &filter.Geometry{Value: gj}
}

// GeometryFromGeoJSON creates a filter.Geometry from a raw GeoJSON map.
// Use this when you have GeoJSON data that's not in orb format.
func GeometryFromGeoJSON(gjson map[string]any) *filter.Geometry {
	return &filter.Geometry{Value: gjson}
}

// Point creates a GeoJSON Point geometry from longitude and latitude.
func Point(lon, lat float64) *filter.Geometry {
	return Geometry(orb.Point{lon, lat})
}

// Point3D creates a GeoJSON Point geometry with elevation.
// Note: orb.Point only supports 2D, so elevation is stored in coordinates array.
func Point3D(lon, lat, elevation float64) *filter.Geometry {
	return &filter.Geometry{
		Value: map[string]any{
			"type":        "Point",
			"coordinates": []float64{lon, lat, elevation},
		},
	}
}

// LineString creates a GeoJSON LineString geometry from coordinate pairs.
// Each coordinate is [lon, lat].
func LineString(coords ...[]float64) *filter.Geometry {
	ls := make(orb.LineString, len(coords))
	for i, c := range coords {
		if len(c) >= 2 {
			ls[i] = orb.Point{c[0], c[1]}
		}
	}
	return Geometry(ls)
}

// LineStringFromOrb creates a filter.Geometry from an orb.LineString.
func LineStringFromOrb(ls orb.LineString) *filter.Geometry {
	return Geometry(ls)
}

// Polygon creates a GeoJSON Polygon geometry from rings.
// The first ring is the exterior ring, subsequent rings are holes.
// Each ring is a slice of [lon, lat] coordinates.
func Polygon(rings ...[][]float64) *filter.Geometry {
	poly := make(orb.Polygon, len(rings))
	for i, ring := range rings {
		r := make(orb.Ring, len(ring))
		for j, c := range ring {
			if len(c) >= 2 {
				r[j] = orb.Point{c[0], c[1]}
			}
		}
		poly[i] = r
	}
	return Geometry(poly)
}

// PolygonFromOrb creates a filter.Geometry from an orb.Polygon.
func PolygonFromOrb(poly orb.Polygon) *filter.Geometry {
	return Geometry(poly)
}

// MultiPoint creates a GeoJSON MultiPoint geometry from coordinate pairs.
func MultiPoint(coords ...[]float64) *filter.Geometry {
	mp := make(orb.MultiPoint, len(coords))
	for i, c := range coords {
		if len(c) >= 2 {
			mp[i] = orb.Point{c[0], c[1]}
		}
	}
	return Geometry(mp)
}

// MultiPointFromOrb creates a filter.Geometry from an orb.MultiPoint.
func MultiPointFromOrb(mp orb.MultiPoint) *filter.Geometry {
	return Geometry(mp)
}

// MultiLineString creates a GeoJSON MultiLineString geometry.
func MultiLineString(lines ...[][]float64) *filter.Geometry {
	mls := make(orb.MultiLineString, len(lines))
	for i, line := range lines {
		ls := make(orb.LineString, len(line))
		for j, c := range line {
			if len(c) >= 2 {
				ls[j] = orb.Point{c[0], c[1]}
			}
		}
		mls[i] = ls
	}
	return Geometry(mls)
}

// MultiLineStringFromOrb creates a filter.Geometry from an orb.MultiLineString.
func MultiLineStringFromOrb(mls orb.MultiLineString) *filter.Geometry {
	return Geometry(mls)
}

// MultiPolygon creates a GeoJSON MultiPolygon geometry.
func MultiPolygon(polygons ...[][][]float64) *filter.Geometry {
	mpoly := make(orb.MultiPolygon, len(polygons))
	for i, poly := range polygons {
		p := make(orb.Polygon, len(poly))
		for j, ring := range poly {
			r := make(orb.Ring, len(ring))
			for k, c := range ring {
				if len(c) >= 2 {
					r[k] = orb.Point{c[0], c[1]}
				}
			}
			p[j] = r
		}
		mpoly[i] = p
	}
	return Geometry(mpoly)
}

// MultiPolygonFromOrb creates a filter.Geometry from an orb.MultiPolygon.
func MultiPolygonFromOrb(mpoly orb.MultiPolygon) *filter.Geometry {
	return Geometry(mpoly)
}

// GeometryCollection creates a GeoJSON GeometryCollection from multiple orb geometries.
func GeometryCollection(geometries ...orb.Geometry) *filter.Geometry {
	gc := make(orb.Collection, len(geometries))
	copy(gc, geometries)
	return Geometry(gc)
}

// GeometryCollectionFromOrb creates a filter.Geometry from an orb.Collection.
func GeometryCollectionFromOrb(gc orb.Collection) *filter.Geometry {
	return Geometry(gc)
}

// BBox creates a 2D bounding box expression.
// Order: minLon, minLat, maxLon, maxLat
func BBox(minLon, minLat, maxLon, maxLat float64) *filter.BoundingBox {
	return &filter.BoundingBox{
		Extent: []float64{minLon, minLat, maxLon, maxLat},
	}
}

// BBox3D creates a 3D bounding box expression.
// Order: minLon, minLat, minElev, maxLon, maxLat, maxElev
func BBox3D(minLon, minLat, minElev, maxLon, maxLat, maxElev float64) *filter.BoundingBox {
	return &filter.BoundingBox{
		Extent: []float64{minLon, minLat, minElev, maxLon, maxLat, maxElev},
	}
}

// toSpatialExpression converts various geometry types to filter.SpatialExpression.
// Accepts: orb.Geometry, *filter.Geometry, *filter.BoundingBox, orb.Bound
func toSpatialExpression(geom any) filter.SpatialExpression {
	switch g := geom.(type) {
	case filter.SpatialExpression:
		return g
	case orb.Bound:
		// Check orb.Bound before orb.Geometry since Bound implements Geometry
		return &filter.BoundingBox{
			Extent: []float64{g.Min.X(), g.Min.Y(), g.Max.X(), g.Max.Y()},
		}
	case orb.Geometry:
		return Geometry(g)
	default:
		// Return nil for unsupported types - will cause runtime error if used
		return nil
	}
}

// SIntersects creates a spatial intersection test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SIntersects(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryIntersects,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SEquals creates a spatial equality test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SEquals(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryEquals,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SDisjoint creates a spatial disjoint test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SDisjoint(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryDisjoint,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// STouches creates a spatial touches test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func STouches(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryTouches,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SWithin creates a spatial within test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SWithin(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryWithin,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SOverlaps creates a spatial overlaps test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SOverlaps(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryOverlaps,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SCrosses creates a spatial crosses test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SCrosses(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryCrosses,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// SContains creates a spatial contains test against the "geometry" property.
// Accepts orb.Geometry, orb.Bound, *filter.Geometry, or *filter.BoundingBox.
func SContains(geom any) *filter.SpatialComparison {
	return &filter.SpatialComparison{
		Name:  filter.GeometryContains,
		Left:  Property("geometry"),
		Right: toSpatialExpression(geom),
	}
}

// -----------------------------------------------------------------------------
// Temporal Types & Operators
// -----------------------------------------------------------------------------

// Timestamp creates a timestamp expression from an ISO 8601 string.
func Timestamp(iso8601 string) *filter.Timestamp {
	t, _ := time.Parse(time.RFC3339, iso8601)
	return &filter.Timestamp{Value: t}
}

// TimestampFromTime creates a timestamp expression from a time.Time.
func TimestampFromTime(t time.Time) *filter.Timestamp {
	return &filter.Timestamp{Value: t.UTC()}
}

// Date creates a date expression from a date string (YYYY-MM-DD).
func Date(dateStr string) *filter.Date {
	t, _ := time.Parse(time.DateOnly, dateStr)
	return &filter.Date{Value: t}
}

// DateFromTime creates a date expression from a time.Time.
func DateFromTime(t time.Time) *filter.Date {
	return &filter.Date{Value: t}
}

// Interval creates a time interval expression from start and end timestamps.
func Interval(start, end filter.InstantExpression) *filter.Interval {
	return &filter.Interval{
		Start: start,
		End:   end,
	}
}

// IntervalFromStrings creates a time interval from ISO 8601 strings.
// Use empty string for open-ended intervals.
func IntervalFromStrings(start, end string) *filter.Interval {
	var startExpr, endExpr filter.InstantExpression
	if start != "" && start != ".." {
		startExpr = Timestamp(start)
	}
	if end != "" && end != ".." {
		endExpr = Timestamp(end)
	}
	return &filter.Interval{
		Start: startExpr,
		End:   endExpr,
	}
}

// IntervalFromTimes creates a time interval from time.Time values.
func IntervalFromTimes(start, end time.Time) *filter.Interval {
	return &filter.Interval{
		Start: TimestampFromTime(start),
		End:   TimestampFromTime(end),
	}
}

// OpenIntervalBefore creates an open-ended interval up to the given time.
func OpenIntervalBefore(end string) *filter.Interval {
	return &filter.Interval{
		Start: nil,
		End:   Timestamp(end),
	}
}

// OpenIntervalAfter creates an open-ended interval from the given time.
func OpenIntervalAfter(start string) *filter.Interval {
	return &filter.Interval{
		Start: Timestamp(start),
		End:   nil,
	}
}

// TAfter creates a temporal "after" test.
func TAfter(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeAfter,
		Left:  left,
		Right: right,
	}
}

// TBefore creates a temporal "before" test.
func TBefore(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeBefore,
		Left:  left,
		Right: right,
	}
}

// TContains creates a temporal "contains" test.
func TContains(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeContains,
		Left:  left,
		Right: right,
	}
}

// TDisjoint creates a temporal "disjoint" test.
func TDisjoint(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeDisjoint,
		Left:  left,
		Right: right,
	}
}

// TDuring creates a temporal "during" test.
func TDuring(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeDuring,
		Left:  left,
		Right: right,
	}
}

// TEquals creates a temporal "equals" test.
func TEquals(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeEquals,
		Left:  left,
		Right: right,
	}
}

// TFinishedBy creates a temporal "finished by" test.
func TFinishedBy(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeFinishedBy,
		Left:  left,
		Right: right,
	}
}

// TFinishes creates a temporal "finishes" test.
func TFinishes(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeFinishes,
		Left:  left,
		Right: right,
	}
}

// TIntersects creates a temporal "intersects" test.
func TIntersects(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeIntersects,
		Left:  left,
		Right: right,
	}
}

// TMeets creates a temporal "meets" test.
func TMeets(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeMeets,
		Left:  left,
		Right: right,
	}
}

// TMetBy creates a temporal "met by" test.
func TMetBy(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeMetBy,
		Left:  left,
		Right: right,
	}
}

// TOverlappedBy creates a temporal "overlapped by" test.
func TOverlappedBy(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeOverlappedBy,
		Left:  left,
		Right: right,
	}
}

// TOverlaps creates a temporal "overlaps" test.
func TOverlaps(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeOverlaps,
		Left:  left,
		Right: right,
	}
}

// TStartedBy creates a temporal "started by" test.
func TStartedBy(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeStartedBy,
		Left:  left,
		Right: right,
	}
}

// TStarts creates a temporal "starts" test.
func TStarts(left, right filter.TemporalExpression) *filter.TemporalComparison {
	return &filter.TemporalComparison{
		Name:  filter.TimeStarts,
		Left:  left,
		Right: right,
	}
}

// -----------------------------------------------------------------------------
// Array Operators
// -----------------------------------------------------------------------------

// Array creates an array literal from items.
func Array(items ...filter.ArrayItemExpression) filter.Array {
	return filter.Array(items)
}

// AEquals tests if two arrays are equal.
func AEquals(left, right filter.ArrayExpression) *filter.ArrayComparison {
	return &filter.ArrayComparison{
		Name:  filter.ArrayEquals,
		Left:  left,
		Right: right,
	}
}

// AContains tests if the first array contains all elements of the second.
func AContains(left, right filter.ArrayExpression) *filter.ArrayComparison {
	return &filter.ArrayComparison{
		Name:  filter.ArrayContains,
		Left:  left,
		Right: right,
	}
}

// AContainedBy tests if all elements of the first array are in the second.
func AContainedBy(left, right filter.ArrayExpression) *filter.ArrayComparison {
	return &filter.ArrayComparison{
		Name:  filter.ArrayContainedBy,
		Left:  left,
		Right: right,
	}
}

// AOverlaps tests if two arrays have at least one common element.
func AOverlaps(left, right filter.ArrayExpression) *filter.ArrayComparison {
	return &filter.ArrayComparison{
		Name:  filter.ArrayOverlaps,
		Left:  left,
		Right: right,
	}
}

// -----------------------------------------------------------------------------
// Filter Builder
// -----------------------------------------------------------------------------

// FilterBuilder provides a fluent interface for building CQL2 filters.
type FilterBuilder struct {
	exprs []filter.BooleanExpression
}

// NewFilterBuilder creates a new FilterBuilder.
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

// Where sets the initial filter expression.
func (b *FilterBuilder) Where(expr filter.BooleanExpression) *FilterBuilder {
	b.exprs = []filter.BooleanExpression{expr}
	return b
}

// And adds an AND condition to the existing filter.
func (b *FilterBuilder) And(expr filter.BooleanExpression) *FilterBuilder {
	b.exprs = append(b.exprs, expr)
	return b
}

// Or creates an OR branch with the given expressions.
func (b *FilterBuilder) Or(exprs ...filter.BooleanExpression) *FilterBuilder {
	b.exprs = append(b.exprs, Or(exprs...))
	return b
}

// Build returns the Filter that can be used in search requests.
func (b *FilterBuilder) Build() *filter.Filter {
	if len(b.exprs) == 0 {
		return nil
	}
	if len(b.exprs) == 1 {
		return &filter.Filter{Expression: b.exprs[0]}
	}
	return &filter.Filter{Expression: And(b.exprs...)}
}
