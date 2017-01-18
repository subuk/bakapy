# How to make a release

* Run tests with `make test`
* Update changelogs in debian/changelog and bakapy.spec.in files
* Commit release changelogs with message "Release X.X.X"
* Tag release commit with `git tag vX.X.X`
* Build native packages `make package-all`
* Draft a new release on github, fill it with description, attach native packages
* Publish release
