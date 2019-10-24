# mageproj

Opiniated utilities to use with [Mage](https://github.com/magefile/mage) in order to keep `magefile.go` in projects as [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) as possible.

## Package library: 'mgl'

A low level [Mage](https://github.com/magefile/mage) **independent** package to build its own targets from scratch.

For an example of using 'mgl', see [mageproj.go](./mgp/mageproj.go)

## Package project: 'mgp'

A high level [Mage](https://github.com/magefile/mage) **dependent** package to reuse as is.

For an example of using 'mgp', see [magefile.go](./magefile.go)

