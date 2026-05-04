# Contributing

## How To Provide Feedback

Please [raise an issue in Github](https://github.com/opendefensecloud/apiserver-kit/issues).

## Code of Conduct

See [Code of Conduct](./CODE_OF_CONDUCT.md).

## Community Meetings (monthly)

There are currently no community meetings. Please raise an issue to reach out.

## Contributor Meetings (twice monthly)

There are currently no public contributor meetings. Please raise an issue to reach out.

## Slack

There is currently no public Slack. Please raise an issue to reach out.

## Commit Convention

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. Both PR titles and individual commit messages are validated in CI.

### Format

```
<type>(optional scope): <description>
```

### Allowed Types

| Type       | Purpose                                              |
| ---------- | ---------------------------------------------------- |
| `feat`     | A new feature                                        |
| `fix`      | A bug fix                                            |
| `docs`     | Documentation changes                                |
| `chore`    | Maintenance tasks (deps, CI config, etc.)            |
| `refactor` | Code changes that neither fix a bug nor add a feature |
| `test`     | Adding or updating tests                             |
| `ci`       | CI/CD pipeline changes                               |
| `perf`     | Performance improvements                             |
| `revert`   | Reverting a previous commit                          |

### Examples

```
feat: add OCI artifact signing support
fix(registry): handle missing manifest digest
docs: update contributing guidelines
chore(deps): update cosign to v2.5.0
refactor(transfer): extract blob streaming logic
```

### Breaking Changes

Append `!` after the type/scope to indicate a breaking change:

```
feat!: change artifact transfer API
refactor(api)!: rename TransferPolicy to SyncPolicy
```

## How To Contribute

We're always looking for contributors.

### Authoring PRs

* Code contribution - investigate a [good first issue](https://github.com/opendefensecloud/apiserver-kit/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) or anything not assigned.
* You can work on an issue without being assigned.

#### Dependencies

Dependencies increase the risk of security issues and have on-going maintenance costs.

The dependency must pass these test:

* A strong use case.
* It has an acceptable license (e.g. MIT).
* It is actively maintained.
* It has no security issues.

#### Test Policy

Changes without either unit or e2e tests are unlikely to be accepted.
See [the pull request template](https://github.com/opendefensecloud/apiserver-kit/blob/main/.github/pull_request_template.md).

### Other Contributions

* Reviewing PRs
* Responding to questions in [Github Discussions](https://github.com/opendefensecloud/apiserver-kit/discussions)

#### Reviewing PRs

Anybody can review a PR.

#### Timeliness

We encourage PR authors and reviewers to respond to change requests in a reasonable time frame.
If you're on vacation or will be unavailable, please let others know on the PR.
