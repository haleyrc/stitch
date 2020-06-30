package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Logger struct {
	Writer io.Writer
	Prefix string
}

func (l Logger) Write(p []byte) (int, error) {
	fmt.Fprintf(l.Writer, "%s ", time.Now().Format(time.RFC3339))
	fmt.Fprintf(l.Writer, "%s ", l.Prefix)
	return l.Writer.Write(p)
}

func main() {
	ctx := context.Background()

	cfg, err := parseConfig(ctx)
	if err != nil {
		panic(err)
	}

	cmds, err := buildCommands(ctx, cfg)
	if err != nil {
		panic(err)
	}

	if err := runAll(cmds...); err != nil {
		panic(err)
	}
}

func buildCommands(ctx context.Context, cfg *Config) ([]*exec.Cmd, error) {
	cmds := []*exec.Cmd{}

	for _, svc := range cfg.Services {
		localEnv, err := parseEnv(ctx, cfg.Environment, svc.Environment)
		if err != nil {
			return nil, err
		}

		logger := Logger{
			Writer: os.Stderr,
			Prefix: "[" + svc.Name + "]",
		}

		parts := strings.Split(svc.Command, " ")
		if len(parts) < 1 {
			return nil, errors.New("no command specified for service")
		}

		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Env = localEnv
		cmd.Dir = svc.WorkDir
		cmd.Stderr = logger
		cmd.Stdout = logger

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func parseEnv(ctx context.Context, envMaps ...map[string]string) ([]string, error) {
	// TODO (RCH): Can we do this better?
	env := envFromSystem()
	for _, envMap := range envMaps {
		for k, v := range envMap {
			env[k] = v
		}
	}
	return toList(env), nil
}

func toList(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func envFromSystem() map[string]string {
	all := os.Environ()
	env := map[string]string{}
	for _, one := range all {
		parts := strings.Split(one, "=")
		env[parts[0]] = parts[1]
	}
	return env
}

type Service struct {
	Name        string
	WorkDir     string
	Command     string
	Environment map[string]string
}

type Config struct {
	Services    []Service
	Environment map[string]string
}

func parseConfig(ctx context.Context) (*Config, error) {
	fileBytes, err := ioutil.ReadFile("example/config.json")
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func runAll(cmds ...*exec.Cmd) error {
	errC := make(chan int, 1)

	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			for j := 0; j < i; j++ {
				if err := cmds[j].Process.Kill(); err != nil {
					fmt.Printf("failed to stop %s: %v\n", cmds[j].Path, err)
				}
			}
			return err
		}
		go func(idx int, cmd *exec.Cmd) {
			cmd.Wait()
			errC <- idx
		}(i, cmd)
	}

	// Wait for first error
	which := <-errC

	// Stop all commands
	for i := range cmds {
		// Don't try to stop a process that already exited in error
		if i == which {
			continue
		}
		if err := cmds[i].Process.Kill(); err != nil {
			fmt.Printf("failed to stop %s: %v\n", cmds[i].Path, err)
		}
	}

	// Return the error
	// TODO (RCH): We need to actually capture the error somehow
	return nil
}
