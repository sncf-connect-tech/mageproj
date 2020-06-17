# Release process

## From master branch

1. Commit all staging files and push to remote
2. Add a git tag locally
3. Generate the changelog
4. Commit the changelog and push changelog to remote
5. Move the tag locally to last commit and push tag to remote

```sh
n="1.0.0"
git tag -a v1.0.0 -m "Version $n" #2
mage changeLog #3
git commit -a -m "Changelog for v$n" #4
git push origin master #4
git tag --force v$n #5
git push origin v$n #5
```

## From another dev branch

All releases are made from master, so merge your branch, then see above.
