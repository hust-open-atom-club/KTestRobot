package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"fmt"
	"crypto/sha256"
	"strings"
	"log"
	"strconv"
)

func (mailinfo MailInfo) Checkpatch_template(KTBot_DIR string, result string, branch string, patchname string, changedpath string) {
	//build check and static analysis
	log.Println("Start BuildCheck in ", branch, ".")
	buildcheck_pass, buildcheck := mailinfo.BuildCheck(filepath.Join(KTBot_DIR, branch))
	result += buildcheck
	log.Println("BuildCheck done.")
	if buildcheck_pass {
		log.Println("Start StaticAnalysis in ", branch, ".")
		staticres:= StaticAnalysis(KTBot_DIR, branch, patchname, changedpath)
		result += staticres
		log.Println("StaticAnalysis  done.")
	}
}

func (mailinfo MailInfo) CheckPatchAll(KTBot_DIR string, patchname string, changedpath string) string {
	var result string
	log.Println("Start CheckPatchpl.")
	checkpatch_pass, checkpatch := CheckPatchpl(KTBot_DIR, patchname)
	result += checkpatch
	log.Println("CheckPatchpl done.")
	if checkpatch_pass {
		log.Println("Start ApplyPatch check.")
		apply_pass1, apply_res1 := ApplyPatch(KTBot_DIR, "linux-next", patchname)
		result += apply_res1
		apply_pass2, apply_res2 := ApplyPatch(KTBot_DIR, "mainline", patchname)
		result += apply_res2
		log.Println("ApplyPatch check done.")
		if apply_pass1 {
			mailinfo.Checkpatch_template(KTBot_DIR, result, "linux-next", patchname, changedpath)
		} else {
			if apply_pass2 {
				mailinfo.Checkpatch_template(KTBot_DIR, result, "mainline", patchname, changedpath)
			}
		}	
	}

	/* boot
		booterr, boot := BootTest()
		if booterr {
			result += boot
		}
	*/

	return result
}

func (mailinfo MailInfo) BuildCheck(Dir string) (bool, string) {
	flag := true
	result := "*** BuildCheck          PASS ***\n"
	cmd := exec.Command("make", "-j" + strconv.Itoa(mailinfo.Procs))
	if Dir != "" {
		cmd.Dir = Dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	errStr := stderr.String()
	if err != nil {
		flag = false
		result = "*** BuildCheck          FAILED ***\n"
		res := errStr + "\n"
		result += res
	}
	return flag, result
}

func StaticAnalysis(KTBot_DIR string, branch string, patchname string, changedpath string) string {
	result := ""
	smatch_err, checksmatch := CheckSmatch(KTBot_DIR, branch, patchname, changedpath)
	if smatch_err {
		result += checksmatch
	}
	cocci_err, cocci := CheckCocci(KTBot_DIR, branch, patchname, changedpath)
	if cocci_err {
		result += cocci
	}
	cppcheck_err, cppcheck := CheckCppcheck(KTBot_DIR, branch, patchname, changedpath)
	if cppcheck_err {
		result += cppcheck
	}
	return result
}

func CheckPatchpl(KTBot_DIR string, patch string) (bool, string) {
	flag := true
	result := "*** CheckPatch          PASS ***\n"
	cmd := exec.Command(filepath.Join(KTBot_DIR, "mainline", "scripts", "checkpatch.pl"), filepath.Join(KTBot_DIR, "patch", patch))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errStr := stderr.String()
	outStr := stdout.String()
	if err != nil {
		flag = false
		result = "*** CheckPatch          FAILED ***\n"
		res := outStr + "\n" + errStr + "\n"
		result += res
	}
	return flag, result
}

func ApplyPatch(KTBot_DIR string, branch string, patchname string) (bool, string) {
	flag := true
	result := "*** ApplyTo" + branch + "\tPASS ***\n"
	cmd := exec.Command("git", "apply", "--check", filepath.Join(KTBot_DIR, "patch", patchname))
	var stderr bytes.Buffer
	// cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = filepath.Join(KTBot_DIR, branch)
	err := cmd.Run()
	errStr := stderr.String()
	if err != nil {
		flag = false
		result = "*** ApplyTo" + branch + "\tFAILED ***\n"
		result += errStr + "\n"
	}
	return flag, result
}

func BugHash(info string) string {
	salt := "KTestRobot"
	hashres := sha256.Sum256([]byte(info + salt))
	res := fmt.Sprintf("%x", hashres)
	return res
}

func CheckCocci(KTBot_DIR string, branch string, patchname string, changedpath string) (bool, string) {
	patch := filepath.Join(KTBot_DIR, "patch", patchname)
	dir := ""
	switch branch {
	case "mainline":
		dir = filepath.Join(KTBot_DIR, "mainline")
	case "linux-next":
		dir = filepath.Join(KTBot_DIR, "linux-next")
	}
	result := "*** CheckCocci          PASS ***\n"
	flag := 0
	paths := strings.Split(changedpath, "\n")
	for _, path := range paths {
		end := strings.Index(path, ".c")
		if end == -1 {
			continue
		}

		cmdargs := []string{"C=1", "CHECK=scripts/coccicheck", path[:end] + ".o"}
		cocci := exec.Command("make", cmdargs...)
		// cocci.Stdout
		var stdout, stderr bytes.Buffer
		cocci.Stdout = &stdout
		cocci.Stderr = &stderr
		cocci.Dir = dir
		cocci.Run()
		errStr := stderr.String()
		outStr := stdout.String()

		//apply the patch
		apply := exec.Command("git", "apply", patch)
		apply.Dir = dir
		apply.Run()
		//check again
		cocci1 := exec.Command("make", cmdargs...)
		var stdout1, stderr1 bytes.Buffer
		cocci1.Stdout = &stdout1
		cocci1.Stderr = &stderr1
		cocci1.Dir = dir
		cocci1.Run()
		errStr1 := stderr.String()
		outStr1 := stdout.String()

		unapply := exec.Command("git", "apply", "-R", patch)
		unapply.Dir = dir
		unapply.Run()

		unsolved_warning, unsolved_error, new_warning, new_error := Logcmp(outStr, outStr1, "WARNING", "ERROR")
		if len(new_error) != 0 || len(new_warning) != 0 || len(unsolved_error) != 0 || len(unsolved_warning) != 0 {
			if flag == 0 {
				flag = 1
				result = "*** CheckCocci          FAILED ***\n"
			}

			if len(unsolved_error) != 0 {
				result += "Unsolved error: \n"
				for _, uerr := range unsolved_error {
					result += uerr + "\n"
				}
			}
			if len(unsolved_warning) != 0 {
				result += "Unsolved warning: \n"
				for _, uwarn := range unsolved_warning {
					result += uwarn + "\n"
				}
			}
			if len(new_error) != 0 {
				result += "New error: \n"
				for _, nerr := range new_error {
					result += nerr + "\n"
				}
			}
			if len(new_warning) != 0 {
				result += "New warning: \n"
				for _, nwarn := range new_warning {
					result += nwarn + "\n"
				}
			}
		}
		if BugHash(errStr) != BugHash(errStr1) {
			if flag == 0 {
				flag = 1
				result = "*** CheckCocci          FAILED ***\n"
			}
			result += "\n" + errStr1 + "\n"
		}
	}
	return true, result
}

// compare prelog and afterlog, return new_warning and new_error
func Logcmp(pre string, after string, swarn string, serr string) ([]string, []string, []string, []string) {
	pre_slice := strings.Split(pre, "\n")
	after_slice := strings.Split(after, "\n")
	var after_warning, after_error []string
	var pre_warning_spf, pre_error_spf []string
	var after_warning_spf, after_error_spf []string

	var new_warning, new_error []string
	var unsolved_warning, unsolved_error []string

	//process prelog
	for _, line := range pre_slice {
		if strings.Contains(line, swarn) {
			end_1 := strings.Index(line, ":")
			line_pre := line[0:end_1]
			start_2 := strings.Index(line, " ") + 1
			var end_2 int
			if strings.Contains(line, "on lines:") {
				end_2 = strings.LastIndex(line, ":")
			} else {
				end_2 = len(line)
			}
			line_after := line[start_2:end_2]

			spf_line := line_pre + line_after
			pre_warning_spf = append(pre_warning_spf, spf_line)

		} else if strings.Contains(line, serr) {
			end_1 := strings.Index(line, ":")
			line_pre := line[0:end_1]
			start_2 := strings.Index(line, " ") + 1
			var end_2 int
			if strings.Contains(line, "on lines:") {
				end_2 = strings.LastIndex(line, ":")
			} else {
				end_2 = len(line)
			}
			line_after := line[start_2:end_2]

			spf_line := line_pre + line_after
			pre_error_spf = append(pre_error_spf, spf_line)
		}
	}
	//process afterlog
	for _, line := range after_slice {
		if strings.Contains(line, swarn) {
			after_warning = append(after_warning, line)

			end_1 := strings.Index(line, ":")
			line_pre := line[0:end_1]
			start_2 := strings.Index(line, " ") + 1
			var end_2 int
			if strings.Contains(line, "on lines:") {
				end_2 = strings.LastIndex(line, ":")
			} else {
				end_2 = len(line)
			}
			line_after := line[start_2:end_2]

			spf_line := line_pre + line_after
			after_warning_spf = append(after_warning_spf, spf_line)
		} else if strings.Contains(line, serr) {
			after_error = append(after_error, line)

			end_1 := strings.Index(line, ":")
			line_pre := line[0:end_1]
			start_2 := strings.Index(line, " ") + 1
			var end_2 int
			if strings.Contains(line, "on lines:") {
				end_2 = strings.LastIndex(line, ":")
			} else {
				end_2 = len(line)
			}
			line_after := line[start_2:end_2]

			spf_line := line_pre + line_after
			after_error_spf = append(after_error_spf, spf_line)
		}
	}
	//compare
	for i, line := range after_warning_spf {
		_, ex := Find(pre_warning_spf, line)
		if !ex {
			new_warning = append(new_warning, after_warning[i])
		} else {
			unsolved_warning = append(unsolved_warning, after_warning[i])
		}
	}
	for j, line := range after_error_spf {
		_, ex := Find(pre_error_spf, line)
		if !ex {
			new_error = append(new_error, after_error[j])
		} else {
			unsolved_error = append(unsolved_error, after_error[j])
		}
	}
	return unsolved_warning, unsolved_error, new_warning, new_error
}

func CheckCppcheck(KTBot_DIR string, branch string, patchname string, changedpath string) (bool, string) {
	patch := filepath.Join(KTBot_DIR, "patch", patchname)
	dir := ""
	switch branch {
	case "mainline":
		dir = filepath.Join(KTBot_DIR, "mainline")
	case "linux-next":
		dir = filepath.Join(KTBot_DIR, "linux-next")
	}
	paths := strings.Split(changedpath, "\n")
	result := "*** CheckCppcheck\tPASS ***\n"
	flag := 0
	for _, path := range paths {
		if !strings.Contains(path, ".c") {
			continue
		}

		cppcheck := exec.Command("cppcheck", path)
		cppcheck.Dir = dir
		var stdout, stderr bytes.Buffer
		cppcheck.Stdout = &stdout
		cppcheck.Stderr = &stderr
		err := cppcheck.Run()
		errStr := stderr.String()
		outStr := stdout.String()
		if err != nil {
			log.Println("CheckCppcheck check pre: ", err)
		}

		//apply the patch
		apply := exec.Command("git", "apply", patch)
		apply.Dir = dir
		apply.Run()
		//check again
		checkagain := exec.Command("cppcheck", path)
		checkagain.Dir = dir
		var stdout1, stderr1 bytes.Buffer
		checkagain.Stdout = &stdout1
		checkagain.Stderr = &stderr1
		err1 := checkagain.Run()
		errStr1 := stderr.String()
		outStr1 := stdout.String()

		if err1 != nil {
			log.Println("Cppcheck check after: ", err1)
		}
		unapply := exec.Command("git", "apply", "-R", patch)
		unapply.Dir = dir
		unapply.Run()

		unsolved_warning, unsolved_error, new_warning, new_error := Logcmp(errStr+outStr, errStr1+outStr1, "warn", "error")
		if len(new_error) != 0 || len(new_warning) != 0 || len(unsolved_error) != 0 || len(unsolved_warning) != 0 {
			if flag == 0 {
				flag = 1
				result = "*** CheckCppcheck\tFAILED ***\n"
			}

			if len(unsolved_error) != 0 {
				result += "Unsolved error: \n"
				for _, uerr := range unsolved_error {
					result += uerr + "\n"
				}
			}
			if len(unsolved_warning) != 0 {
				result += "Unsolved warning: \n"
				for _, uwarn := range unsolved_warning {
					result += uwarn + "\n"
				}
			}
			if len(new_error) != 0 {
				result += "New error: \n"
				for _, nerr := range new_error {
					result += nerr + "\n"
				}
			}
			if len(new_warning) != 0 {
				result += "New warning: \n"
				for _, nwarn := range new_warning {
					result += nwarn + "\n"
				}
			}
		}
		
	}
	return true, result
}

// find an item in a slice
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func CheckSmatch(KTBot_DIR string, branch string, patchname string, changedpath string) (bool, string) {
    patch := filepath.Join(KTBot_DIR, "patch", patchname)
	dir := ""
	switch branch {
	case "mainline":
		dir = filepath.Join(KTBot_DIR, "mainline")
	case "linux-next":
		dir = filepath.Join(KTBot_DIR, "linux-next")
	}

    result := "*** CheckSmatch         PASS ***\n"
    flag := 0
    paths := strings.Split(changedpath, "\n")
    for _, path := range paths {
        if !strings.Contains(path, ".c") {
            continue
        }

        precheck := exec.Command(filepath.Join(KTBot_DIR, "smatch", "smatch_scripts", "kchecker"), path)
        precheck.Dir = dir
        var stdout, stderr bytes.Buffer
        precheck.Stdout = &stdout
        precheck.Stderr = &stderr
        precheck.Run()
        errStr := stderr.String()
        outStr := stdout.String()

        apply := exec.Command("git", "apply", patch)
        apply.Dir = dir
        apply.Run()

        checkagain := exec.Command(filepath.Join(KTBot_DIR, "smatch", "smatch_scripts", "kchecker"), path)
        checkagain.Dir = dir
        var stdout1, stderr1 bytes.Buffer
        checkagain.Stdout = &stdout1
        checkagain.Stderr = &stderr1
        checkagain.Run()
        errStr1 := stderr1.String()
        outStr1 := stdout1.String()

        unapply := exec.Command("git", "apply", "-R", patch)
		unapply.Dir = dir
		unapply.Run()

		unsolved_warning, unsolved_error, new_warning, new_error := Logcmp(outStr, outStr1, "warn", "error")
		if len(new_error) != 0 || len(new_warning) != 0 || len(unsolved_error) != 0 || len(unsolved_warning) != 0 {
			if flag == 0 {
				flag = 1
				result = "*** CheckSmatch         FAILED ***\n"
			}

			if len(unsolved_error) != 0 {
				result += "Unsolved error: \n"
				for _, uerr := range unsolved_error {
					result += uerr + "\n"
				}
			}
			if len(unsolved_warning) != 0 {
				result += "Unsolved warning: \n"
				for _, uwarn := range unsolved_warning {
					result += uwarn + "\n"
				}
			}
			if len(new_error) != 0 {
				result += "New error: \n"
				for _, nerr := range new_error {
					result += nerr + "\n"
				}
			}
			if len(new_warning) != 0 {
				result += "New warning: \n"
				for _, nwarn := range new_warning {
					result += nwarn + "\n"
				}
			}
		}
		if BugHash(errStr) != BugHash(errStr1) {
			if flag == 0 {
				flag = 1
				result = "*** CheckSmatch         FAILED ***\n"
			}
			result += "\n" + errStr1 + "\n"
		}
	}
	return true, result
}


//Boot Test, bzImage has built in BuildTest
//return false means internal error, true means test done
// func BootTest() (bool, string) {
// 	boot := exec.Command("./boot")
// 	boot.Dir = BOOT_DIR
// 	//set PGID, will be used to kill boot process and child process
// 	boot.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
// 	boot.Dir = BOOT_DIR

// 	stdout, err := boot.StdoutPipe()
// 	if err != nil {
// 		log.Println("BootTest stdoutpipe: ", err)
// 		return false, ""
// 	}
// 	stdin, inerr := boot.StdinPipe()
// 	if inerr != nil {
// 		log.Println("BootTest stdinpipe: ", inerr)
// 		return false, ""
// 	}

// 	result := "*** BootTest\tFAILED ***\n"
// 	boot.Start()

// 	reader := bufio.NewReader(stdout)

// 	var count = 1
// 	for {
// 		line, err2 := reader.ReadString('\n')
// 		if err2 != nil || io.EOF == err2 || count > 1000 {
// 			break
// 		}
// 		log.Print(line)
// 		count++
// 		if strings.HasPrefix(line, "Boot took") {
// 			result = "*** BootTest\tPASS ***\n"
// 			// log.Println("boot success, count: ", count)
// 			io.WriteString(stdin, "exit \n")
// 			return true, result
// 		}
// 	}
// 	//Kill the process
// 	syscall.Kill(-boot.Process.Pid, syscall.SIGKILL)

// 	return true, result
// }
