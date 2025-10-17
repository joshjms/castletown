package allocator_test

import (
	"testing"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/sandbox/allocator"
	"github.com/stretchr/testify/require"
)

func TestAllocator(t *testing.T) {
	config.UseDefaults()

	a, err := allocator.NewAllocator(2)
	require.NoError(t, err, "Failed to create allocator: %v", err)

	i1, _ := a.Allocate()
	i2, _ := a.Allocate()
	i3, _ := a.Allocate()
	a.Free(i1)
	i4, _ := a.Allocate()

	require.Equal(t, 0, i1, "incorrect index for first allocation, expected 0")
	require.Equal(t, 1, i2, "incorrect index for second allocation, expected 1")
	require.Equal(t, -1, i3, "incorrect index for third allocation, expected -1")
	require.Equal(t, 0, i4, "incorrect index for fourth allocation, expected 0")
}
