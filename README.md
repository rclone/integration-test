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
memstore:
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
