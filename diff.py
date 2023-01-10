import git
import os
from unidiff import PatchSet
from io import StringIO

#get the .c file dir and the content of a warning or error, to avoid the line number's influence 
def extract(str):
    right = len(str) - 1
    while(str[right] == ' '):
        right = right - 1
    str = str[0 : right]
    file_dir = str[0 : str.find(':')]
    content = ""
    #index_1 is the right of the first ' '
    index_1 = str.find(' ') + 1 
    #if the warning/error is ended as "(see line xxx)"
    if str[len(str) - 1] == ')':
        index_2 = len(str) - 1
        while str[index_2] != '(':
            index_2 = index_2 - 1
        index_2 = index_2 - 1
        while str[index_2] == ' ':
            index_2 = index_2 - 1
        index_2 = index_2 + 1
        content =  str[index_1 : index_2]
    else:
        content =  str[index_1 : len(str)]
    return file_dir + ': ' + content
    

repo_directory_address = "."
repository = git.Repo(repo_directory_address)

for i in range(21, 41):
    commit_po = repository.head.commit
    commit_sha1_po = str(commit_po)[0:9]
    print(commit_sha1_po)
    commit_pr = repository.commit(commit_sha1_po + "~1")
    commit_sha1_pr = str(commit_pr)[0:9]
    print(commit_sha1_pr)
    uni_diff_text = repository.git.diff(commit_sha1_pr, commit_sha1_po,
                                        ignore_blank_lines=True, 
                                        ignore_space_at_eol=True)
    patch_set = PatchSet(StringIO(uni_diff_text))
    change_list = []  # list of changes 
                      # [(file_name, [row_number_of_deleted_line],
                      # [row_number_of_added_lines]), ... ]
    os.system("make allyesconfig")
    count_1 = 0
    for patched_file in patch_set:
        file_path = patched_file.path  # file name
        print('file name :' + file_path)
        if(file_path.split('.')[-1] != 'c'):
            continue
        cmd = "/home/lsc20011130/smatch/smatch_scripts/kchecker {} >> ../log_smatch/{}_po.log".format(file_path,commit_sha1_po)
        os.system(cmd)
        count_1 = count_1 + 1
    if count_1 > 0:
        cmd = "cat ../log_smatch/{}_po.log | grep \"error:\" > ../logss/{}_{}_error_po.log".format(commit_sha1_po, i, commit_sha1_po)
        os.system(cmd)
        cmd = "cat ../log_smatch/{}_po.log | grep \"warn\" > ../logss/{}_{}_warn_po.log".format(commit_sha1_po, i, commit_sha1_po)
        os.system(cmd)
        # os.system("rm ../{}.log".format(commit_sha1_pr))
    else:
        print("No change to .c file")

    repository.git.reset("--hard",commit_pr)
    os.system("make allyesconfig")
    count_2 = 0
    for patched_file in patch_set:
        file_path = patched_file.path  # file name
        print('file name :' + file_path)
        if(file_path.split('.')[-1] != 'c'):
            continue
        cmd = "/home/lsc20011130/smatch/smatch_scripts/kchecker {} >> ../log_smatch/{}_pr.log".format(file_path,commit_sha1_pr)
        os.system(cmd)
        count_2 = count_2 + 1
    if count_2 > 0:
        cmd = "cat ../log_smatch/{}_pr.log | grep \"error:\" > ../logss/{}_{}_error_pr.log".format(commit_sha1_pr, i, commit_sha1_pr)
        os.system(cmd)
        cmd = "cat ../log_smatch/{}_pr.log | grep \"warn\" > ../logss/{}_{}_warn_pr.log".format(commit_sha1_pr, i, commit_sha1_pr)
        os.system(cmd)
        # os.system("rm ../{}.log".format(commit_sha1_po))
    else:
        print("No change to .c file")

    #analyse the difference of warnings and errors between two commits
    #save warning and error as string format in different lists    
    pre_warn_list = []
    pre_error_list = []
    cur_warn_list = []
    cur_error_list = []
    
    pre_warn_extract_list = []
    pre_error_extract_list = []
    cur_warn_extract_list = []
    cur_error_extract_list = []
    
    if count_2 > 0:
        file = "../logss/{}_{}_warn_pr.log".format(i, commit_sha1_pr)
        with open(file, "r+") as pre_warn:
            line = pre_warn.readline()
            while line is not None and line != "":
                line_extract = extract(line)
                #to avoid smatch shows a same warning/error in the same line many times
                if pre_warn_list.count(line) == 0:
                    pre_warn_list.append(line)
                    pre_warn_extract_list.append(line_extract)
                line = pre_warn.readline()
            
        file = "../logss/{}_{}_error_pr.log".format(i, commit_sha1_pr)
        with open(file, "r+") as pre_error:
            line = pre_error.readline()
            while line is not None and line != "":
                line_extract = extract(line)
                if pre_error_list.count(line) == 0:
                    pre_error_list.append(line)
                    pre_error_extract_list.append(line_extract)
                line = pre_error.readline()
    if count_1 > 0:        
        file = "../logss/{}_{}_warn_po.log".format(i, commit_sha1_po)        
        with open(file, "r+") as cur_warn:
            line = cur_warn.readline()
            while line is not None and line != "":
                line_extract = extract(line)
                if cur_warn_list.count(line) == 0:
                    cur_warn_list.append(line)
                    cur_warn_extract_list.append(line_extract)
                line = cur_warn.readline()
            
        file = "../logss/{}_{}_error_po.log".format(i, commit_sha1_po)        
        with open(file, "r+") as cur_error:
            line = cur_error.readline()
            while line is not None and line != "":
                line_extract = extract(line)
                if cur_error_list.count(line) == 0:
                    cur_error_list.append(line)
                    cur_error_extract_list.append(line_extract)
                line = cur_error.readline()
    
    solved_warn_list = []
    unsolved_warn_list = []
    new_warn_list = []
    
    solved_error_list = []
    unsolved_error_list = []
    new_error_list = []
    
    #analyse the errors and warnings influenced by this commit
    #warnings' situation
    file = "../log_analysis_warning/{}_pr-{}_po.log".format(commit_sha1_pr, commit_sha1_po)
    with open(file, "a+") as f3:
        for j in range(0, len(pre_warn_extract_list)):
            if cur_warn_extract_list.count(pre_warn_extract_list[j]) == 0:
                solved_warn_list.append(pre_warn_list[j])
            else:
                unsolved_warn_list.append(pre_warn_list[j])
                idx = cur_warn_extract_list.index(pre_warn_extract_list[j])
                cur_warn_extract_list.remove(cur_warn_extract_list[idx])
                cur_warn_list.remove(cur_warn_list[idx])
        for j in range (0, len(cur_warn_list)):
            new_warn_list.append(cur_warn_list[j])
    
        f3.write("Solved Warning List:{}".format("\n"))
        for j in range(0, len(solved_warn_list)):
            f3.write("{}:{}".format(j + 1, solved_warn_list[j]))
    
        f3.write("{}Unsolved Warning List:{}".format("\n", "\n"))
        for j in range(0, len(unsolved_warn_list)):
            f3.write("{}:{}".format(j + 1, unsolved_warn_list[j]))
        
        f3.write("{}New Warning List:{}".format("\n", "\n"))
        for j in range(0, len(new_warn_list)):
            f3.write("{}:{}".format(j + 1, new_warn_list[j]))
            
    #errors' situation
    file = "../log_analysis_error/{}_pr-{}_po.log".format(commit_sha1_pr, commit_sha1_po)
    with open(file, "a+") as f3:
        for j in range(0, len(pre_error_extract_list)):
            if cur_error_extract_list.count(pre_error_extract_list[j]) == 0:
                solved_error_list.append(pre_error_list[j])
            else:
                unsolved_error_list.append(pre_error_list[j])
                idx = cur_error_extract_list.index(pre_error_extract_list[j])
                cur_error_extract_list.remove(cur_error_extract_list[idx])
                cur_error_list.remove(cur_error_list[idx])
        for j in range (0, len(cur_error_list)):
            new_error_list.append(cur_error_list[j])
    
        f3.write("Solved Error List:{}".format("\n"))
        for j in range(0, len(solved_error_list)):
            f3.write("{}:{}".format(j + 1, solved_error_list[j]))
    
        f3.write("{}Unsolved Error List:{}".format("\n", "\n"))
        for j in range(0, len(unsolved_error_list)):
            f3.write("{}:{}".format(j + 1, unsolved_error_list[j]))
        
        f3.write("{}New Error List:{}".format("\n", "\n"))
        for j in range(0, len(new_error_list)):
            f3.write("{}:{}".format(j + 1, new_error_list[j]))  
