# Release Checklist

In order to cut a new release, a few things must be done:

1. auto-generate the CHANGELOG using the provided script
2. bump version.go and docs/installation.md to the new release
3. push a tag for the new release
4. draft a [new release](https://github.com/kelseyhightower/confd/releases/new)
5. bump version.go to the next release, appending `-dev`

For the former, you can use the following script:

    $ echo -e "$(./contrib/generate-changelog.sh v$LATEST_RELEASE)\n" | cat - CHANGELOG | sponge CHANGELOG

You can find `sponge` in the `moreutils` package on Ubuntu.

This script will generate all merged changes since $LATEST_RELEASE and append it to the top of the CHANGELOG. However, this will show up as "HEAD" at the top:

    $ ./contrib/generate-changelog.sh v$LATEST_RELEASE
    ### HEAD

    abc123 Some merged PR summary
    ...

You'll need to manually modify "HEAD" to show up as the latest release.

When drafting a new release, you must make sure that a `darwin` and `linux` build of confd have
been uploaded. If you have cross-compile support, you can use the following command to generate
those binaries:

    $ CONFD_CROSSPLATFORMS="darwin/amd64 linux/amd64" NEW_RELEASE="x.y.z"
    $ for platform in $CONFD_CROSSPLATFORMS; do \
        GOOS=${platform%/*} GOARCH=${platform##*/} ./build; \
        mv bin/confd bin/confd-$NEW_RELEASE-${platform%/*}-${platform##*/}; \
    done

You can then drag and drop these binaries into the release draft.
