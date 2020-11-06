# Source Removal Cloud Native Buildpack

This buildpack is meant to be used at the end of the buildpack order definition
and will delete files in the application directory.

## Integration

This buildpack will always pass detection and will delete all files that are
not flagged to be included using the environment variable `$BP_INCLUDE_FILES`
which is a list of paths.
```shell
BP_INCLUDE_FILES=file/glob/*
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.
