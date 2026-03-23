// Package structdiff provides field-level struct comparison for Go.
package structdiff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Change represents a single field difference between two structs.
type Change struct {
	// Path is the dot-notation path to the changed field (e.g., "Address.City").
	Path string
	// Old is the value in the first struct.
	Old any
	// New is the value in the second struct.
	New any
}

// String returns a human-readable representation of the change.
func (c Change) String() string {
	return fmt.Sprintf("%s: %v → %v", c.Path, c.Old, c.New)
}

// Option configures comparison behavior.
type Option func(*config)

type config struct {
	ignoreFields map[string]bool
	onlyFields   map[string]bool
	ignoreTag    string
}

// Ignore returns an Option that skips the given field names during comparison.
func Ignore(fields ...string) Option {
	return func(c *config) {
		for _, f := range fields {
			c.ignoreFields[f] = true
		}
	}
}

// IgnoreTag returns an Option that skips fields whose struct tag matches the
// given value. For example, IgnoreTag("diff:\"-\"") skips fields tagged with
// `diff:"-"`.
func IgnoreTag(tag string) Option {
	return func(c *config) {
		c.ignoreTag = tag
	}
}

// OnlyFields returns an Option that restricts comparison to only the specified
// field names. This is the opposite of Ignore.
func OnlyFields(fields ...string) Option {
	return func(c *config) {
		if c.onlyFields == nil {
			c.onlyFields = make(map[string]bool)
		}
		for _, f := range fields {
			c.onlyFields[f] = true
		}
	}
}

// Compare performs a deep, field-level comparison of two values and returns a
// slice of changes. It handles all primitive types, strings, slices, maps,
// nested structs, and pointers. Paths use dot notation (e.g., "Address.City")
// and bracket notation for indices (e.g., "Items[2].Name"). Returns nil if the
// values are equal.
func Compare(a, b any, opts ...Option) []Change {
	cfg := &config{
		ignoreFields: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Handle type mismatches at the top level.
	if va.Type() != vb.Type() {
		return []Change{{Path: "", Old: a, New: b}}
	}

	var changes []Change
	compare(va, vb, "", cfg, &changes)
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// Equal returns true if a and b are deeply equal according to the given options.
func Equal(a, b any, opts ...Option) bool {
	return len(Compare(a, b, opts...)) == 0
}

func compare(a, b reflect.Value, path string, cfg *config, changes *[]Change) {
	// Handle invalid (zero) values.
	if !a.IsValid() && !b.IsValid() {
		return
	}
	if !a.IsValid() || !b.IsValid() {
		*changes = append(*changes, Change{Path: path, Old: valueToAny(a), New: valueToAny(b)})
		return
	}

	// Unwrap interfaces.
	if a.Kind() == reflect.Interface {
		a = a.Elem()
	}
	if b.Kind() == reflect.Interface {
		b = b.Elem()
	}

	// Re-check validity after unwrapping.
	if !a.IsValid() && !b.IsValid() {
		return
	}
	if !a.IsValid() || !b.IsValid() {
		*changes = append(*changes, Change{Path: path, Old: valueToAny(a), New: valueToAny(b)})
		return
	}

	// Dereference pointers.
	if a.Kind() == reflect.Ptr || b.Kind() == reflect.Ptr {
		comparePointers(a, b, path, cfg, changes)
		return
	}

	// Type mismatch after dereferencing.
	if a.Type() != b.Type() {
		*changes = append(*changes, Change{Path: path, Old: valueToAny(a), New: valueToAny(b)})
		return
	}

	switch a.Kind() {
	case reflect.Struct:
		compareStructs(a, b, path, cfg, changes)
	case reflect.Slice, reflect.Array:
		compareSlices(a, b, path, cfg, changes)
	case reflect.Map:
		compareMaps(a, b, path, cfg, changes)
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			*changes = append(*changes, Change{Path: path, Old: a.Interface(), New: b.Interface()})
		}
	}
}

func comparePointers(a, b reflect.Value, path string, cfg *config, changes *[]Change) {
	aIsNil := !a.IsValid() || (a.Kind() == reflect.Ptr && a.IsNil())
	bIsNil := !b.IsValid() || (b.Kind() == reflect.Ptr && b.IsNil())

	if aIsNil && bIsNil {
		return
	}
	if aIsNil {
		val := b
		if b.Kind() == reflect.Ptr {
			val = b.Elem()
		}
		*changes = append(*changes, Change{Path: path, Old: nil, New: val.Interface()})
		return
	}
	if bIsNil {
		val := a
		if a.Kind() == reflect.Ptr {
			val = a.Elem()
		}
		*changes = append(*changes, Change{Path: path, Old: val.Interface(), New: nil})
		return
	}

	ae := a
	if a.Kind() == reflect.Ptr {
		ae = a.Elem()
	}
	be := b
	if b.Kind() == reflect.Ptr {
		be = b.Elem()
	}
	compare(ae, be, path, cfg, changes)
}

func compareStructs(a, b reflect.Value, path string, cfg *config, changes *[]Change) {
	t := a.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		// Skip ignored fields.
		if cfg.ignoreFields[field.Name] {
			continue
		}

		// If onlyFields is set, skip fields not in the set (root level only).
		if len(cfg.onlyFields) > 0 && path == "" && !cfg.onlyFields[field.Name] {
			continue
		}

		// Skip fields with the ignore tag.
		if cfg.ignoreTag != "" {
			tagVal := field.Tag.Get("diff")
			if tagVal == "-" {
				continue
			}
		}

		fieldPath := field.Name
		if path != "" {
			fieldPath = path + "." + field.Name
		}

		compare(a.Field(i), b.Field(i), fieldPath, cfg, changes)
	}
}

func compareSlices(a, b reflect.Value, path string, cfg *config, changes *[]Change) {
	aLen := a.Len()
	bLen := b.Len()
	minLen := aLen
	if bLen < minLen {
		minLen = bLen
	}

	for i := 0; i < minLen; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		compare(a.Index(i), b.Index(i), elemPath, cfg, changes)
	}

	// Handle extra elements.
	for i := minLen; i < aLen; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		*changes = append(*changes, Change{Path: elemPath, Old: a.Index(i).Interface(), New: nil})
	}
	for i := minLen; i < bLen; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		*changes = append(*changes, Change{Path: elemPath, Old: nil, New: b.Index(i).Interface()})
	}
}

func compareMaps(a, b reflect.Value, path string, cfg *config, changes *[]Change) {
	// Collect all keys.
	keySet := make(map[string]reflect.Value)
	for _, k := range a.MapKeys() {
		keySet[fmt.Sprintf("%v", k.Interface())] = k
	}
	for _, k := range b.MapKeys() {
		keySet[fmt.Sprintf("%v", k.Interface())] = k
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, keyStr := range keys {
		k := keySet[keyStr]
		elemPath := fmt.Sprintf("%s[%s]", path, keyStr)

		aVal := a.MapIndex(k)
		bVal := b.MapIndex(k)

		if !aVal.IsValid() {
			*changes = append(*changes, Change{Path: elemPath, Old: nil, New: bVal.Interface()})
			continue
		}
		if !bVal.IsValid() {
			*changes = append(*changes, Change{Path: elemPath, Old: aVal.Interface(), New: nil})
			continue
		}

		compare(aVal, bVal, elemPath, cfg, changes)
	}
}

func valueToAny(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

// Patch applies a slice of changes to a struct pointer, setting each field
// identified by its dot-notation path to the change's New value. The target
// must be a non-nil pointer to a struct. Returns an error if a field path
// is not found or cannot be set.
func Patch(target any, changes []Change) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("structdiff: target must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("structdiff: target must be a pointer to a struct, got pointer to %s", v.Kind())
	}

	for _, c := range changes {
		if err := setFieldByPath(v, c.Path, c.New); err != nil {
			return err
		}
	}
	return nil
}

func setFieldByPath(v reflect.Value, path string, value any) error {
	parts := strings.Split(path, ".")
	current := v

	for i, part := range parts {
		// Dereference pointers along the way.
		for current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return fmt.Errorf("structdiff: nil pointer at path %q", strings.Join(parts[:i], "."))
			}
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return fmt.Errorf("structdiff: field %q is not a struct", strings.Join(parts[:i], "."))
		}

		field := current.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("structdiff: field %q not found", path)
		}
		if !field.CanSet() {
			return fmt.Errorf("structdiff: field %q is not settable", path)
		}

		if i == len(parts)-1 {
			// Last segment — set the value.
			if value == nil {
				field.Set(reflect.Zero(field.Type()))
			} else {
				field.Set(reflect.ValueOf(value))
			}
			return nil
		}

		current = field
	}

	return fmt.Errorf("structdiff: empty path")
}

// Format returns a human-readable multi-line string describing the changes.
// Each line has the form: Path: old → new
// For string values, the output wraps them in quotes.
func Format(changes []Change) string {
	if len(changes) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, c := range changes {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(fmt.Sprintf("%s: %s → %s", c.Path, formatValue(c.Old), formatValue(c.New)))
	}
	return sb.String()
}

func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}

// jsonChange is the JSON serialization format for a single change.
type jsonChange struct {
	Path string `json:"path"`
	Old  any    `json:"old"`
	New  any    `json:"new"`
}

// FormatJSON returns a JSON representation of the changes as an array of
// objects, each with "path", "old", and "new" fields.
func FormatJSON(changes []Change) ([]byte, error) {
	out := make([]jsonChange, len(changes))
	for i, c := range changes {
		out[i] = jsonChange{Path: c.Path, Old: c.Old, New: c.New}
	}
	return json.Marshal(out)
}
