package allocator_test

import (
	"testing"

	"github.com/joshjms/castletown/sandbox/allocator"
	"github.com/stretchr/testify/require"
)

func TestAllocator(t *testing.T) {
	a, err := allocator.NewAllocator()
	require.NoError(t, err, "Failed to create allocator: %v", err)

	i1, _ := a.Allocate()
	i2, _ := a.Allocate()
	i3, _ := a.Allocate()
	a.Free(i1)
	i4, _ := a.Allocate()

	require.Equal(t, i1, 0, "incorrect index for first allocation, expected 0")
	require.Equal(t, i2, 1, "incorrect index for second allocation, expected 1")
	require.Equal(t, i3, -1, "incorrect index for third allocation, expected -1")
	require.Equal(t, i4, 0, "incorrect index for fourth allocation, expected 0")
}
