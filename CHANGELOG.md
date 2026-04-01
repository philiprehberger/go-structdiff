# Changelog

## 0.2.2

- Standardize README to 3-badge format with emoji Support section
- Update CI checkout action to v5 for Node.js 24 compatibility
- Add GitHub issue templates, dependabot config, and PR template

## 0.2.1

- Fix README structure: remove extra Features section

## 0.2.0

- Add `Patch` function to apply changes to a struct pointer via dot-notation paths
- Add `Format` function for human-readable multi-line diff output
- Add `FormatJSON` function for JSON serialization of changes
- Add `OnlyFields` option to restrict comparison to specified fields
- Map comparison support for `map[string]any` with key-by-key diffing

## 0.1.3

- Consolidate README badges onto single line

## 0.1.1

- Add badges and Development section to README

## 0.1.0

- Initial release
- Field-level struct comparison via reflection
- `Ignore` and `IgnoreTag` options
- Dot-notation paths for nested fields
