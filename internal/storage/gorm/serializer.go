package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
	"time"
)

// UnixMilliSecondSerializer milli serializer
type UnixMilliSecondSerializer struct{}

// Scan implements serializer interface
func (UnixMilliSecondSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	t := sql.NullTime{}
	if err = t.Scan(dbValue); err == nil && t.Valid {
		err = field.Set(ctx, dst, t.Time.UnixMilli())
	}

	return
}

// Value implements serializer interface
func (UnixMilliSecondSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (result interface{}, err error) {
	rv := reflect.ValueOf(fieldValue)
	switch v := fieldValue.(type) {
	case int64, int, uint, uint64, int32, uint32, int16, uint16:
		result = time.UnixMilli(reflect.Indirect(rv).Int()).UTC()
	case *int64, *int, *uint, *uint64, *int32, *uint32, *int16, *uint16:
		if rv.IsZero() {
			return nil, nil
		}
		result = time.UnixMilli(reflect.Indirect(rv).Int()).UTC()
	default:
		err = fmt.Errorf("invalid field type %#v for UnixMilliSecondSerializer, only int, uint supported", v)
	}
	return
}

func InitSerializer() {
	schema.RegisterSerializer("unixmstime", UnixMilliSecondSerializer{})
}
