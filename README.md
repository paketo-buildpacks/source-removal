# Source Removal Cloud Native Buildpack

This buildpack is meant to be used at the end of the buildpack order definition and will delete files in the application directory.

## Integration

The Source Removal CNB provides source-removal. source-removal can be required by generating a [Build Plan TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml) file or a `plan.toml` file that can be used with the [Build Plan CNB](https://github.com/ForestEkhardt/build-plan) that looks like the following:

```toml
[[requires]]
  name = "source-removal"
  [requires.metadata]
    keep = [
    'file/glob/*'
    ]
```


## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.
