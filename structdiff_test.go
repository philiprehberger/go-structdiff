package structdiff

import (
	"encoding/json"
	"strings"
	"testing"
)

type Address struct {
	Street string
	City   string
	Zip    string
}

type Person struct {
	Name    string
	Age     int
	Address Address
}

type WithPointer struct {
	Name  string
	Value *int
}

type WithSlice struct {
	Items []string
}

type WithMap struct {
	Data map[string]int
}

type WithTags struct {
	Name     string
	Internal string `diff:"-"`
	Score    int    `diff:"-"`
}

type WithUnexported struct {
	Name     string
	internal string //nolint:unused
}

type Item struct {
	Name  string
	Price float64
}

type Order struct {
	ID    int
	Items []Item
}

type Empty struct{}

func TestSimpleEqual(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY", Zip: "10001"}}
	b := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY", Zip: "10001"}}

	changes := Compare(a, b)
	if changes != nil {
		t.Errorf("expected nil changes for equal structs, got %v", changes)
	}
}

func TestSimpleDifferent(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Bob", Age: 25}

	changes := Compare(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}

	found := map[string]bool{}
	for _, c := range changes {
		found[c.Path] = true
		switch c.Path {
		case "Name":
			if c.Old != "Alice" || c.New != "Bob" {
				t.Errorf("unexpected Name change: %v", c)
			}
		case "Age":
			if c.Old != 30 || c.New != 25 {
				t.Errorf("unexpected Age change: %v", c)
			}
		default:
			t.Errorf("unexpected path: %s", c.Path)
		}
	}
	if !found["Name"] || !found["Age"] {
		t.Errorf("missing expected changes")
	}
}

func TestNestedStruct(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY", Zip: "10001"}}
	b := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "LA", Zip: "90001"}}

	changes := Compare(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}

	paths := map[string]bool{}
	for _, c := range changes {
		paths[c.Path] = true
	}
	if !paths["Address.City"] {
		t.Error("expected Address.City change")
	}
	if !paths["Address.Zip"] {
		t.Error("expected Address.Zip change")
	}
}

func TestSliceDifferentLength(t *testing.T) {
	a := WithSlice{Items: []string{"a", "b", "c"}}
	b := WithSlice{Items: []string{"a", "b"}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Items[2]" {
		t.Errorf("expected path Items[2], got %s", changes[0].Path)
	}
	if changes[0].Old != "c" {
		t.Errorf("expected old value 'c', got %v", changes[0].Old)
	}
	if changes[0].New != nil {
		t.Errorf("expected new value nil, got %v", changes[0].New)
	}
}

func TestSliceDifferentElements(t *testing.T) {
	a := WithSlice{Items: []string{"a", "b", "c"}}
	b := WithSlice{Items: []string{"a", "x", "c"}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Items[1]" {
		t.Errorf("expected path Items[1], got %s", changes[0].Path)
	}
}

func TestMapDifferences(t *testing.T) {
	a := WithMap{Data: map[string]int{"x": 1, "y": 2}}
	b := WithMap{Data: map[string]int{"x": 1, "y": 3, "z": 4}}

	changes := Compare(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}

	paths := map[string]bool{}
	for _, c := range changes {
		paths[c.Path] = true
	}
	if !paths["Data[y]"] {
		t.Error("expected Data[y] change")
	}
	if !paths["Data[z]"] {
		t.Error("expected Data[z] change")
	}
}

func TestMapRemovedKey(t *testing.T) {
	a := WithMap{Data: map[string]int{"x": 1, "y": 2}}
	b := WithMap{Data: map[string]int{"x": 1}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Data[y]" {
		t.Errorf("expected path Data[y], got %s", changes[0].Path)
	}
	if changes[0].New != nil {
		t.Errorf("expected new value nil, got %v", changes[0].New)
	}
}

func TestPointerNilVsNonNil(t *testing.T) {
	val := 42
	a := WithPointer{Name: "test", Value: nil}
	b := WithPointer{Name: "test", Value: &val}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Value" {
		t.Errorf("expected path Value, got %s", changes[0].Path)
	}
	if changes[0].Old != nil {
		t.Errorf("expected old nil, got %v", changes[0].Old)
	}
	if changes[0].New != 42 {
		t.Errorf("expected new 42, got %v", changes[0].New)
	}
}

func TestPointerDifferentValues(t *testing.T) {
	v1, v2 := 10, 20
	a := WithPointer{Name: "test", Value: &v1}
	b := WithPointer{Name: "test", Value: &v2}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Value" {
		t.Errorf("expected path Value, got %s", changes[0].Path)
	}
}

func TestPointerBothNil(t *testing.T) {
	a := WithPointer{Name: "test", Value: nil}
	b := WithPointer{Name: "test", Value: nil}

	changes := Compare(a, b)
	if changes != nil {
		t.Errorf("expected nil changes, got %v", changes)
	}
}

func TestIgnoreOption(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Bob", Age: 25}

	changes := Compare(a, b, Ignore("Age"))
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Name" {
		t.Errorf("expected path Name, got %s", changes[0].Path)
	}
}

func TestIgnoreMultipleFields(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Bob", Age: 25}

	changes := Compare(a, b, Ignore("Name", "Age"))
	if changes != nil {
		t.Errorf("expected nil changes, got %v", changes)
	}
}

func TestIgnoreTagOption(t *testing.T) {
	a := WithTags{Name: "Alice", Internal: "secret1", Score: 100}
	b := WithTags{Name: "Bob", Internal: "secret2", Score: 200}

	changes := Compare(a, b, IgnoreTag("diff"))
	if len(changes) != 1 {
		t.Fatalf("expected 1 change (only Name), got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Name" {
		t.Errorf("expected path Name, got %s", changes[0].Path)
	}
}

func TestEqualTrue(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Alice", Age: 30}

	if !Equal(a, b) {
		t.Error("expected Equal to return true")
	}
}

func TestEqualFalse(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Bob", Age: 30}

	if Equal(a, b) {
		t.Error("expected Equal to return false")
	}
}

func TestEqualWithOptions(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Alice", Age: 25}

	if !Equal(a, b, Ignore("Age")) {
		t.Error("expected Equal with Ignore(Age) to return true")
	}
}

func TestChangeString(t *testing.T) {
	c := Change{Path: "Name", Old: "Alice", New: "Bob"}
	expected := "Name: Alice → Bob"
	if c.String() != expected {
		t.Errorf("expected %q, got %q", expected, c.String())
	}
}

func TestChangeStringNilValues(t *testing.T) {
	c := Change{Path: "Value", Old: nil, New: 42}
	expected := "Value: <nil> → 42"
	if c.String() != expected {
		t.Errorf("expected %q, got %q", expected, c.String())
	}
}

func TestTypeMismatch(t *testing.T) {
	type A struct{ X int }
	type B struct{ X string }

	changes := Compare(A{X: 1}, B{X: "hello"})
	if len(changes) != 1 {
		t.Fatalf("expected 1 change for type mismatch, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "" {
		t.Errorf("expected empty path for top-level mismatch, got %q", changes[0].Path)
	}
}

func TestEmptyStructs(t *testing.T) {
	a := Empty{}
	b := Empty{}

	changes := Compare(a, b)
	if changes != nil {
		t.Errorf("expected nil changes for empty structs, got %v", changes)
	}
}

func TestUnexportedFieldsSkipped(t *testing.T) {
	a := WithUnexported{Name: "Alice"}
	b := WithUnexported{Name: "Alice"}

	changes := Compare(a, b)
	if changes != nil {
		t.Errorf("expected nil changes, got %v", changes)
	}
}

func TestNestedSliceOfStructs(t *testing.T) {
	a := Order{
		ID: 1,
		Items: []Item{
			{Name: "Widget", Price: 9.99},
			{Name: "Gadget", Price: 19.99},
		},
	}
	b := Order{
		ID: 1,
		Items: []Item{
			{Name: "Widget", Price: 9.99},
			{Name: "Gadget", Price: 24.99},
		},
	}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Items[1].Price" {
		t.Errorf("expected path Items[1].Price, got %s", changes[0].Path)
	}
}

func TestSliceGrowth(t *testing.T) {
	a := WithSlice{Items: []string{"a"}}
	b := WithSlice{Items: []string{"a", "b", "c"}}

	changes := Compare(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}

	for _, c := range changes {
		if c.Old != nil {
			t.Errorf("expected old nil for new elements, got %v at %s", c.Old, c.Path)
		}
	}
}

func TestBooleanFields(t *testing.T) {
	type Flags struct {
		Active  bool
		Visible bool
	}

	a := Flags{Active: true, Visible: false}
	b := Flags{Active: false, Visible: true}

	changes := Compare(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}
}

func TestFloatFields(t *testing.T) {
	type Measurement struct {
		Value float64
	}

	a := Measurement{Value: 3.14}
	b := Measurement{Value: 2.71}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Value" {
		t.Errorf("expected path Value, got %s", changes[0].Path)
	}
}

// --- Patch tests ---

func TestPatchSimple(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	changes := []Change{
		{Path: "Name", Old: "Alice", New: "Bob"},
		{Path: "Age", Old: 30, New: 31},
	}

	err := Patch(&p, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Bob" {
		t.Errorf("expected Name=Bob, got %s", p.Name)
	}
	if p.Age != 31 {
		t.Errorf("expected Age=31, got %d", p.Age)
	}
}

func TestPatchNested(t *testing.T) {
	p := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY", Zip: "10001"}}
	changes := []Change{
		{Path: "Address.City", Old: "NY", New: "LA"},
	}

	err := Patch(&p, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Address.City != "LA" {
		t.Errorf("expected Address.City=LA, got %s", p.Address.City)
	}
}

func TestPatchFieldNotFound(t *testing.T) {
	p := Person{Name: "Alice"}
	changes := []Change{
		{Path: "NonExistent", Old: nil, New: "value"},
	}

	err := Patch(&p, changes)
	if err == nil {
		t.Fatal("expected error for non-existent field")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestPatchNonPointer(t *testing.T) {
	p := Person{Name: "Alice"}
	err := Patch(p, nil)
	if err == nil {
		t.Fatal("expected error for non-pointer target")
	}
}

func TestPatchNilPointer(t *testing.T) {
	var p *Person
	err := Patch(p, nil)
	if err == nil {
		t.Fatal("expected error for nil pointer target")
	}
}

func TestPatchEmptyChanges(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	err := Patch(&p, []Change{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Alice" || p.Age != 30 {
		t.Error("struct should not change with empty changes")
	}
}

func TestPatchNilChanges(t *testing.T) {
	p := Person{Name: "Alice"}
	err := Patch(&p, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Format tests ---

func TestFormatSimple(t *testing.T) {
	changes := []Change{
		{Path: "Name", Old: "Alice", New: "Bob"},
		{Path: "Age", Old: 30, New: 31},
	}

	result := Format(changes)
	if !strings.Contains(result, "Name:") {
		t.Errorf("expected Name in output, got: %s", result)
	}
	if !strings.Contains(result, "Age:") {
		t.Errorf("expected Age in output, got: %s", result)
	}
	// String values should be quoted.
	if !strings.Contains(result, `"Alice"`) {
		t.Errorf("expected quoted Alice in output, got: %s", result)
	}
	if !strings.Contains(result, `"Bob"`) {
		t.Errorf("expected quoted Bob in output, got: %s", result)
	}
}

func TestFormatEmpty(t *testing.T) {
	result := Format(nil)
	if result != "" {
		t.Errorf("expected empty string for nil changes, got: %q", result)
	}

	result = Format([]Change{})
	if result != "" {
		t.Errorf("expected empty string for empty changes, got: %q", result)
	}
}

func TestFormatNilValues(t *testing.T) {
	changes := []Change{
		{Path: "Value", Old: nil, New: 42},
	}

	result := Format(changes)
	if !strings.Contains(result, "<nil>") {
		t.Errorf("expected <nil> in output, got: %s", result)
	}
}

func TestFormatMultiLine(t *testing.T) {
	changes := []Change{
		{Path: "A", Old: 1, New: 2},
		{Path: "B", Old: 3, New: 4},
	}

	result := Format(changes)
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), result)
	}
}

// --- FormatJSON tests ---

func TestFormatJSONSimple(t *testing.T) {
	changes := []Change{
		{Path: "Name", Old: "Alice", New: "Bob"},
		{Path: "Age", Old: 30, New: 31},
	}

	data, err := FormatJSON(changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0]["path"] != "Name" {
		t.Errorf("expected path=Name, got %v", result[0]["path"])
	}
	if result[0]["old"] != "Alice" {
		t.Errorf("expected old=Alice, got %v", result[0]["old"])
	}
	if result[0]["new"] != "Bob" {
		t.Errorf("expected new=Bob, got %v", result[0]["new"])
	}
}

func TestFormatJSONEmpty(t *testing.T) {
	data, err := FormatJSON(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d entries", len(result))
	}
}

func TestFormatJSONNilValues(t *testing.T) {
	changes := []Change{
		{Path: "Value", Old: nil, New: 42},
	}

	data, err := FormatJSON(changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result[0]["old"] != nil {
		t.Errorf("expected old=nil, got %v", result[0]["old"])
	}
}

// --- OnlyFields tests ---

func TestOnlyFields(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{City: "NY"}}
	b := Person{Name: "Bob", Age: 25, Address: Address{City: "LA"}}

	changes := Compare(a, b, OnlyFields("Name"))
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Name" {
		t.Errorf("expected path Name, got %s", changes[0].Path)
	}
}

func TestOnlyFieldsMultiple(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{City: "NY"}}
	b := Person{Name: "Bob", Age: 25, Address: Address{City: "LA"}}

	changes := Compare(a, b, OnlyFields("Name", "Age"))
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(changes), changes)
	}

	paths := map[string]bool{}
	for _, c := range changes {
		paths[c.Path] = true
	}
	if !paths["Name"] || !paths["Age"] {
		t.Error("expected Name and Age changes")
	}
}

func TestOnlyFieldsNoMatch(t *testing.T) {
	a := Person{Name: "Alice", Age: 30}
	b := Person{Name: "Bob", Age: 25}

	changes := Compare(a, b, OnlyFields("Address"))
	if changes != nil {
		t.Errorf("expected nil changes when only comparing unchanged fields, got %v", changes)
	}
}

func TestOnlyFieldsWithNested(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{City: "NY"}}
	b := Person{Name: "Bob", Age: 25, Address: Address{City: "LA"}}

	changes := Compare(a, b, OnlyFields("Address"))
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Address.City" {
		t.Errorf("expected path Address.City, got %s", changes[0].Path)
	}
}

// --- Map comparison with map[string]any ---

func TestMapStringAnyComparison(t *testing.T) {
	type Container struct {
		Meta map[string]any
	}

	a := Container{Meta: map[string]any{"name": "Alice", "score": 100}}
	b := Container{Meta: map[string]any{"name": "Bob", "score": 100}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Meta[name]" {
		t.Errorf("expected path Meta[name], got %s", changes[0].Path)
	}
}

func TestMapStringAnyNewKey(t *testing.T) {
	type Container struct {
		Meta map[string]any
	}

	a := Container{Meta: map[string]any{"name": "Alice"}}
	b := Container{Meta: map[string]any{"name": "Alice", "age": 30}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Meta[age]" {
		t.Errorf("expected path Meta[age], got %s", changes[0].Path)
	}
}

func TestMapStringAnyRemovedKey(t *testing.T) {
	type Container struct {
		Meta map[string]any
	}

	a := Container{Meta: map[string]any{"name": "Alice", "age": 30}}
	b := Container{Meta: map[string]any{"name": "Alice"}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Meta[age]" {
		t.Errorf("expected path Meta[age], got %s", changes[0].Path)
	}
	if changes[0].New != nil {
		t.Errorf("expected new=nil for removed key, got %v", changes[0].New)
	}
}

func TestMapStringAnyNestedMap(t *testing.T) {
	type Container struct {
		Meta map[string]any
	}

	a := Container{Meta: map[string]any{
		"nested": map[string]any{"key": "val1"},
	}}
	b := Container{Meta: map[string]any{
		"nested": map[string]any{"key": "val2"},
	}}

	changes := Compare(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(changes), changes)
	}
	if changes[0].Path != "Meta[nested][key]" {
		t.Errorf("expected path Meta[nested][key], got %s", changes[0].Path)
	}
}

// --- Patch roundtrip test ---

func TestPatchRoundtrip(t *testing.T) {
	a := Person{Name: "Alice", Age: 30, Address: Address{Street: "Main St", City: "NY", Zip: "10001"}}
	b := Person{Name: "Bob", Age: 31, Address: Address{Street: "Main St", City: "LA", Zip: "90001"}}

	changes := Compare(a, b)
	err := Patch(&a, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !Equal(a, b) {
		t.Errorf("after patching, a should equal b. a=%+v b=%+v", a, b)
	}
}
