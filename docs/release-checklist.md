# Release Checklist

In order to cut a new release, a few things must be done:

1. auto-generate the CHANGELOG using the provided script
2. bump version.go to the new release

For the former, you can use the following script:

    $ echo -e "$(./contrib/generate-changelog.sh $LATEST_RELEASE)\n" | cat - CHANGELOG | sponge CHANGELOG

You can find `sponge` in the `moreutils` package on Ubuntu.

This script will generate all merged changes since $LATEST_RELEASE and append it to the top of the CHANGELOG. However, this will show up as "HEAD" at the top:

    $ ./contrib/generate-changelog.sh $LATEST_RELEASE
    ### HEAD

    abc123 Some merged PR summary
    ...

You'll need to manually modify "HEAD" to show up as the latest release.
