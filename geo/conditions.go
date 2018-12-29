package geo

import (
	"fmt"
)

// NearCondition stores the point and distance to filter around
type NearCondition struct {
	column string
	point  Point
	value  float64
}

// Near filters by distance around a geographic point
func Near(column string, point Point, value float64) *NearCondition {
	return &NearCondition{column, point, value}
}

// SQL returns the built SQL to apply the filter.
func (cond *NearCondition) SQL() string {
	const s string = `
		6371 * 2 * ASIN(
			SQRT(
				POWER(
					SIN(
						RADIANS(%f - ABS(Y(%s)))/2
					),
					2
				)
				+ COS(RADIANS(%f)) * COS(RADIANS(ABS(Y(%s))))
				* POWER(
					SIN(
						RADIANS(%f - X(%s))/2
					),
					2
				)
			)
		) < ?
	`
	return fmt.Sprintf(s, cond.point.Lat, cond.column, cond.point.Lat, cond.column, cond.point.Lng, cond.column)
}

// Values returns the SQL placeholders to apply the filter
func (cond *NearCondition) Values() []interface{} {
	return []interface{}{cond.value}
}
