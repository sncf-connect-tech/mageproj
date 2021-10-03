# mageproj

Opiniated utilities to use with [Mage](https://github.com/magefile/mage) in order to keep `magefile.go` in projects as [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) as possible.

## Packages
### Library: 'mgl'

A low level [Mage](https://github.com/magefile/mage) **independent** package to build its own targets from scratch.

For an example of using 'mgl', see [mageproj.go](./mgp/mageproj.go)

### Project: 'mgp'

A high level [Mage](https://github.com/magefile/mage) **dependent** package to reuse as is.

For an example of using 'mgp', see [magefile.go](./magefile.go)

## Example

### Magefile

See [magefile.go](./example/magefile.go) for an example of magefile.

### Commands

Use `-v` to enable verbose mode

* Build

```sh
$ mage clean build
===== validate
===== test
===== build

$ ls build/
build-info.json myapp
```

* ChangeLog (optional)

```sh
$ MAGEP_VERSION="1.0.0" mage changelog
File ChangeLog.md generated

# then commit changes...
```

* Release

```sh
$ MAGEP_VERSION="1.0.0" mage release
Tag v1.0.0 created and pushed to remote
```

* Deploy

```sh
$ MAGEP_ARTIFACT_PWD='mypassword' mage deploy 
===== validate
===== test
===== package
===== deploy
Uploading file: /home/anonymous/workspace/tools/myapp/build/myapp_v1.0.0_darwin-amd64.tar.gz
Received HTTP status code: 201
Uploading file: /home/anonymous/workspace/tools/myapp/build/myapp_v1.0.0_linux-amd64.tar.gz
Received HTTP status code: 201
Uploading file: /home/anonymous/workspace/tools/myapp/build/myapp_v1.0.0_windows-amd64.zip
Received HTTP status code: 201
```
