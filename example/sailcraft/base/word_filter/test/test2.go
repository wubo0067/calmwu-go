package main

import (
	"filter"
	_ "fmt"
	"net/http"
	"runtime"
)

var dicFile1 string
var dicFiles []string

func filtermsg(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	msgarr := r.Form["msg"]
	separr := r.Form["sep"]
	reparr := r.Form["rep"]

	var msg []rune = nil
	var sep []rune = nil
	var rep rune = 0

	if msgarr != nil {
		msg = []rune(msgarr[0])
	}
	if separr != nil {
		sep = []rune(separr[0])
	}
	if reparr != nil {
		tmp := []rune(reparr[0])
		if len(tmp) > 0 {
			rep = tmp[0]
		}
	}
	filter.FilterText(dicFile1, msg, sep, rep)
	w.Write(append([]byte(string(msg)), byte('\n')))
}

func main() {

	dicFile1 = "dic.txt"
	dicFiles = append(dicFiles, dicFile1)

	filter.LoadDicFiles(dicFiles)

	runtime.GOMAXPROCS(2)

	http.HandleFunc("/filtermsg", filtermsg)
	http.ListenAndServe(":9090", nil)
}
