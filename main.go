package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type any = interface{}

var (
	upoladDir string

	// server address
	address string
)

func main() {
	var auto bool
	var host, port, dir string

	flag.StringVar(&host, "h", "127.0.0.1", "server host.")
	flag.StringVar(&port, "p", "8080", "server port.")
	flag.StringVar(&dir, "dir", "./public", "upload file dir.")
	flag.BoolVar(&auto, "auto", true, "auto get wifi ip.")
	flag.Parse()

	// check upload dir
	dir, err := filepath.Abs(dir)
	if err != nil {
		if err = os.Mkdir(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	upoladDir = dir

	// get ip address
	if auto {
		a := getIP("192")
		if a != "" {
			host = a
		}
	}
	addr := net.JoinHostPort(host, port)

	// init router
	mux := InitRouter()
	Debug("Serve:%s, Dir:\"%s\"", addr, upoladDir)
	http.ListenAndServe(addr, mux)
}

func Debug(str string, arg ...any) {
	// format:
	// [01-02 15:04] msg ...

	format := "[%s]%s\n"
	t := time.Now().Format("01-02 15:04")
	msg := fmt.Sprintf(str, arg...)
	fmt.Printf(format, t, msg)
}

func getIP(prefix string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	var addr string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				a := ipnet.IP.To4().String()
				if strings.HasPrefix(a, prefix) {
					addr = a
					break
				}
			}
		}
	}
	return addr
}
