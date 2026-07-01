package ptr

import "testing"

func TestOfReturnsPointerToValueCopy(t *testing.T) {
	value := "fox"
	pointer := Of(value)
	value = "admin"

	if pointer == nil {
		t.Fatal("Of() returned nil")
	}
	if *pointer != "fox" {
		t.Fatalf("*Of() = %q, want fox", *pointer)
	}
}

func TestCloneReturnsNilForNilPointer(t *testing.T) {
	if got := Clone[int](nil); got != nil {
		t.Fatalf("Clone(nil) = %#v, want nil", got)
	}
}

func TestCloneReturnsPointerToValueCopy(t *testing.T) {
	source := Of("fox")
	clone := Clone(source)
	*source = "admin"

	if clone == nil {
		t.Fatal("Clone() returned nil")
	}
	if clone == source {
		t.Fatal("Clone() returned original pointer")
	}
	if *clone != "fox" {
		t.Fatalf("*Clone() = %q, want fox", *clone)
	}
}

func TestValueReturnsZeroForNilPointer(t *testing.T) {
	if got := Value[int](nil); got != 0 {
		t.Fatalf("Value[int](nil) = %d, want 0", got)
	}
	if got := Value[string](nil); got != "" {
		t.Fatalf("Value[string](nil) = %q, want empty string", got)
	}
	if got := Value[bool](nil); got {
		t.Fatal("Value[bool](nil) = true, want false")
	}
}

func TestValueReturnsPointedValue(t *testing.T) {
	pointer := Of(42)

	if got := Value(pointer); got != 42 {
		t.Fatalf("Value() = %d, want 42", got)
	}
}

func TestValueOrReturnsFallbackForNilPointer(t *testing.T) {
	if got := ValueOr[int](nil, 7); got != 7 {
		t.Fatalf("ValueOr(nil, 7) = %d, want 7", got)
	}

	pointer := Of(3)
	if got := ValueOr(pointer, 7); got != 3 {
		t.Fatalf("ValueOr(pointer, 7) = %d, want 3", got)
	}
}

func TestFromReportsPointerPresence(t *testing.T) {
	if got, ok := From[int](nil); got != 0 || ok {
		t.Fatalf("From(nil) = %d, %v; want 0, false", got, ok)
	}

	pointer := Of(9)
	if got, ok := From(pointer); got != 9 || !ok {
		t.Fatalf("From(pointer) = %d, %v; want 9, true", got, ok)
	}
}

func TestIsNilReportsNilPointer(t *testing.T) {
	if !IsNil[int](nil) {
		t.Fatal("IsNil(nil) = false, want true")
	}

	pointer := Of(1)
	if IsNil(pointer) {
		t.Fatal("IsNil(pointer) = true, want false")
	}
}

func TestEqualComparesOptionalValues(t *testing.T) {
	if !Equal[int](nil, nil) {
		t.Fatal("Equal(nil, nil) = false, want true")
	}
	if Equal(nil, Of(1)) {
		t.Fatal("Equal(nil, Of(1)) = true, want false")
	}
	if !Equal(Of(1), Of(1)) {
		t.Fatal("Equal(Of(1), Of(1)) = false, want true")
	}
	if Equal(Of(1), Of(2)) {
		t.Fatal("Equal(Of(1), Of(2)) = true, want false")
	}
}

func TestSliceConvertsValuesToPointers(t *testing.T) {
	values := []int{1, 2, 3}
	pointers := Slice(values)
	values[0] = 99

	if len(pointers) != 3 {
		t.Fatalf("len(Slice()) = %d, want 3", len(pointers))
	}
	for i, want := range []int{1, 2, 3} {
		if pointers[i] == nil {
			t.Fatalf("pointers[%d] is nil", i)
		}
		if *pointers[i] != want {
			t.Fatalf("*pointers[%d] = %d, want %d", i, *pointers[i], want)
		}
	}
}

func TestSlicePreservesNilSlice(t *testing.T) {
	if got := Slice[int](nil); got != nil {
		t.Fatalf("Slice(nil) = %#v, want nil", got)
	}
}

func TestValuesConvertsPointersToValues(t *testing.T) {
	values := Values([]*int{Of(1), nil, Of(3)})

	if len(values) != 3 {
		t.Fatalf("len(Values()) = %d, want 3", len(values))
	}
	for i, want := range []int{1, 0, 3} {
		if values[i] != want {
			t.Fatalf("values[%d] = %d, want %d", i, values[i], want)
		}
	}
}

func TestValuesPreservesNilSlice(t *testing.T) {
	if got := Values[int](nil); got != nil {
		t.Fatalf("Values(nil) = %#v, want nil", got)
	}
}
