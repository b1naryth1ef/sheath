package deno

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type BaseOpts struct {
	Location  string
	Reload    []string
	ImportMap string
}

func (b BaseOpts) Args() []string {
	args := []string{}
	if b.Location != "" {
		args = append(args, fmt.Sprintf("--location=%s", b.Location))
	}
	if b.Reload != nil {
		args = parg(args, "reload", b.Reload)
	}
	if b.ImportMap != "" {
		args = append(args, "--import-map="+b.ImportMap)
	}
	return args
}

type PermissionOpts struct {
	All    bool
	Prompt bool

	Read   []string
	Write  []string
	Import []string
	Net    []string
	Env    []string
	Sys    []string
	Run    []string
	FFI    []string
}

func parg(v []string, name string, values []string) []string {
	if values == nil {
		return v
	}

	value := "--" + name
	if len(values) > 0 {
		value = fmt.Sprintf("%s=%s", value, strings.Join(values, ","))
	}
	return append(v, value)
}

func (p PermissionOpts) Args() []string {
	args := []string{}

	if !p.Prompt {
		args = append(args, "--no-prompt")
	}

	if p.All {
		return append(args, "-A")
	}

	if p.Read != nil {
		args = parg(args, "allow-read", p.Read)
	}
	if p.Write != nil {
		args = parg(args, "allow-write", p.Write)
	}
	if p.Import != nil {
		args = parg(args, "allow-import", p.Import)
	}
	if p.Net != nil {
		args = parg(args, "allow-net", p.Net)
	}
	if p.Env != nil {
		args = parg(args, "allow-import", p.Import)
	}
	if p.Sys != nil {
		args = parg(args, "allow-sys", p.Sys)
	}
	if p.Run != nil {
		args = parg(args, "allow-run", p.Run)
	}
	if p.FFI != nil {
		args = parg(args, "allow-ffi", p.FFI)
	}
	return args
}

type DebuggingOpts struct {
	Inspect     *string
	InspectBrk  *string
	InspectWait *string
}

func (d DebuggingOpts) Args() []string {
	args := []string{}

	if d.Inspect != nil {
		args = append(args, "--inspect="+*d.Inspect)
	} else if d.InspectBrk != nil {
		args = append(args, "--inspect-brk="+*d.InspectBrk)
	} else if d.InspectWait != nil {
		args = append(args, "--inspect-wait="+*d.InspectWait)
	}

	return args
}

type Run struct {
	BaseOpts
	PermissionOpts
	DebuggingOpts

	Target string
	Args   []string
	Env    []string

	Path string

	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func (r *Run) command() *exec.Cmd {
	args := []string{
		"run",
		"--no-prompt",
		"--no-lock",
	}
	args = append(args, r.BaseOpts.Args()...)
	args = append(args, r.PermissionOpts.Args()...)
	args = append(args, r.DebuggingOpts.Args()...)
	args = append(args, r.Target)

	if r.Args != nil {
		args = append(args, r.Args...)
	}

	cmd := exec.Command(
		"deno",
		args...,
	)

	cmd.Env = append([]string{}, os.Environ()...)
	if r.Env != nil {
		cmd.Env = append(cmd.Env, r.Env...)
	}
	cmd.Dir = r.Path
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	cmd.Stdin = r.Stdin
	return cmd
}

func (r *Run) Launch() (*RunInstance, error) {
	cmd := r.command()
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return &RunInstance{
		Run: r,
		Cmd: cmd,
	}, nil
}

func (r *Run) Run() error {
	cmd := r.command()
	return cmd.Run()
}

type RunInstance struct {
	Run *Run
	Cmd *exec.Cmd
}
