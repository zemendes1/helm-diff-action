# Roadmap

## v1: Marketplace release

### Post the diff as a pull request comment, creating a new comment on new pushes
- [ ] Truncate the comment to GitHub's 65,536-character limit, keeping the code block closed and pointing to the job log
- [ ] `action.yml`: an input to turn commenting off

### Diff several charts in one run, or only the charts a PR touches
- [ ] `chart` input accepts several newline-separated paths, with one `chartDiff` per chart
- [ ] `changed` output is `true` if any chart changed
- [ ] Optionally detect which charts a PR touches from the changed files
- [ ] Split the 65,536-character budget fairly across charts

### Per-resource diff instead of one big text diff
- [ ] Split rendered manifests into resources by kind, namespace and name
- [ ] Report added, removed and changed resources, with a diff per changed resource
