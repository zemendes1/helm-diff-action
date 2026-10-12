# Helm Diff Action

Show how a Helm chart's rendered manifests change between the base branch and your feature branch. The action runs `helm template` on both branches and prints a unified diff. It needs no cluster and no Helm plugins, and runs on Linux runners (x64 and ARM64).

## Usage

```yaml
permissions:
  contents: read
  pull-requests: write

jobs:
  helm-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7.0.1
      - uses: azure/setup-helm@v5.0.1
      - uses: zemendes1/helm-diff-action@v1
        with:
          chart: charts/my-app
          values: |
            charts/my-app/values-prod.yaml
```


## Inputs

| Name           | Required | Default                         | Description                                                   |
|----------------|----------|---------------------------------|---------------------------------------------------------------|
| `chart`        | yes      |                                 | Path to the chart, relative to the repository root            |
| `values`       | no       |                                 | Newline-separated list of values files, relative to repo root |
| `base-ref`     | no       | PR base branch / default branch | Branch or commit to compare against                           |
| `args`         | no       |                                 | Extra arguments passed to `helm template`                     |
| `github-token` | no       | `${{ github.token }}`           | Token used to comment on the pull request                     |

Values files that don't exist on the base branch are skipped there. If the chart itself doesn't exist on the base branch, every manifest shows as added.

## Permissions

On pull requests, the action posts the diff as a comment, so the token needs these permissions:

| Permission      | Why                                    |
|-----------------|----------------------------------------|
| `contents`      | `read`, to check out the repository    |
| `pull-requests` | `write`, to comment on the pull request |

When a workflow sets `permissions:`, every permission it doesn't list is set to `none`, so list `contents: read` as well. To comment as a GitHub App or another account instead of `github-actions[bot]`, pass that token as `github-token`.

## Outputs

| Name      | Description                                               |
|-----------|-----------------------------------------------------------|
| `changed` | `true` if the rendered manifests differ, else `false`     |

## License

[MIT](LICENSE)
