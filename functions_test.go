package gocache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   string
	Name string
}

type Product struct {
	SKU   string
	Price float64
}

type OrderItem struct {
	Quantity int
}

func TestGetTypeName(t *testing.T) {
	t.Run("user type name", func(t *testing.T) {
		typeName := GetTypeName[User]()
		assert.Equal(t, "user", typeName)
	})

	t.Run("product type name", func(t *testing.T) {
		typeName := GetTypeName[Product]()
		assert.Equal(t, "product", typeName)
	})

	t.Run("orderitem type name", func(t *testing.T) {
		typeName := GetTypeName[OrderItem]()
		assert.Equal(t, "orderitem", typeName)
	})

	t.Run("string type name", func(t *testing.T) {
		typeName := GetTypeName[string]()
		assert.Equal(t, "string", typeName)
	})

	t.Run("int type name", func(t *testing.T) {
		typeName := GetTypeName[int]()
		assert.Equal(t, "int", typeName)
	})
}
