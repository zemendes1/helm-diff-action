# Roadmap

## v1: Marketplace release

### Test CI workflow
- [x] Run the action against the test chart in this repo (`.github/workflows/test.yml`)
- [ ] Also cover a run where the manifests change, and assert `changed` is `true`
- [ ] Run `go test` in the pre-commit hooks, not only `gofmt` and `go vet`

### Post the diff as a pull request comment, creating a new comment on new pushes
- [x] Capture the diff in memory while still printing it to the job log
- [x] `formatComment`: Markdown with a header and one section per chart, diff in a ` ```diff ` block
- [x] Tests for `formatComment`: unchanged charts, a changed chart, several charts
- [ ] Test `formatComment` with a diff that doesn't end in a newline
- [ ] Truncate the comment to GitHub's 65,536-character limit, keeping the code block closed and pointing to the job log
- [x] `getPrNumber`: read the pull request number from the event file (`GITHUB_EVENT_PATH`), 0 outside pull requests
- [ ] Tests for `getPrNumber` with fixtures in `testdata/`: pull request event, push event, missing file
- [ ] Don't fail when `GITHUB_EVENT_PATH` or `GITHUB_OUTPUT` is unset, so the program can run locally
- [x] `run()` calls `postPRComment` only on pull requests, reading `GITHUB_API_URL`, `GITHUB_REPOSITORY` and `GITHUB_TOKEN` only then
- [ ] `postPRComment`: send `POST /repos/{repo}/issues/{number}/comments` with `net/http`, error on any status other than 201
- [ ] Test `postPRComment` against an `httptest` server: method, path, headers and body
- [ ] Decide whether to comment when no chart changed
- [x] `action.yml`: `github-token` input (default `github.token`) passed as `GITHUB_TOKEN`
- [ ] `action.yml`: an input to turn commenting off
- [x] `test.yml`: `pull-requests: write` permission; README: inputs and required permissions
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
