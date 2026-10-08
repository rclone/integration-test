# Rclone integration server #

This describes the setup for the rclone integration server.  More details are probably needed!

## How to install ##

Install enough tools to build rclone

    apt-get build-essentials

Install the latest version of go from source.

Make sure go path is added to .profile

    export GOPATH=$HOME/go

set PATH so it includes go binary path in .profile

    export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"

Install hugo from .deb, make sure /usr/local/bin is on the path

    export PATH="/usr/local/bin:$PATH"

create an rclone user and have something like this on the crontab

```
SHELL=/bin/bash
MAILTO=your@email-address.com

0 5 * * * (cd ~/integration-test; date -Is; source ~/.profile; ./integration-test.sh) >> integration-test.log 2>&1

0 9 * * * (cd ~/integration-test; date -Is; source ~/.profile; ./upload-tip.sh) >> upload-tip.log 2>&1
```

Make sure you have an rclone config with credentials for all the cloud providers.  `rclone listremotes` should look something like

```
TestAmazonCloudDrive:
TestAzureBlob:
TestB2:
TestBox:
TestCache:
TestCryptDrive:
TestCryptSwift:
TestDrive:
TestDropbox:
TestFTP:
TestGoogleCloudStorage:
TestHubic:
TestMega:
TestOneDrive:
TestOss:
TestPcloud:
TestQingStor:
TestS3:
TestSftp:
TestSwift:
TestWebdav:
TestYandex:
```


## FTP ##

Make a new user called testdata - this will be used to run the SFTP
and FTP integration tests.  Make sure they have a very secure password
and add it to the rclone config.

Install pure-ftpd with the extra config file

    # cat /etc/pure-ftpd/conf/Bind
    127.0.0.1,21

## SSH ##

Make an ssh key for rclone user and install it in testdata user

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
