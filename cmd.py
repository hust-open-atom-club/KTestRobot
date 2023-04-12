import sys
import os
import subprocess
from io import StringIO

def re(argv):
    cmd = "git apply -R {}".format(argv[0])
    os.system(cmd)
    
def main(argv):
    #pre smatch
    cmd = "/home/lsc20011130/smatch/smatch_scripts/kchecker {} >> ../log_smatch/pre_smatch.log".format(argv[1])
    os.system(cmd)
    
    cmd = "cat ../log_smatch/pre_smatch.log | grep \"error:\" > ../logss/pre_smatch_error.log"
    os.system(cmd)
    
    cmd = "cat ../log_smatch/pre_smatch.log | grep \"warn\" > ../logss/pre_smatch_warn.log"
    os.system(cmd)
    
    #check patch
    cmd = "git apply --check {}".format(argv[0])
    result = subprocess.getoutput(cmd)
    result = result + "\n"
    if result != "\n":
        sys.stderr.write(result)
        raise SystemExit(1) 
    
    #apply patch
    cmd = "git apply {}".format(argv[0])
    os.system(cmd)
    
    #smatch again
    cmd = "/home/lsc20011130/smatch/smatch_scripts/kchecker {} >> ../log_smatch/after_smatch.log".format(argv[1])
    os.system(cmd)
    
    cmd = "cat ../log_smatch/after_smatch.log | grep \"error:\" > ../logss/after_smatch_error.log"
    os.system(cmd)
    
    cmd = "cat ../log_smatch/after_smatch.log | grep \"warn\" > ../logss/after_smatch_warn.log"
    os.system(cmd)
    
    #depatch
    cmd = "git apply -R {}".format(argv[0])
    os.system(cmd)
    
main(sys.argv[1:])

