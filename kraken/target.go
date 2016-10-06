package kraken

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sean-jc/pipes"
	yaml "gopkg.in/yaml.v2"
)

// Script defines a post-processing script for extracting subroutine latencies
// from the output of the target application.
type Script struct {
	Name   string `yaml:"name"`
	Script string `yaml:"script"`
}

type targetMetadata struct {
	// Path is the path to the executable to run
	Path string `yaml:"path"`

	// Args contains
	Args string `yaml:"args"`

	// Post processing map
	Post []Script `yaml:"post"`
}

type Target struct {
	path   string
	args   []string
	cgroup string
	post   []Script
}

func NewTarget(mpath string, args string, cgroup string) (*Target, error) {
	raw, err := ioutil.ReadFile(mpath)
	if err != nil {
		return nil, fmt.Errorf("cannot read target file: %v", err)
	}

	var m targetMetadata
	err = yaml.Unmarshal(raw, &m)
	if err != nil {
		return nil, fmt.Errorf("cannot parse target file: %v", err)
	}

	t := &Target{
		path:   m.Path,
		args:   append(strings.Fields(m.Args), strings.Fields(args)...),
		cgroup: cgroup,
		post:   m.Post,
	}
	return t, nil
}

func (t *Target) Hit(w io.Writer) error {
	if t.cgroup != "" {
		args := append([]string{"-g", t.cgroup, t.path}, t.args...)
		return pipes.ExecStdout(exec.Command("cgexec", args...), w)
	}
	return pipes.ExecStdout(exec.Command(t.path, t.args...), w)
}

func runScript(script string, stdin io.Reader) ([]byte, error) {
	ops := strings.Split(script, "|")
	cmds := make([]*exec.Cmd, len(ops))

	rex := regexp.MustCompile("'.+'|\".+\"|\\S+")
	for i, op := range ops {
		f := rex.FindAllString(op, -1)
		for j := range f {
			f[j] = strings.Trim(f[j], "'")
		}
		cmds[i] = exec.Command(f[0], f[1:]...)
	}
	return pipes.ExecPipelineO(cmds, stdin)
}

func (t *Target) ProcessOutput(stdin io.Reader, res *Result) error {
	for _, s := range t.post {
		var in bytes.Buffer

		// Recapture the incoming data into a buffer and point stdin to the new
		// buffer after running the script.  This allows processing the output
		// multiple times (reading from stdin is destructive).
		out, err := runScript(s.Script, io.TeeReader(stdin, &in))
		if err != nil {
			return err
		}
		stdin = &in

		l, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err != nil {
			return err
		}

		res.Latencies = append(res.Latencies, Latency{s.Name, time.Duration(l * float64(time.Second))})
	}
	return nil
}
