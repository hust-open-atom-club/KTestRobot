# KTestRobot

## Installation

### Install Go

Download the golang package (`go1.20.5.linux-amd64.tar.gz`) and execute the following command:

```
# rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.5.linux-amd64.tar.gz
```

<!-- ### Install packages for kernel compilation

```
sudo apt-get install build-essential make libncurses5-dev flex libssl-dev libelf-dev bison
```

### Install other dependencies

```
sudo apt-get install git sqlite3 libsqlite3-dev
sudo apt-get install libtry-tiny-perl tofrodos cppcheck ocaml coccinelle
``` -->

## Complete config.json

Fill in the config.json file with sensitive information such as SMTP and IMAP server details, username, password, and a list of email addresses whitelist for filtering.

## Building

```
git clone https://gitee.com/dzm91_hust/KTestRobot
cd KTBot
make
./KTestRobot -config config.json
```

## Configuration Explanation

The `config.json` file contains the following fields:

- `username`: The email account used to receive kernel patches. This field should be set to a email address, e.g., "ktestrobot@126.com"
- `password`: The password used to log in the mail server. This field should be set to the email password
- `whiteLists`: A list of email addresses that are considered as white-listed recipients. This field should be set to an array of email addresses

