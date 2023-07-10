package main

import (
	"log"
	"os"
	"os/exec"
)

func BotInit() bool {
	log.Println("Bot init......")
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Init: ", err)
		return false
	}

	KTBot_DIR = dir
	PATCH_DIR = KTBot_DIR + "/patch/"
	SMATCH_DIR = KTBot_DIR + "/smatch/"
	MAINLINE_DIR = KTBot_DIR + "/linux-master/"
	LINUX_NEXT_DIR = KTBot_DIR + "/linux-next-master/"

	os.MkdirAll("./patch", 0777)
	os.MkdirAll("./log", 0777)

	//git clone smatch and make smatch
	cmd := exec.Command("ls", "-l", "smatch")
	cmd.Dir = KTBot_DIR
	err = cmd.Run()
	if err != nil {
		cmd = exec.Command("git", "clone", "git://repo.or.cz/smatch.git")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch clone failed: %v", err)
		}

		cmd = exec.Command("make")
		cmd.Dir = SMATCH_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch make failed: %v", err)
		}
	}

	//download mainline
	cmd = exec.Command("ls", "-l", "linux-master")
	cmd.Dir = KTBot_DIR
	err = cmd.Run()
	if err != nil {
		mainline_url := "https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/snapshot/linux-master.tar.gz"
		cmd = exec.Command("wget", mainline_url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = KTBot_DIR
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Download mainline failed: %v", err)
		}

		cmd = exec.Command("tar", "zxvf", "linux-master.tar.gz")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("decompress mainline failed: %v", err)
		}

		cmd = exec.Command("rm", "-rf", "linux-master.tar.gz")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("delete mainline.tar.gz failed: %v", err)
		}

		//set config
		cmd = exec.Command("make", "defconfig")
		cmd.Dir = MAINLINE_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("mainline make defconfig failed: %v", err)
		}

		//smatch build
		cmd = exec.Command(SMATCH_DIR + "smatch_scripts/build_kernel_data.sh")
		cmd.Dir = MAINLINE_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch build mainline kernel failed: %v", err)
		}

		cmd = exec.Command(SMATCH_DIR + "smatch_scripts/test_kernel.sh")
		cmd.Dir = MAINLINE_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch test mainline kernel failed: %v", err)
		}
	}

	//download linux-next
	cmd = exec.Command("ls", "-l", "linux-next-master")
	cmd.Dir = KTBot_DIR
	err = cmd.Run()
	if err != nil {
		linux_next_url := "https://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git/snapshot/linux-next-master.tar.gz"
		cmd = exec.Command("wget", linux_next_url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Download linux_next failed: %v", err)
		}

		cmd = exec.Command("tar", "zxvf", "linux-next-master.tar.gz")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("decompress linux_next failed: %v", err)
		}

		cmd = exec.Command("rm", "-rf", "linux-next-master.tar.gz")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("delete linux_next.tar.gz failed: %v", err)
		}

		//set config
		cmd = exec.Command("make", "defconfig")
		cmd.Dir = LINUX_NEXT_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("linux-next make defconfig failed: %v", err)
		}

		//smatch build
		cmd = exec.Command(SMATCH_DIR + "smatch_scripts/build_kernel_data.sh")
		cmd.Dir = LINUX_NEXT_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch build linux-next kernel failed: %v", err)
		}

		cmd = exec.Command(SMATCH_DIR + "smatch_scripts/test_kernel.sh")
		cmd.Dir = LINUX_NEXT_DIR
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("smatch test linux-next kernel failed: %v", err)
		}
	}

	log.Println("Bot init success!")
 	return true
}