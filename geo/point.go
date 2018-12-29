package geo

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
)

// Point maps against MySQL geographical point.
type Point struct {
	Lng float64
	Lat float64
}

// String returns the WKT (Well Known Text) representation of the point.
func (p Point) String() string {
	return fmt.Sprintf("POINT(%v %v)", p.Lng, p.Lat)
}

// Scan implements the SQL driver.Scanner interface and will scan the
// MySQL POINT(x y) into the Point struct.
func (p *Point) Scan(val interface{}) error {
	b, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("geo: cannot scan type into bytes: %T", b)
	}

	// MySQL bug, it returns the internal representation with 4 zero bytes before
	// the value: https://bugs.mysql.com/bug.php?id=69798
	b = b[4:]

	r := bytes.NewReader(b)

	var wkbByteOrder uint8
	if err := binary.Read(r, binary.LittleEndian, &wkbByteOrder); err != nil {
		return err
	}

	var byteOrder binary.ByteOrder
	switch wkbByteOrder {
	case 0:
		byteOrder = binary.BigEndian
	case 1:
		byteOrder = binary.LittleEndian
	default:
		return fmt.Errorf("geo: invalid byte order %v", wkbByteOrder)
	}

	var wkbGeometryType uint32
	if err := binary.Read(r, byteOrder, &wkbGeometryType); err != nil {
		return err
	}

	if wkbGeometryType != 1 {
		return fmt.Errorf("geo: unexpected geometry type: wanted 1 (point), got %d", wkbGeometryType)
	}

	if err := binary.Read(r, byteOrder, p); err != nil {
		return err
	}

	return nil
}

// Value implements the SQL driver.Valuer interface and will return the string
// representation of the Point struct by calling the String() method
func (p Point) Value() (driver.Value, error) {
	w := bytes.NewBuffer(nil)

	// MySQL bug, it returns the internal representation with 4 zero bytes before
	// the value: https://bugs.mysql.com/bug.php?id=69798
	w.Write([]byte{0, 0, 0, 0})

	var wkbByteOrder uint8 = 1
	if err := binary.Write(w, binary.LittleEndian, wkbByteOrder); err != nil {
		return nil, err
	}

	var wkbGeometryType uint32 = 1
	if err := binary.Write(w, binary.LittleEndian, wkbGeometryType); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, p); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Valid returns whether a GeoPoint is within [-90, 90] latitude and [-180, 180] longitude.
func (p Point) Valid() bool {
	return -90 <= p.Lat && p.Lat <= 90 && -180 <= p.Lng && p.Lng <= 180
}
