# Roadmap

## v1: Marketplace release

### Diff every chart under a directory
- [ ] Find charts under a directory: every directory with a `Chart.yaml` on the base ref or on HEAD, so new and deleted charts are included
- [ ] Run `helmDiff` once per chart, with one `chartDiff` per chart, fetching the base ref only once
- [ ] `changed` output is `true` if any chart changed
- [ ] `values` paths are relative to each chart's directory and skipped where missing; move `testdata/values-prod.yaml` into the test chart
- [ ] `action.yml`: replace `chart` with `charts-dir` (a single chart's directory still works); update the README
- [ ] `test.yml`: diff a directory with two charts where only one changes
