package main

import (
    "fmt"
    "path"
    "path/filepath"
    "os/exec"
    "os"
)

func main() {
    curr_dir := path.Base("./")
    fmt.Println("dir: ", curr_dir)

    curr_dir = path.Dir("./")
    fmt.Println("dir: ", curr_dir)
    // 获得绝对路径
    path, _ := filepath.Abs("./")
    fmt.Println("dir: ", path)

    file, _ := exec.LookPath(os.Args[0])
    fmt.Println(file)
    // 获得该执行程序的绝对路径
    path, _ = filepath.Abs(file)
    fmt.Println(path)
}