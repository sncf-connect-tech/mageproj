# Release process

## From master branch

1. Commit all staging files and push to remote
1. Add a git tag locally: `git tag -a v1.0.0 -m "Version 1.0.0"`
1. Generate the changelog: `mage changeLog`
1. Commit the changelog: `git commit -a -m "Changelog for v1.0.0"` and push changelog to remote: `git push origin master`
1. Move the tag locally to last commit: `git tag --force v1.0.0` and push tag to remote: `git push origin v1.0.0`

Gitlab-ci will then use: `mage deploy` to deploy binaries to artifactory.

## From another dev branch

All releases are made from master, so merge your branch, then see above.
