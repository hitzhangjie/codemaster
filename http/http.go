package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/phayes/freeport"
)

type Response struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	Port    int    `json:"port"`
}

func process(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	resp := &Response{}

	defer func() {
		buf, err := json.Marshal(resp)
		if err != nil {
			log.Printf("resp marshal error: %v", err)
			return
		}
		fmt.Fprintf(w, string(buf))
	}()

	if err := req.ParseForm(); err != nil {
		resp.ErrCode = 1
		resp.ErrMsg = err.Error()
		return
	}

	log.Printf("--------------------------------")
	log.Printf("start Web IDE")

	repo := req.FormValue("repo")
	if len(repo) == 0 {
		log.Printf("check param error: missing `repo`")
		return
	}
	log.Printf("start Web IDE, repo: %s", repo)

	port, err := freeport.GetFreePort()
	if err != nil {
		resp.ErrCode = 2
		resp.ErrMsg = err.Error()
		log.Printf("allocate port error: %v", err)
		return
	}
	log.Printf("start Web IDE, port %d", port)

	resp.ErrCode = 0
	resp.ErrMsg = "success"
	resp.Port = port

	go func() {
		if err := startWebIDE(repo, port); err != nil {
			log.Printf("start Web IDE, port: %v, repo: %s error: %v", port, repo, err)
			return
		}
		log.Printf("start Web IDE, port: %v, repo: %s, success", port, repo)
	}()

	return
}

func main() {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", process)
	if err := http.Serve(ln, nil); err != nil {
		panic(err)
	}
}

func startWebIDE(repo string, port int) error {

	idx := strings.LastIndex(repo, "/")
	dir := repo[idx+1:]
	//os.TempDir() is not shared in Docker settings, use /tmp instead
	//target := filepath.Join(os.TempDir(), dir)
	target := filepath.Join("/tmp", dir)

	_ = os.RemoveAll(target)

	// clone
	clone := exec.Command("git", "clone", repo, target)
	if buf, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("clone error: %v, details: %s", err, string(buf))
	}
	log.Printf("clone repo success")

	// docker
	var docker *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		docker = exec.Command("docker", "run",
			"--name", fmt.Sprintf("panda-%s", dir),
			"-p", fmt.Sprintf("%d:3000", port),
			"-v", fmt.Sprintf("%s:/home/project", target),
			"--rm",
			"hitzhangjie/panda")
	case "darwin":
		docker = exec.Command("docker", "run",
			"--name", fmt.Sprintf("panda-%s", dir),
			"-p", fmt.Sprintf("%d:3000", port),
			"-v", fmt.Sprintf("%s:/home/project:cached", target),
			"--rm",
			"hitzhangjie/panda")
	default:
		return errors.New("not supported OS")
	}

	if buf, err := docker.CombinedOutput(); err != nil {
		return fmt.Errorf("docker error: %v, details: %s", err, string(buf))
	}
	log.Printf("docker run success")
	return nil
}
