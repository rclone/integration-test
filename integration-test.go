package main

import (
       "fmt"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

var (
	// Flags
	debug        = flag.Bool("d", false, "Print commands instead of running them.")
	verbose      = flag.Bool("v", false, "Print commands as they are being executed and command output.")
	gopath       = flag.String("gopath", os.ExpandEnv("${HOME}/go"), "GOPATH to use.")
	prepareOnly  = flag.Bool("prepare-only", false, "Do everything except run tests")
	pr           = flag.String("pr", "", "PR number to test")
	branch       = flag.String("branch", "", "branch to test (defaults to master)")
	backends     = flag.String("backends", "", "to pass to test_all -backends")
	remotes      = flag.String("remotes", "", "to pass to test_all -remotes")
	tests        = flag.String("tests", "", "to pass to test_all -tests")
	runRegexp    = flag.String("run", "", "to pass to test_all -run")
	outputDir    = flag.String("output", "/home/rclone/integration-test/rclone-integration-tests", "write test output here")
	outputDirMax = flag.Int("output-max", 60, "maximum number of directories in outputDir")
	maxTries     = flag.Int("maxtries", -1, "if set, overrides the -maxtries for test_all")
	// Globals
	gobin         string // place for go binaries - filled in by main()
	rcloneVersion string // version of rclone
	rcloneCommit  string // commit of rclone
)

// cmdEnv - create a shell command with env
func cmdEnv(args, env []string) *exec.Cmd {
	if *debug {
		args = append([]string{"echo"}, args...)
	}
	cmd := exec.Command(args[0], args[1:]...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	if *debug || *verbose {
		log.Printf("args = %v, env = %v\n", args, cmd.Env)
	}
	log.Println(strings.Join(args, " "))
	return cmd
}

// runEnv - run a shell command with env
//
// The outputs the command output to stderr and returns it
func runEnv(args, env []string) (string, error) {
	cmd := cmdEnv(args, env)
	var b bytes.Buffer
	cmd.Stderr = io.MultiWriter(&b, os.Stderr)
	cmd.Stdout = cmd.Stderr
	err := cmd.Run()
	out := b.Bytes()
	outStr := string(out)
	if *verbose {
		log.Print(string(out))
	}
	if err != nil {
		log.Print("----------------------------")
		log.Printf("Failed to run %v: %v", args, err)
		log.Print("----------------------------")
	}
	return outStr, err
}

// run a shell command
func run(args ...string) string {
	out, err := runEnv(args, nil)
	if err != nil {
		log.Fatalf("Exiting after error: %v", err)
	}
	return out
}

// xrun runs a shell command ignoring errors
func xrun(args ...string) string {
	out, err := runEnv(args, nil)
	if err != nil {
		log.Printf("Ignoring error: %v", err)
	}
	return out
}

// chdir or die
func chdir(dir string) {
	log.Printf("cd %s", dir)
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("Couldn't cd into %q: %v", dir, err)
	}
}

// set env or die
func setenv(key, value string) {
	log.Printf("export %s=%s", key, value)
	err := os.Setenv(key, value)
	if err != nil {
		log.Fatalf("Couldn't Setenv(%q, %q): %v", key, value, err)
	}
}

// make all the directories or die
func mkdirall(path string) {
	log.Printf("mkdir -p %s", path)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Fatalf("Couldn't MkdirAll(%q): %v", path, err)
	}
}

// check the path or directory exists
func exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		log.Fatalf("exists(%q): %v", path, err)
	}
	return false
}

// install a github repo into the GOPATH
func installGitHubRepo(repo string) {
	gitPath := path.Join(*gopath, "src/github.com/"+repo)

	// make path and cd
	mkdirall(gitPath)
	chdir(gitPath)

	// checkout the code
	if exists(".git") {
		run("git", "stash", "--include-untracked") // stash any local changes just in case
		run("git", "checkout", "master")
		run("git", "pull")
	} else {
		run("git", "clone", "https://github.com/"+repo+".git", ".")
	}
}

// install rclone and checkout correct branch
func installRclone() {
	installGitHubRepo("rclone/rclone")

	// tidy up from previous runs
	run("rm", "-f", "fs/operations/operations.test", "fs/sync/sync.test", "fs/test_all.log", "summary", "test.log")

	branchName := "master"
	pullName := ""
	if *pr != "" {
		branchName = "pr-" + *pr
		pullName = "pull/" + *pr + "/head"
	} else if *branch != "" {
		branchName = *branch
		pullName = branchName
	}
	if pullName != "" {
		xrun("git", "branch", "-D", branchName)
		run("git", "fetch", "origin", pullName+":"+branchName)
		run("git", "checkout", branchName)
	}

	// build rclone
	run("make")
}

func installRestic() {
	// make sure restic is up to date for the cmd/serve/restic integration tests
	run("go", "get", "-u", "github.com/restic/restic/...")
}

// run the rclone integration tests against all the remotes
func runTests(rclonePath string) {
	chdir(rclonePath)
	run("make", "test_all")
	if !*prepareOnly {
		args := []string{
			path.Join(gobin, "test_all"),
			"-verbose",
			"-upload", "memstore:pub-rclone-org/integration-tests",
			"-email", "nick@craig-wood.com",
			"-output", *outputDir,
		}
		if *backends != "" {
			args = append(args, "-backends", *backends)
		}
		if *remotes != "" {
			args = append(args, "-remotes", *remotes)
		}
		if *tests != "" {
			args = append(args, "-tests", *tests)
		}
		if *runRegexp != "" {
			args = append(args, "-run", *runRegexp)
		}
		if *maxTries > 0 {
			args = append(args, "-maxtries", fmt.Sprint(*maxTries))
		}
		xrun(args...)
	}
}

// make sure there aren't too many items in the output dir
func tidyOutputDir() {
	fis, err := os.ReadDir(*outputDir)
	if err != nil {
		log.Fatalf("Failed to read output directory %q: %v", outputDir, err)
	}
	var names []string
	for _, fi := range fis {
		if fi.IsDir() {
			names = append(names, fi.Name())
		}
	}
	sort.Strings(names)
	if trim := len(names) - *outputDirMax; trim > 0 {
		log.Printf("Need to trim %d directories", trim)
		for _, dir := range names[:trim] {
			dir = path.Join(*outputDir, dir)
			log.Printf("Trimming %s", dir)
			err := os.RemoveAll(dir)
			if err != nil {
				log.Printf("Failed to remove %q: %v", dir, err)
			}
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		log.Fatalf("Syntax: %s [opts]", os.Args[0])
	}

	tidyOutputDir()

	gobin = path.Join(*gopath, "bin")
	setenv("GOPATH", *gopath)
	setenv("GOTAGS", "cmount") // make sure we build the optional extras

	installGitHubRepo("restic/restic") // install restic source so we can use its tests
	installRclone()

	rclonePath := path.Join(*gopath, "src/github.com/rclone/rclone")
	runTests(rclonePath)
}
