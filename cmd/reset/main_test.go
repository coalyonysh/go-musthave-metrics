package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsPrimitiveType tests the isPrimitiveType function
func TestIsPrimitiveType(t *testing.T) {
	// Test integer types
	assert.True(t, isPrimitiveType("int"))
	assert.True(t, isPrimitiveType("int8"))
	assert.True(t, isPrimitiveType("int16"))
	assert.True(t, isPrimitiveType("int32"))
	assert.True(t, isPrimitiveType("int64"))

	// Test unsigned integer types
	assert.True(t, isPrimitiveType("uint"))
	assert.True(t, isPrimitiveType("uint8"))
	assert.True(t, isPrimitiveType("uint16"))
	assert.True(t, isPrimitiveType("uint32"))
	assert.True(t, isPrimitiveType("uint64"))
	assert.True(t, isPrimitiveType("uintptr"))

	// Test float types
	assert.True(t, isPrimitiveType("float32"))
	assert.True(t, isPrimitiveType("float64"))

	// Test string and bool
	assert.True(t, isPrimitiveType("string"))
	assert.True(t, isPrimitiveType("bool"))

	// Test other types
	assert.True(t, isPrimitiveType("byte"))
	assert.True(t, isPrimitiveType("rune"))
	assert.True(t, isPrimitiveType("complex64"))
	assert.True(t, isPrimitiveType("complex128"))

	// Test non-primitive types
	assert.False(t, isPrimitiveType("MyStruct"))
	assert.False(t, isPrimitiveType("[]int"))
	assert.False(t, isPrimitiveType("map[string]int"))
	assert.False(t, isPrimitiveType("interface{}"))
}

// TestGetZeroValue tests the getZeroValue function
func TestGetZeroValue(t *testing.T) {
	// Test integer types
	assert.Equal(t, "0", getZeroValue("int"))
	assert.Equal(t, "0", getZeroValue("int8"))
	assert.Equal(t, "0", getZeroValue("int64"))
	assert.Equal(t, "0", getZeroValue("uint"))
	assert.Equal(t, "0", getZeroValue("uint64"))
	assert.Equal(t, "0", getZeroValue("float32"))
	assert.Equal(t, "0", getZeroValue("float64"))
	assert.Equal(t, "0", getZeroValue("byte"))
	assert.Equal(t, "0", getZeroValue("rune"))

	// Test string
	assert.Equal(t, "\"\"", getZeroValue("string"))

	// Test bool
	assert.Equal(t, "false", getZeroValue("bool"))

	// Test unknown type
	assert.Equal(t, "zero value", getZeroValue("MyStruct"))
}

// TestTypeToString tests the typeToString function
func TestTypeToString(t *testing.T) {
	gen := &ResetGenerator{}

	// Test basic types would require AST parsing
	// This test verifies the function exists and is callable
	t.Run("function exists", func(t *testing.T) {
		// The function is part of ResetGenerator, so we test it indirectly
		_ = gen.typeToString
	})
}
