package execute

import "testing"

func TestRing(t *testing.T) {
	// Ensure that the ring works properly when we append exactly
	// the correct number of elements to it.
	r := newRing(4)
	for i := 0; i < 4; i++ {
		r.Append(i)
	}

	// Now we will check the length, pop the next value, and append another
	// value in succession and check the ring is working properly.
	for i := 0; i < 6; i++ {
		if want, got := 4, r.Len(); want != got {
			t.Fatalf("unexpected ring length -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		if want, got := i, r.Next().(int); want != got {
			t.Fatalf("unexpected value -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		r.Append(i + 4)
	}

	// Append a value to ensure expansion works.
	// Do not fill the resized queue as we also want to verify that a non-filled
	// ring will continue to work properly.
	r.Append(10)
	for i := 0; i < 6; i++ {
		if want, got := 5, r.Len(); want != got {
			t.Fatalf("unexpected ring length -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		if want, got := i+6, r.Next().(int); want != got {
			t.Fatalf("unexpected value -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		r.Append(i + 11)
	}

	// If we take the remainder, we will continue to get values and the
	// length goes down to zero.
	for i := 0; i < 5; i++ {
		if want, got := 5-i, r.Len(); want != got {
			t.Fatalf("unexpected ring length -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
		if want, got := i+12, r.Next().(int); want != got {
			t.Fatalf("unexpected value -want/+got:\n\t- %d\n\t+ %d", want, got)
		}
	}
}
