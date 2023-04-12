package main

import (
	"bufio"
	"os/exec"
	"strings"
	"os"
    "fmt"
)

//find an item in a slice
func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}

//remove an item in a slice
func Remove(slice []string, val string) []string {
	j := 0
	for _, v := range slice {
		if v != val {
			slice[j] = v
			j++
		}
	}
    return slice[:len(slice) - 1]
}

//get the .c file dir and the content of a warning or error
func extract(str string) string {
	str = strings.TrimSpace(str)
	index := strings.Index(str, ":");
	file_dir := str[:index]
	content := ""
	index_1 := strings.Index(str, " ") + 1;
	if str[len(str) - 1] == ')' {
        index_2 := len(str) - 1
		for{
			if str[index_2] == '(' {
				break
			}
			index_2 = index_2 - 1
		}
		index_2 = index_2 - 1
		for{
			if str[index_2] != ' ' {
				break
			}
			index_2 = index_2 - 1
		}
		index_2 = index_2 + 1
        content =  str[index_1 : index_2]
	}else{
        content =  str[index_1 : len(str)]
	}
    return file_dir + ": " + content
}

func main(){
    //smatch, patch and smatch
    patch_dir := "/home/lsc20011130/robot/linux/0001-scsi-lpfc-fix-ioremap-issues-in-lpfc_sli4_pci_mem_se.patch" //patch file's dir
    file, err := os.OpenFile(patch_dir, os.O_RDONLY, 0666)
    if err != nil {
	    fmt.Println("Open patch file error!", err)
	    return
    }
    defer file.Close()
    buf := bufio.NewScanner(file)
    for {
        if !buf.Scan() {
            break
        }
        line := buf.Text()
        line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "diff --git a"){
        	sub_line := line[13:len(line)]
			index := strings.Index(sub_line, " ")
			//changed file's dir
        	dir := sub_line[0:index]    
			args := []string{"/home/lsc20011130/robot/cmd.py", patch_dir, dir}           
			cmd := exec.Command("python", args...)
			err := cmd.Run()
			if err != nil{
				panic(err)
			}
        }
    }

    //organization of the analysis result of smatch

    //put into slice and eliminate the duplication
    var pre_warn_slice []string
    var pre_error_slice []string
    var after_warn_slice []string
    var after_error_slice []string
    
    var pre_warn_extract_slice []string
    var pre_error_extract_slice []string
    var after_warn_extract_slice []string
    var after_error_extract_slice []string

    //logss/pre_smatch_warn.log    
    file1 := "../logss/pre_smatch_warn.log"
    pre_warn, err1 := os.OpenFile(file1, os.O_RDWR|os.O_CREATE, 0666)
    if err1 != nil {
	    fmt.Println("Open pre_smatch_warn.log error!", err1)
	    return
    }
    defer pre_warn.Close()
    buf1 := bufio.NewScanner(pre_warn)
    for {
        if !buf1.Scan() {
            break
        }
        line := buf1.Text()
        line = strings.TrimSpace(line)
        line_extract := extract(line)
        //to avoid smatch shows a same warning/error in the same line many times
        _, ex := Find(pre_warn_slice, line)
        if ex == false {
            pre_warn_slice = append(pre_warn_slice, line)
            pre_warn_extract_slice = append(pre_warn_extract_slice, line_extract)
        } 
    }

    //logss/pre_smatch_error.log
    file2 := "../logss/pre_smatch_error.log"
    pre_error, err2 := os.OpenFile(file2, os.O_RDWR|os.O_CREATE, 0666)
    if err2 != nil {
	    fmt.Println("Open pre_smatch_error.log error!", err2)
	    return
    }   
    defer pre_error.Close()
    buf2 := bufio.NewScanner(pre_error)
    for {
        if !buf2.Scan() {
            break
        }
        line := buf2.Text()
        line = strings.TrimSpace(line)
        line_extract := extract(line)
        _, ex := Find(pre_error_slice, line)
        if ex == false {
            pre_error_slice = append(pre_error_slice, line)
            pre_error_extract_slice = append(pre_error_extract_slice, line_extract)
        }
    }

    //logss/after_smatch_warn.log
    file3 := "../logss/after_smatch_warn.log"
    after_warn, err3 := os.OpenFile(file3, os.O_RDWR|os.O_CREATE, 0666)
    if err3 != nil {
	    fmt.Println("Open after_smatch_warn.log error!", err3)
	    return
    }
    defer after_warn.Close()
    buf3 := bufio.NewScanner(after_warn)
    for {
        if !buf3.Scan() {
            break
        }
        line := buf3.Text()
        line = strings.TrimSpace(line)
        line_extract := extract(line)
        _, ex := Find(after_warn_slice, line)
        if ex == false {
            after_warn_slice = append(after_warn_slice, line)
            after_warn_extract_slice = append(after_warn_extract_slice, line_extract)
        }
    }
            
    //logss/after_smatch_error.log
    file4 := "../logss/pre_smatch_error.log"
    after_error, err4 := os.OpenFile(file4, os.O_RDWR|os.O_CREATE, 0666)
    if err4 != nil {
	    fmt.Println("Open after_smatch_error.log error!", err4)
	    return
    }
    defer after_error.Close()
    buf4 := bufio.NewScanner(after_error)
    for {
        if !buf4.Scan() {
            break
        }
        line := buf4.Text()
        line = strings.TrimSpace(line)
        line_extract := extract(line)
        _, ex := Find(after_error_slice, line)
        if ex == false {
            after_error_slice = append(after_error_slice, line)
            after_error_extract_slice = append(after_error_extract_slice, line_extract)
        }
    }

    //write back into new files
    var solved_warn_slice []string
    var unsolved_warn_slice []string
    var new_warn_slice []string
    
    var solved_error_slice []string
    var unsolved_error_slice []string
    var new_error_slice []string
    

    //log_analysis_warning/warn_diff.log
    file5 := "../log_analysis_warning/warn_situation.log"
    f5, err5 := os.OpenFile(file5, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err5 != nil {
	    fmt.Println("Create warn_situation.log error!", err5)
	    return
    }
    defer f5.Close()
    write5 := bufio.NewWriter(f5)
	
    n1 := len(pre_warn_extract_slice);
    for j := 0; j < n1; j++{
        idx, ex := Find(after_warn_extract_slice, pre_warn_extract_slice[j])
        if ex == false{
            solved_warn_slice = append(solved_warn_slice, pre_warn_slice[j])
        }else{
            unsolved_warn_slice = append(unsolved_warn_slice, pre_warn_slice[j])
            after_warn_extract_slice = Remove(after_warn_extract_slice, after_warn_extract_slice[idx])
            after_warn_slice = Remove(after_warn_slice, after_warn_slice[idx])
        }
	}

    n2 := len(after_warn_slice)
    for j := 0; j < n2; j++{
        new_warn_slice = append(new_warn_slice, after_warn_slice[j])
    }

    n3 := len(solved_warn_slice)
    write5.WriteString("Solved Warnings:\n")
    write5.Flush()
    for j := 0; j < n3; j++{
        write5.WriteString(solved_warn_slice[j])
        write5.Flush()
    }

    n4 := len(unsolved_warn_slice)
    write5.WriteString("\n\nUnsolved Warnings:\n")
    write5.Flush()
    for j := 0; j < n4; j++{
        write5.WriteString(unsolved_warn_slice[j])
        write5.Flush()
    }

    n5 := len(new_warn_slice)
    write5.WriteString("\n\nNew Warnings:\n")
    write5.Flush()
    for j := 0; j < n5; j++{
        write5.WriteString(new_warn_slice[j])
        write5.Flush()
    }
            
    //log_analysis_error/error_diff.log
    file6 := "../log_analysis_error/error_situation.log"
    f6, err11 := os.OpenFile(file6, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err11 != nil {
	    fmt.Println("Create error_situation.log error!", err11)
	    return
    }
    defer f6.Close()
    write6 := bufio.NewWriter(f6)

    n1 = len(pre_error_extract_slice);
    for j := 0; j < n1; j++{
        idx, ex := Find(after_error_extract_slice, pre_error_extract_slice[j])
        if ex == false{
            solved_error_slice = append(solved_error_slice, pre_error_slice[j])
        }else{
            unsolved_error_slice = append(unsolved_error_slice, pre_error_slice[j])
            after_error_extract_slice = Remove(after_error_extract_slice, after_error_extract_slice[idx])
            after_error_slice = Remove(after_error_slice, after_error_slice[idx])
        }
    }

    n2 = len(after_error_slice)
    for j := 0; j < n2; j++{
        new_error_slice = append(new_error_slice, after_error_slice[j])
    }

    n3 = len(solved_error_slice)
    write6.WriteString("Solved Errors:\n")
    write6.Flush()
    for j := 0; j < n3; j++{
        write6.WriteString(solved_error_slice[j])
        write6.Flush()
    }

    n4 = len(unsolved_error_slice)
    write6.WriteString("\n\nUnsolved Errors:\n")
    write6.Flush()
    for j := 0; j < n4; j++{
        write6.WriteString(unsolved_error_slice[j])
        write6.Flush()
    }

    n5 = len(new_error_slice)
    write6.WriteString("\n\nNew Errors:\n")
    write6.Flush()
    for j := 0; j < n5; j++{
        write6.WriteString(new_error_slice[j])
        write6.Flush()
    }
}
