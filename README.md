work in golang 1.2
sudo apt-get install tofrodos
install smatch, coccicheck, cppcheck 

set BUILD_DIR, BOOT_DIR, MAINLINE_DIR, LINUX_NEXT_DIR, SMATCH_DIR, KT_Bot_DIR in main.go,
and all the DIRS are ended with '/'.
maybe BOOT_DIR is unneccessary, you can put the boot script in workdir
and if you do so, do not forget to modify the code of BootTest() function in checkers.go

To start the bot, run 'go build -o KTestRobot *.go', and then run ./KTestRobot
The code remains untested since last edit