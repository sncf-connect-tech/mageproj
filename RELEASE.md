# Release process

## From master branch

* Generate changeLog

```sh
ver="1.0.0"
$ MAGEFILEP_VERSION=$ver mage changelog
File ChangeLog.md generated

# then commit change -> Changelog for $ver
```

* Release creating a git tag

```sh
$ MAGEFILEP_VERSION=$ver mage release
Tag v1.0.0 created and pushed to remote
```

## From another dev branch

All releases are made from master, so merge your branch, then see above.
