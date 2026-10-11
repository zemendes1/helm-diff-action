# Roadmap

## v1: Marketplace release

### Test CI workflow
- [x] Run the action against the test chart in this repo (`.github/workflows/test.yml`)
- [ ] Also cover a run where the manifests change, and assert `changed` is `true`

### Post the diff as a pull request comment, creating a new comment on new pushes
- [x] Capture the diff in memory while still printing it to the job log
- [x] `formatComment`: Markdown with a header and one section per chart, diff in a ` ```diff ` block
- [ ] Tests for `formatComment`: unchanged charts, a changed chart, several charts
- [ ] Truncate the comment to GitHub's 65,536-character limit, keeping the code block closed and pointing to the job log
- [ ] Read the pull request number from the event file (`GITHUB_EVENT_PATH`); skip commenting outside pull requests
- [ ] Post the comment with the GitHub REST API using `net/http`, tested against an `httptest` server
- [ ] `action.yml`: `github-token` input (default `github.token`) and an input to turn commenting off
- [ ] `test.yml`: `pull-requests: write` permission; README: inputs and required permissions
- [ ] Check on a real PR: each push adds a new comment, pushes to `main` post nothing

### Diff several charts in one run, or only the charts a PR touches
- [ ] `chart` input accepts several newline-separated paths, with one `chartDiff` per chart
- [ ] `changed` output is `true` if any chart changed
- [ ] Optionally detect which charts a PR touches from the changed files
- [ ] Split the 65,536-character budget fairly across charts

### Per-resource diff instead of one big text diff
- [ ] Split rendered manifests into resources by kind, namespace and name
- [ ] Report added, removed and changed resources, with a diff per changed resource
- [ ] Summary table at the top of the comment
