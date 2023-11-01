package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(Dir string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if Dir != "" {
		cmd.Dir = Dir
	}
	cmdStr := command + " " + strings.Join(args, " ")
	log.Println("Executed command:", cmdStr)
	err := cmd.Run()
	return err
}

func BotInit() bool {
	log.Println("Bot init......")
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Init: ", err)
		return false
	}

	KTBot_DIR = dir
	PATCH_DIR = KTBot_DIR + "/patch/"
	//SMATCH_DIR = KTBot_DIR + "/smatch/"
	MAINLINE_DIR = KTBot_DIR + "/linux-master/"
	//LINUX_NEXT_DIR = KTBot_DIR + "/linux-next-master/"

	os.MkdirAll("./patch", 0777)
	os.MkdirAll("./log", 0777)

	//git clone smatch and make smatch
	// err = RunCommand(KTBot_DIR, "ls", "-l", "smatch")
	// if err != nil {
	// 	err = RunCommand("", "git", "clone", "git://repo.or.cz/smatch.git")
	// 	if err != nil {
	// 		log.Fatalf("smatch clone failed: %v", err)
	// 	}

	// 	err = RunCommand(SMATCH_DIR, "make")
	// 	if err != nil {
	// 		log.Fatalf("smatch make failed: %v", err)
	// 	}
	// }

	//download mainline
	err = RunCommand(KTBot_DIR, "ls", "-l", "linux-master")
	if err != nil {
		mainline_url := "https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/snapshot/linux-master.tar.gz"
		err = RunCommand(KTBot_DIR, "wget", mainline_url)
		if err != nil {
			log.Fatalf("Download mainline failed: %v", err)
		}
		
		err = RunCommand("", "tar", "zxvf", "linux-master.tar.gz")
		if err != nil {
			log.Fatalf("decompress mainline failed: %v", err)
		}

		err = RunCommand("", "rm", "-rf", "linux-master.tar.gz")
		if err != nil {
			log.Fatalf("delete mainline.tar.gz failed: %v", err)
		}

		//set config
		// err = RunCommand(MAINLINE_DIR, "make", "defconfig")
		// if err != nil {
		// 	log.Fatalf("mainline make defconfig failed: %v", err)
		// }

		// //smatch build
		// err = RunCommand(MAINLINE_DIR, SMATCH_DIR + "smatch_scripts/build_kernel_data.sh")
		// if err != nil {
		// 	log.Fatalf("smatch build mainline kernel failed: %v", err)
		// }

		// err = RunCommand(MAINLINE_DIR, SMATCH_DIR + "smatch_scripts/test_kernel.sh")
		// if err != nil {
		// 	log.Fatalf("smatch test mainline kernel failed: %v", err)
		// }
	}

	//download linux-next
	// err = RunCommand(KTBot_DIR, "ls", "-l", "linux-next-master")
	// if err != nil {
	// 	linux_next_url := "https://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git/snapshot/linux-next-master.tar.gz"
	// 	err = RunCommand(KTBot_DIR, "wget", linux_next_url)
	// 	if err != nil {
	// 		log.Fatalf("Download linux_next failed: %v", err)
	// 	}

	// 	err = RunCommand("", "tar", "zxvf", "linux-next-master.tar.gz")
	// 	if err != nil {
	// 		log.Fatalf("decompress linux_next failed: %v", err)
	// 	}

	// 	err = RunCommand("", "rm", "-rf", "linux-next-master.tar.gz")
	// 	if err != nil {
	// 		log.Fatalf("delete linux_next.tar.gz failed: %v", err)
	// 	}

	// 	//set config
	// 	err = RunCommand(LINUX_NEXT_DIR, "make", "defconfig")
	// 	if err != nil {
	// 		log.Fatalf("linux-next make defconfig failed: %v", err)
	// 	}

	// 	//smatch build
	// 	err = RunCommand(LINUX_NEXT_DIR, SMATCH_DIR + "smatch_scripts/build_kernel_data.sh")
	// 	if err != nil {
	// 		log.Fatalf("smatch build linux-next kernel failed: %v", err)
	// 	}

	// 	err = RunCommand(LINUX_NEXT_DIR, SMATCH_DIR + "smatch_scripts/test_kernel.sh")
	// 	if err != nil {
	// 		log.Fatalf("smatch test linux-next kernel failed: %v", err)
	// 	}
	//}

	log.Println("Bot init success!")
 	return true
}
