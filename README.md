# Rclone integration server #

This describes the setup for the rclone integration server, which runs
on `rclone-testing` as the `rclone` user.

It runs the rclone integration tests every night and uploads the
results to `pub.rclone.org:integration-tests`, and builds and uploads
the tip website to https://tip.rclone.org/.

## Files ##

- `integration-test.go` - checks out and builds rclone, then runs
  `test_all` against all the remotes. Run `go run integration-test.go
  -h` to see the flags, eg `-branch`, `-pr` and `-backends`.
- `upload-tip.sh` - builds the website from the rclone checkout and
  uploads it to `tip.rclone.org:`.
- `tidy-integration-test.sh` - keeps the newest 30 test runs in
  `pub.rclone.org:integration-tests` and deletes the rest.
- `update-docker-images.sh` - pulls the latest version of every docker
  image and prunes the rest.
- `crontab` - the crontab for the `rclone` user, which runs all of the
  above.

The rclone source is checked out in `~/go/src/github.com/rclone/rclone`.
This is shared by the integration tests and `upload-tip.sh`. The
restic source is also checked out in `~/go/src/github.com/restic/restic`
for the `cmd/serve/restic` tests.

Test output is written to `~/integration-test/rclone-integration-tests`
(newest 30 runs kept) and emailed to nick@craig-wood.com.

## How to install ##

The server runs Ubuntu. Install enough tools to build rclone and
run the mount tests

    apt install build-essential fuse fuse3 libfuse-dev

Install the latest version of go into `/usr/local/go` and the latest
version of hugo into `/usr/local/bin` (needed for `upload-tip.sh`).

Install docker and add the `rclone` user to the `docker` group. Many
of the test servers (FTP, SFTP, SMB, WebDAV, Swift, etc) are run in
docker by `test_all`, using the scripts in
`fstest/testserver/init.d` in the rclone source.

Add this to the `rclone` user's `.profile`

    export PATH="/usr/local/bin:$PATH"
    export GOPATH=$HOME/go
    export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"

Check out this repo into `~/integration-test` and install the crontab

    git clone https://github.com/rclone/integration-test.git ~/integration-test
    crontab ~/integration-test/crontab

Make sure `~/.rclone.conf` has

- credentials for all the remotes in `fstest/test_all/config.yaml` in
  the rclone source, eg `TestS3:`, `TestDrive:`, `TestB2:`.
- `pub.rclone.org:` for the test results.
- `tip.rclone.org:` for the tip website.

## Testing security releases ##

Security fixes must be tested before they are public, so the results
of these tests must not be uploaded to pub.rclone.org.

Passing `-repo` with anything other than `origin` to
`integration-test.go` will:

- fetch `-branch` or `-pr` from that repo (one of these is required)
- not upload the results (they are still emailed)
- write the output to `-private-output` (default
  `~/integration-test/rclone-integration-tests-private`) instead of
  `-output`
- check out `master` again when finished so `upload-tip.sh` doesn't
  publish the private docs

If the run fails part way through it won't check out `master` again,
so check `git status` in `~/go/src/github.com/rclone/rclone` before
`upload-tip.sh` runs at 06:00.

### Pushing the branch to the server ###

The rclone user on the server doesn't have GitHub credentials, so the
easiest way is to push the branch to a bare repo on the server.

Do this once:

    ssh rclone@rclone-testing git init --bare private-rclone.git

Then for each test, push from your rclone checkout (`-f` so it can be
pushed again after rebasing):

    git push -f rclone@rclone-testing:private-rclone.git v1.75-stable

Then log in and run the tests in `screen`:

    ssh rclone@rclone-testing
    screen
    cd ~/integration-test && go run integration-test.go -repo /home/rclone/private-rclone.git -branch v1.75-stable

Don't push into `~/go/src/github.com/rclone/rclone` directly: the
script deletes the local branch before fetching it.

### Fetching from a private GitHub repo ###

Alternatively you can fetch directly from a private GitHub repo (eg a
security advisory fork) using ssh agent forwarding:

    ssh -A rclone@rclone-testing
    screen
    cd ~/integration-test && go run integration-test.go -repo git@github.com:rclone/rclone-ghsa-XXXX-XXXX-XXXX.git -branch v1.75-stable

Start a **new** `screen` session after logging in: a reattached
session will have a stale `SSH_AUTH_SOCK`. The agent is only needed
for the `git fetch`, so once the log shows `git checkout` you can
detach and log out.
