def getCols():
    import subprocess
    return int(subprocess.Popen(\
            ['tput','cols'], stdout=subprocess.PIPE, stderr=subprocess.PIPE\
        ).communicate()[0])
print(getCols())