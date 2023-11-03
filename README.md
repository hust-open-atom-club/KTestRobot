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

## Configure the config.json file

Fill in the config.json file with sensitive information such as SMTP and IMAP server details, username, password, and a list of email addresses whitelist for filtering.

## Building

```
git clone https://gitee.com/dzm91_hust/KTestRobot
cd KTBot
make
./KTestRobot
```

**tips: The first time would spend 1 ~ 2 hours to initialize the running environment**

## Configuration Explanation

The `config.json` file contains the following fields:

1. `smtpServer`: The SMTP server used for sending emails. This field should be set to the address of the SMTP server.
2. `smtpPort`: The port number for the SMTP server. This field should be set to the port number used by the SMTP server.
3. `smtpUsername`: The username for authenticating with the SMTP server. This field should be set to the username used for authentication.
4. `smtpPassword`: The password for authenticating with the SMTP server. This field should be set to the password used for authentication.
5. `imapServer`: The IMAP server used for receiving emails. This field should be set to the address of the IMAP server.
6. `imapPort`: The port number for the IMAP server. This field should be set to the port number used by the IMAP server.
7. `imapUsername`: The username for authenticating with the IMAP server. This field should be set to the username used for authentication.
8. `imapPassword`: The password for authenticating with the IMAP server. This field should be set to the password used for authentication.
9. `whiteLists`: A list of email addresses that are considered as white-listed recipients. This field should be set to an array of email addresses.

