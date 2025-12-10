package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/paulmach/orb"
	"github.com/planetlabs/go-ogc/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteralExpressions(t *testing.T) {
	t.Run("string literal", func(t *testing.T) {
		s := String("hello")
		assert.Equal(t, "hello", s.Value)
	})

	t.Run("number literal", func(t *testing.T) {
		n := Number(42.5)
		assert.Equal(t, 42.5, n.Value)
	})

	t.Run("boolean literal", func(t *testing.T) {
		b := Boolean(true)
		assert.True(t, b.Value)
	})
}

func TestPropertyExpression(t *testing.T) {
	prop := Property("eo:cloud_cover")
	assert.Equal(t, "eo:cloud_cover", prop.Name)
}

func TestComparisonOperatorsErgonomic(t *testing.T) {
	t.Run("equals with string", func(t *testing.T) {
		cmp := Eq("status", "published")
		assert.Equal(t, filter.Equals, cmp.Name)
		prop, ok := cmp.Left.(*filter.Property)
		require.True(t, ok)
		assert.Equal(t, "status", prop.Name)
		str, ok := cmp.Right.(*filter.String)
		require.True(t, ok)
		assert.Equal(t, "published", str.Value)
	})

	t.Run("equals with int", func(t *testing.T) {
		cmp := Eq("count", 42)
		assert.Equal(t, filter.Equals, cmp.Name)
		num, ok := cmp.Right.(*filter.Number)
		require.True(t, ok)
		assert.Equal(t, 42.0, num.Value)
	})

	t.Run("equals with float64", func(t *testing.T) {
		cmp := Eq("score", 98.5)
		num, ok := cmp.Right.(*filter.Number)
		require.True(t, ok)
		assert.Equal(t, 98.5, num.Value)
	})

	t.Run("equals with bool", func(t *testing.T) {
		cmp := Eq("enabled", true)
		b, ok := cmp.Right.(*filter.Boolean)
		require.True(t, ok)
		assert.True(t, b.Value)
	})

	t.Run("not equals", func(t *testing.T) {
		cmp := Neq("status", "draft")
		assert.Equal(t, filter.NotEquals, cmp.Name)
	})

	t.Run("less than", func(t *testing.T) {
		cmp := Lt("eo:cloud_cover", 10.0)
		assert.Equal(t, filter.LessThan, cmp.Name)
	})

	t.Run("less than or equal", func(t *testing.T) {
		cmp := Lte("count", 100)
		assert.Equal(t, filter.LessThanOrEquals, cmp.Name)
	})

	t.Run("greater than", func(t *testing.T) {
		cmp := Gt("elevation", 1000.0)
		assert.Equal(t, filter.GreaterThan, cmp.Name)
	})

	t.Run("greater than or equal", func(t *testing.T) {
		cmp := Gte("quality", 80)
		assert.Equal(t, filter.GreaterThanOrEquals, cmp.Name)
	})

	t.Run("between", func(t *testing.T) {
		b := Between("temp", -10.0, 30.0)
		assert.NotNil(t, b.Value)
		assert.NotNil(t, b.Low)
		assert.NotNil(t, b.High)
	})

	t.Run("in with strings", func(t *testing.T) {
		in := In("collection", "a", "b", "c")
		assert.NotNil(t, in.Item)
		assert.Len(t, in.List, 3)
	})

	t.Run("in with ints", func(t *testing.T) {
		in := In("quality", 1, 2, 3)
		assert.Len(t, in.List, 3)
		num, ok := in.List[0].(*filter.Number)
		require.True(t, ok)
		assert.Equal(t, 1.0, num.Value)
	})

	t.Run("isNull", func(t *testing.T) {
		isN := IsNull("optional_field")
		assert.NotNil(t, isN.Value)
	})

	t.Run("like", func(t *testing.T) {
		l := Like("id", "S2A_%")
		assert.NotNil(t, l.Value)
		assert.NotNil(t, l.Pattern)
	})
}

func TestComparisonOperatorsExplicit(t *testing.T) {
	t.Run("equals", func(t *testing.T) {
		cmp := EqExpr(Property("status"), String("published"))
		assert.Equal(t, filter.Equals, cmp.Name)
	})

	t.Run("not equals", func(t *testing.T) {
		cmp := NeqExpr(Property("status"), String("draft"))
		assert.Equal(t, filter.NotEquals, cmp.Name)
	})

	t.Run("less than", func(t *testing.T) {
		cmp := LtExpr(Property("eo:cloud_cover"), Number(10.0))
		assert.Equal(t, filter.LessThan, cmp.Name)
	})

	t.Run("less than or equal", func(t *testing.T) {
		cmp := LteExpr(Property("count"), Number(100))
		assert.Equal(t, filter.LessThanOrEquals, cmp.Name)
	})

	t.Run("greater than", func(t *testing.T) {
		cmp := GtExpr(Property("elevation"), Number(1000.0))
		assert.Equal(t, filter.GreaterThan, cmp.Name)
	})

	t.Run("greater than or equal", func(t *testing.T) {
		cmp := GteExpr(Property("quality"), Number(80))
		assert.Equal(t, filter.GreaterThanOrEquals, cmp.Name)
	})

	t.Run("between", func(t *testing.T) {
		b := BetweenExpr(Property("temp"), Number(-10.0), Number(30.0))
		assert.NotNil(t, b.Value)
		assert.NotNil(t, b.Low)
		assert.NotNil(t, b.High)
	})

	t.Run("in", func(t *testing.T) {
		in := InExpr(Property("collection"), String("a"), String("b"), String("c"))
		assert.NotNil(t, in.Item)
		assert.Len(t, in.List, 3)
	})

	t.Run("isNull", func(t *testing.T) {
		isN := IsNullExpr(Property("optional_field"))
		assert.NotNil(t, isN.Value)
	})
}

func TestLogicalOperators(t *testing.T) {
	t.Run("and with multiple expressions", func(t *testing.T) {
		and := And(
			Eq("a", 1),
			Eq("b", 2),
			Eq("c", 3),
		)
		assert.Len(t, and.Args, 3)
	})

	t.Run("or with multiple expressions", func(t *testing.T) {
		or := Or(
			Eq("status", "active"),
			Eq("status", "pending"),
		)
		assert.Len(t, or.Args, 2)
	})

	t.Run("not", func(t *testing.T) {
		not := Not(Eq("deleted", true))
		assert.NotNil(t, not.Arg)
	})
}

func TestSpatialGeometries(t *testing.T) {
	t.Run("point", func(t *testing.T) {
		geom := Point(-122.4194, 37.7749)
		// The value is now a *geojson.Geometry
		data, err := json.Marshal(geom.Value)
		require.NoError(t, err)
		var v map[string]any
		require.NoError(t, json.Unmarshal(data, &v))
		assert.Equal(t, "Point", v["type"])
		coords := v["coordinates"].([]any)
		assert.Equal(t, -122.4194, coords[0])
		assert.Equal(t, 37.7749, coords[1])
	})

	t.Run("point 3D", func(t *testing.T) {
		geom := Point3D(-122.4194, 37.7749, 100.0)
		v := geom.Value.(map[string]any)
		assert.Equal(t, "Point", v["type"])
		coords := v["coordinates"].([]float64)
		assert.Len(t, coords, 3)
	})

	t.Run("polygon", func(t *testing.T) {
		ring := [][]float64{
			{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0},
		}
		geom := Polygon(ring)
		data, err := json.Marshal(geom.Value)
		require.NoError(t, err)
		var v map[string]any
		require.NoError(t, json.Unmarshal(data, &v))
		assert.Equal(t, "Polygon", v["type"])
	})

	t.Run("geometry collection", func(t *testing.T) {
		geom := GeometryCollection(
			orb.Point{0, 0},
			orb.Point{1, 1},
		)
		data, err := json.Marshal(geom.Value)
		require.NoError(t, err)
		var v map[string]any
		require.NoError(t, json.Unmarshal(data, &v))
		assert.Equal(t, "GeometryCollection", v["type"])
		geoms := v["geometries"].([]any)
		assert.Len(t, geoms, 2)
	})
}

func TestBBox(t *testing.T) {
	t.Run("2D bbox", func(t *testing.T) {
		bbox := BBox(-122.5, 37.5, -122.0, 38.0)
		assert.Equal(t, []float64{-122.5, 37.5, -122.0, 38.0}, bbox.Extent)
	})

	t.Run("3D bbox", func(t *testing.T) {
		bbox := BBox3D(-122.5, 37.5, 0, -122.0, 38.0, 1000)
		assert.Equal(t, []float64{-122.5, 37.5, 0, -122.0, 38.0, 1000}, bbox.Extent)
	})
}

func TestSpatialOperators(t *testing.T) {
	bbox := BBox(-122.5, 37.5, -122.0, 38.0)

	tests := []struct {
		name     string
		expr     *filter.SpatialComparison
		expected string
	}{
		{"s_intersects", SIntersects(bbox), filter.GeometryIntersects},
		{"s_equals", SEquals(bbox), filter.GeometryEquals},
		{"s_disjoint", SDisjoint(bbox), filter.GeometryDisjoint},
		{"s_touches", STouches(bbox), filter.GeometryTouches},
		{"s_within", SWithin(bbox), filter.GeometryWithin},
		{"s_overlaps", SOverlaps(bbox), filter.GeometryOverlaps},
		{"s_crosses", SCrosses(bbox), filter.GeometryCrosses},
		{"s_contains", SContains(bbox), filter.GeometryContains},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expr.Name)
			// Verify the left side is always Property("geometry")
			prop, ok := tt.expr.Left.(*filter.Property)
			require.True(t, ok)
			assert.Equal(t, "geometry", prop.Name)
		})
	}
}

func TestSpatialOperatorsWithOrbTypes(t *testing.T) {
	t.Run("with orb.Point", func(t *testing.T) {
		pt := orb.Point{-122.4194, 37.7749}
		expr := SIntersects(pt)
		assert.Equal(t, filter.GeometryIntersects, expr.Name)
		assert.NotNil(t, expr.Right)
	})

	t.Run("with orb.Polygon", func(t *testing.T) {
		poly := orb.Polygon{
			{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}},
		}
		expr := SWithin(poly)
		assert.Equal(t, filter.GeometryWithin, expr.Name)
	})

	t.Run("with orb.Bound", func(t *testing.T) {
		bound := orb.Bound{Min: orb.Point{-122.5, 37.5}, Max: orb.Point{-122.0, 38.0}}
		expr := SIntersects(bound)
		assert.Equal(t, filter.GeometryIntersects, expr.Name)
		// Verify it was converted to BoundingBox
		bbox, ok := expr.Right.(*filter.BoundingBox)
		require.True(t, ok)
		assert.Equal(t, []float64{-122.5, 37.5, -122.0, 38.0}, bbox.Extent)
	})

	t.Run("filter serialization with orb geometry", func(t *testing.T) {
		poly := orb.Polygon{
			{{-122.5, 37.5}, {-122.0, 37.5}, {-122.0, 38.0}, {-122.5, 38.0}, {-122.5, 37.5}},
		}
		f := NewFilterBuilder().
			Where(SIntersects(poly)).
			Build()

		data, err := json.Marshal(f)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"op":"s_intersects"`)
		assert.Contains(t, string(data), `"Polygon"`)
	})
}

func TestTemporalExpressions(t *testing.T) {
	t.Run("timestamp from string", func(t *testing.T) {
		ts := Timestamp("2023-06-15T12:00:00Z")
		assert.Equal(t, 2023, ts.Value.Year())
		assert.Equal(t, time.June, ts.Value.Month())
		assert.Equal(t, 15, ts.Value.Day())
	})

	t.Run("timestamp from time.Time", func(t *testing.T) {
		tm := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		ts := TimestampFromTime(tm)
		assert.Equal(t, tm, ts.Value)
	})

	t.Run("date from string", func(t *testing.T) {
		d := Date("2023-06-15")
		assert.Equal(t, 2023, d.Value.Year())
		assert.Equal(t, time.June, d.Value.Month())
		assert.Equal(t, 15, d.Value.Day())
	})

	t.Run("interval", func(t *testing.T) {
		interval := Interval(Timestamp("2023-01-01T00:00:00Z"), Timestamp("2023-12-31T23:59:59Z"))
		assert.NotNil(t, interval.Start)
		assert.NotNil(t, interval.End)
	})

	t.Run("interval from strings", func(t *testing.T) {
		interval := IntervalFromStrings("2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z")
		assert.NotNil(t, interval.Start)
		assert.NotNil(t, interval.End)
	})

	t.Run("open interval before", func(t *testing.T) {
		interval := OpenIntervalBefore("2023-12-31T23:59:59Z")
		assert.Nil(t, interval.Start)
		assert.NotNil(t, interval.End)
	})

	t.Run("open interval after", func(t *testing.T) {
		interval := OpenIntervalAfter("2023-01-01T00:00:00Z")
		assert.NotNil(t, interval.Start)
		assert.Nil(t, interval.End)
	})
}

func TestTemporalOperators(t *testing.T) {
	datetime := Property("datetime")
	interval := IntervalFromStrings("2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z")

	tests := []struct {
		name     string
		expr     *filter.TemporalComparison
		expected string
	}{
		{"t_after", TAfter(datetime, Timestamp("2023-01-01T00:00:00Z")), filter.TimeAfter},
		{"t_before", TBefore(datetime, Timestamp("2023-12-31T23:59:59Z")), filter.TimeBefore},
		{"t_contains", TContains(interval, datetime), filter.TimeContains},
		{"t_disjoint", TDisjoint(datetime, interval), filter.TimeDisjoint},
		{"t_during", TDuring(datetime, interval), filter.TimeDuring},
		{"t_equals", TEquals(datetime, Timestamp("2023-06-15T00:00:00Z")), filter.TimeEquals},
		{"t_intersects", TIntersects(datetime, interval), filter.TimeIntersects},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expr.Name)
		})
	}
}

func TestArrayOperators(t *testing.T) {
	arr1 := Property("tags")
	arr2 := Array(String("a"), String("b"), String("c"))

	tests := []struct {
		name     string
		expr     *filter.ArrayComparison
		expected string
	}{
		{"a_equals", AEquals(arr1, arr2), filter.ArrayEquals},
		{"a_contains", AContains(arr1, arr2), filter.ArrayContains},
		{"a_containedby", AContainedBy(arr1, arr2), filter.ArrayContainedBy},
		{"a_overlaps", AOverlaps(arr1, arr2), filter.ArrayOverlaps},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expr.Name)
		})
	}

	t.Run("array literal", func(t *testing.T) {
		arr := Array(String("x"), Number(1), Boolean(true))
		assert.Len(t, arr, 3)
	})
}

func TestFilterBuilder(t *testing.T) {
	t.Run("basic filter", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Lt("eo:cloud_cover", 10.0)).
			Build()

		require.NotNil(t, f)
		assert.NotNil(t, f.Expression)
	})

	t.Run("chained and conditions", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Lt("eo:cloud_cover", 10.0)).
			And(Gt("quality", 80)).
			Build()

		require.NotNil(t, f)
		and, ok := f.Expression.(*filter.And)
		require.True(t, ok)
		assert.Len(t, and.Args, 2)
	})

	t.Run("with or branch", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Eq("status", "active")).
			Or(
				Eq("type", "image"),
				Eq("type", "video"),
			).
			Build()

		require.NotNil(t, f)
	})

	t.Run("empty filter", func(t *testing.T) {
		f := NewFilterBuilder().Build()
		assert.Nil(t, f)
	})
}

func TestFilterJSONSerialization(t *testing.T) {
	t.Run("simple comparison", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Lt("eo:cloud_cover", 10)).
			Build()

		data, err := json.Marshal(f)
		require.NoError(t, err)
		// Note: JSON marshaling escapes < as \u003c
		assert.Contains(t, string(data), `"op"`)
		assert.Contains(t, string(data), `"eo:cloud_cover"`)
	})

	t.Run("complex filter", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Lt("eo:cloud_cover", 10)).
			And(SIntersects(BBox(-122.5, 37.5, -122.0, 38.0))).
			And(TIntersects(Property("datetime"), IntervalFromStrings("2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z"))).
			Build()

		data, err := json.Marshal(f)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"op":"and"`)
	})

	t.Run("spatial filter with bbox", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(SIntersects(BBox(-122.5, 37.5, -122.0, 38.0))).
			Build()

		data, err := json.Marshal(f)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"op":"s_intersects"`)
		assert.Contains(t, string(data), `"bbox"`)
	})

	t.Run("in list filter", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(In("collection", "sentinel-1", "sentinel-2")).
			Build()

		data, err := json.Marshal(f)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"op":"in"`)
	})
}

func TestRealWorldScenarios(t *testing.T) {
	t.Run("satellite imagery search", func(t *testing.T) {
		// Find low-cloud Sentinel-2 imagery over San Francisco in 2023
		f := NewFilterBuilder().
			Where(Eq("collection", "sentinel-2")).
			And(Lt("eo:cloud_cover", 15)).
			And(TIntersects(Property("datetime"), IntervalFromStrings("2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z"))).
			And(SIntersects(BBox(-122.5, 37.7, -122.35, 37.85))).
			And(Eq("platform", "sentinel-2a")).
			Build()

		require.NotNil(t, f)
		and, ok := f.Expression.(*filter.And)
		require.True(t, ok)
		assert.Len(t, and.Args, 5)

		// Verify it serializes correctly
		data, err := json.Marshal(f)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("exclude deleted items", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Eq("collection", "my-collection")).
			And(Not(Eq("deleted", true))).
			Build()

		require.NotNil(t, f)
	})

	t.Run("search by id pattern", func(t *testing.T) {
		f := NewFilterBuilder().
			Where(Like("id", "S2A_MSIL2A_%")).
			And(TAfter(Property("datetime"), Timestamp("2023-06-01T00:00:00Z"))).
			Build()

		require.NotNil(t, f)
	})
}
