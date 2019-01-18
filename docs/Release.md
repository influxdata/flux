### Creating Release tag

We are using semantic versioning with the format "vMajor.Minor.Patch".
We have a utility `./internal/cmd/changelog` that will generate a changelog and automatically pick the next tag based on the changes.

```sh
make release
```

