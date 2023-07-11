# KTestRobot

## Installation

### Install Go

Download the golang package (`go1.20.5.linux-amd64.tar.gz`) and execute the following command:

```
# rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.5.linux-amd64.tar.gz
```

### Install packages for kernel compilation

```
sudo apt-get install build-essential make libncurses5-dev flex libssl-dev libelf-dev bison
```

### Install other dependencies

```
sudo apt-get install git sqlite3 libsqlite3-dev
sudo apt-get install libtry-tiny-perl tofrodos cppcheck ocaml coccinelle
```

## Building

```
git clone https://gitee.com/dzm91_hust/KTestRobot
cd KTBot
go build -o KTestRobot *.go
./KTestRobot
```

**tips: The first time would spend 1 ~ 2 hours to initialize the running environment**
