package geo

import (
	"fmt"

	"libs.altipla.consulting/database"
)

type pointSorter struct {
	column string
	point  Point
}

// DistanceToPoint orders by distance to a geographic point.
func DistanceToPoint(column string, point Point) database.Sorter {
	return &pointSorter{column, point}
}

// SQL returns the built SQL to apply the order.
func (order *pointSorter) SQL() string {
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
		)
	`
	return fmt.Sprintf(s, order.point.Lat, order.column, order.point.Lat, order.column, order.point.Lng, order.column)
}
