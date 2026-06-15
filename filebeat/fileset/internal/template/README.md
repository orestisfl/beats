# Forked `text/template` (dead-code-elimination safe)

This package is a verbatim copy of the Go standard library `text/template`
package (and the one `internal/fmtsort` helper it depends on, under
`fmtsort/`), with a single behavioral change: it does **not** resolve template
field names to Go methods.

## Why

The standard `text/template.(*state).evalField` resolves `{{ .Foo }}` against a
method named `Foo` via `reflect.Value.MethodByName` with a non-constant name.
Any reachable call to `reflect.Value.Method`/`MethodByName` (with a non-constant
argument) forces the Go linker to retain **every exported method of every
reachable type** in the binary, disabling method-level dead-code elimination
(DCE) for the whole program. For Beats this inflates binary size significantly
and, more importantly, blocks the DCE work needed elsewhere (e.g. the script
processor / goja).

Beats only uses templates to render module config and ingest-pipeline files
over plain decoded data (`map[string]interface{}`, slices, scalars) plus a small
set of registered template functions. It never invokes Go methods on template
values, so dropping method resolution is behavior-preserving for our usage.

This mirrors the approach taken by the Datadog Agent
(https://www.datadoghq.com/blog/engineering/agent-go-binaries/) and tracks the
upstream proposal https://github.com/golang/go/issues/72895.

## The modification

In `exec.go`, `evalField` no longer attempts `ptr.MethodByName(fieldName)`;
field names resolve only to struct fields or map keys. The change is marked with
a `NOTE(elastic):` comment.

## Updating

Re-copy the files from the matching Go release and re-apply the single
`evalField` edit. Run `whydeadcode` (see the repo's DCE notes) to confirm
`text/template.(*state).evalField` is no longer reachable.

## License

Upstream Go source, BSD-3-Clause. See `LICENSE`.
