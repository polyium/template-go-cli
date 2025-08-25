# Contributing Guide

> [!NOTE]
> **The following documentation is intended for project maintainers & developers, and those looking to contribute to the project**.

## Commit Messages

```bash
git config --local commit.template ".gitmessage"
```

## Initial Setup

Install pre-requisites, ci tools, and local development hooks:

```bash
make pre-requisites
```

## Deployment

### `Makefile`

1. Prior to any releases, commit and publish all changes.
2. Install `goreleaser` and test the deployment.
    ```bash
    make pre-requisites test-release
    ```
3. Execute a release target - options include: `patch-release`, `minor-release`, `major-release`. These commands will publish various changes and perform a version bump.
    ```bash
    make patch-release
    ```

The default `Makefile` target includes the testing commands, and an additional target
to `patch-release`.

## Package Publication

_The following section refers to publishing package(s) to https://pkg.go.dev._

- See GO's [*Publishing a Module*](https://go.dev/doc/modules/publishing) for additional details.

1. Establish a [`LICENSE`](https://spdx.org/licenses/) to the project.
2. Ensure dependencies are updated.
    ```bash
    go mod tidy
    ```
3. Sync the working tree's `HEAD` with its remote.
    ```bash
    git add .
    git commit --message "<commit-msg>"
    git push --set-upstream origin main
    ```
4. Assign a tag and push.
    ```bash
    git tag "v$(head VERSION)" && git push origin "v$(head VERSION)"
    ```
5. Make the module available, publicly.
    ```bash
    GOPROXY=proxy.golang.org go list -mutex "example/package@v$(head VERSION)"
    ```

Adding the package to `pkg.go.dev` may need to be requested. Navigate to the mirror's expected url, and follow
instructions for requesting the addition.

Upon successful request, a message should be displayed:

> _We're still working on “example/package”. Check back in a few minutes!_

For any other issues, consult the [official](https://pkg.go.dev/about#adding-a-package) documentation.

## Documentation

Tool `godoc` is required to render the documentation; go's documentation rendering tool includes runnable examples if the package
has written them.

Installation Steps:

1. Install `godoc`.
    ```bash
    go install golang.org/x/tools/cmd/godoc@latest
    ```
1. Backup shell profile and update `PATH`.
    ```bash
    cp ~/.zshrc ~/.zshrc.bak
    printf "export PATH=\"\${PATH}:%s\"\n" "$(go env --json | jq -r ".GOPATH")/bin" >> ~/.zshrc
    source ~/.zshrc
    ```
1. Start the `godoc` server.
    ```bash
    godoc -http=:6060
    ```
1. Open the webpage.
    ```bash
    open "http://localhost:6060/pkg/"
    ```
