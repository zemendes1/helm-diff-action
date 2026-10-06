# Helm Diff Action

Show how a Helm chart's rendered manifests change between the base branch and your feature branch. The action runs `helm template` on both branches and prints a unified diff. It needs no cluster and no Helm plugins.

## Usage

```yaml
jobs:
  helm-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-helm@v4
      - uses: <owner>/helm-diff-action@v1
        with:
          chart: charts/my-app
          values: |
            charts/my-app/values-prod.yaml
```


## Inputs

| Name       | Required | Default                       | Description                                                   |
|------------|----------|-------------------------------|---------------------------------------------------------------|
| `chart`    | yes      |                               | Path to the chart, relative to the repository root            |
| `values`   | no       |                               | Newline-separated list of values files, relative to repo root |
| `base-ref` | no       | PR base branch / default branch | Branch or commit to compare against                         |
| `args`     | no       |                               | Extra arguments passed to `helm template`                     |

Values files that don't exist on the base branch are skipped there. If the chart itself doesn't exist on the base branch, every manifest shows as added.

## Outputs

| Name      | Description                                               |
|-----------|-----------------------------------------------------------|
| `changed` | `true` if the rendered manifests differ, else `false`     |

## License

[MIT](LICENSE)
