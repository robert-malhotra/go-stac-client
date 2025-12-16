// Package stac provides types for working with SpatioTemporal Asset Catalog (STAC) data.
//
// This package implements STAC Item, Collection, Link, and Asset types with support
// for "foreign members" - additional JSON fields not defined in the STAC specification.
// Foreign members are preserved during JSON unmarshaling in the AdditionalFields map.
//
// Example usage:
//
//	var item stac.Item
//	json.Unmarshal(data, &item)
//
//	// Access standard fields
//	fmt.Println(item.ID)
//
//	// Access foreign members
//	if val, ok := item.AdditionalFields["custom_field"]; ok {
//	    fmt.Println(val)
//	}
package stac
