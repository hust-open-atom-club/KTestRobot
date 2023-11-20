package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// TODO: execute the command in the current directory
func RunCommand(Dir string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	if Dir != "" {
		cmd.Dir = Dir
	}
	cmdStr := command + " " + strings.Join(args, " ")
	log.Println("Executed command:", cmdStr)
	err := cmd.Run()
	return err
}

func (mailinfo MailInfo) botInit(KTBot_DIR string) bool {
	log.Println("Kernel Testing Robot is initializing......")

	os.MkdirAll("./patch", 0777)
	os.MkdirAll("./log", 0777)

	//TODO: we need to update these repositories
	// download mainline if not provided
	err := RunCommand(KTBot_DIR, "ls", "-l", "mainline")
	if err != nil {
		mainline_url := "https://mirrors.hust.college/git/linux.git"
		err = RunCommand(KTBot_DIR, "git", "clone", "--depth=1", mainline_url)
		if err != nil {
			log.Fatalf("Download mainline failed: %v", err)
		}
		err = RunCommand(KTBot_DIR, "mv", "linux", "mainline")
		if err != nil {
			log.Fatalf("Remove filename failed: %v", err)
		}
		err = RunCommand(filepath.Join(KTBot_DIR, "mainline"), "make", "allyesconfig")
		if err != nil {
			log.Fatalf("Failed to configure config: %v", err)
		}
		err = RunCommand(filepath.Join(KTBot_DIR, "mainline"), "make", "-j" + strconv.Itoa(mailinfo.Procs))
		if err != nil {
			log.Fatalf("Compilation failed: %v", err)
		}
		
	}

	// download linux-next if not provided
	err = RunCommand(KTBot_DIR, "ls", "-l", "linux-next")
	if err != nil {
		linux_next_url := "https://mirrors.hust.college/git/linux-next.git"
		err = RunCommand(KTBot_DIR, "git", "clone", "--depth=1", linux_next_url)
		if err != nil {
			log.Fatalf("Download linux_next failed: %v", err)
		}
		err = RunCommand(filepath.Join(KTBot_DIR, "linux-next"), "make", "allyesconfig")
		if err != nil {
			log.Fatalf("Failed to configure config: %v", err)
		}
		err = RunCommand(filepath.Join(KTBot_DIR, "linux-next"), "make", "-j" + strconv.Itoa(mailinfo.Procs))
		if err != nil {
			log.Fatalf("Compilation failed: %v", err)
		}
		
	}

	// clone smatch and compile
	err = RunCommand(KTBot_DIR, "ls", "-l", "smatch")
	if err != nil {
		err = RunCommand(KTBot_DIR, "git", "clone", "git://repo.or.cz/smatch.git")
		if err != nil {
			log.Fatalf("smatch clone failed: %v", err)
		}

		err = RunCommand(filepath.Join(KTBot_DIR, "smatch"), "make")
		if err != nil {
			log.Fatalf("smatch make failed: %v", err)
		}
	}

	// TODO: this global cross-function database should be established at initialization
	// err = RunCommand(KTBot_DIR, "ls", "-l", "linux-master")
	// if err != nil {
	// 	//smatch build
	// 	err = RunCommand(MAINLINE_DIR, SMATCH_DIR + "smatch_scripts/build_kernel_data.sh")
	// 	if err != nil {
	// 		log.Fatalf("smatch build mainline kernel failed: %v", err)
	// 	}

	// 	err = RunCommand(MAINLINE_DIR, SMATCH_DIR + "smatch_scripts/test_kernel.sh")
	// 	if err != nil {
	// 		log.Fatalf("smatch test mainline kernel failed: %v", err)
	// 	}
	// }

	log.Println("The initializtion of Kernel Testing Robot is done!")
	return true
}

func (mailinfo MailInfo) update(KTBot_DIR string) bool {
	log.Println("Kernel Testing Robot is updating......")
	err := RunCommand(filepath.Join(KTBot_DIR, "mainline"), "git", "pull")
	if err != nil {
		log.Println("Update mainline failed: ", err)
	}
	
	err = RunCommand(filepath.Join(KTBot_DIR, "linux-next"), "git", "pull")
	if err != nil {
		log.Println("Update linux_next failed: ", err)
		log.Println("Start cloning linux-next again...")
		err = RunCommand(KTBot_DIR, "rm", "-rf", "linux-next")
		if err != nil {
			log.Println("Failed to delete current linux-next folder!")
		}
		err = RunCommand(KTBot_DIR, "ls", "-l", "linux-next")
		if err != nil {
			linux_next_url := "https://mirrors.hust.college/git/linux-next.git"
			err = RunCommand(KTBot_DIR, "git", "clone", "--depth=1", linux_next_url)
			if err != nil {
				log.Fatalf("Download linux_next failed: %v", err)
			}
			err = RunCommand(filepath.Join(KTBot_DIR, "linux-next"), "make", "allyesconfig")
			if err != nil {
				log.Fatalf("Failed to configure config: %v", err)
			}
			err = RunCommand(filepath.Join(KTBot_DIR, "linux-next"), "make", "-j" + strconv.Itoa(mailinfo.Procs))
			if err != nil {
				log.Fatalf("Compilation failed: %v", err)
			}	
		}
	}

	err = RunCommand(filepath.Join(KTBot_DIR, "smatch"), "git", "pull")
	if err != nil {
		log.Println("smatch update failed: ", err)
	}
	return true
}
