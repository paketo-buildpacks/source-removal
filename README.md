# Entrypoint Cloud Native Buildpack

This buildpack is meant to be used at the end of the buildpack order definition and will delete everything in the application directory.

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.
