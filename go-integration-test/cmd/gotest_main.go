package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	ctx := context.Background()

	file := flag.String("file", "docker-compose.yml", "Compore configuration files.")
	envFile := flag.String("env-file", "", "Environment file.")

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("nothing to run")
	}

	if err := setupEnv(ctx, *file, *envFile); err != nil {
		log.Fatal(err)
	}

	args := flag.Args()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupEnv(ctx context.Context, file, envFile string) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx,
		"docker", "compose", "-f", file, "config",
		"--format", "json",
	)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	var conf Config
	if err := json.Unmarshal(out, &conf); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx,
		"docker", "compose", "-f", file, "ps",
		"--format", "json",
	)
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	var cs []Container
	if err := json.Unmarshal(out, &cs); err != nil {
		return err
	}

	pat := make(map[string]string)
	for _, c := range cs {
		for _, pub := range c.Publishers {
			from := fmt.Sprintf("%s:%d", c.Service, pub.TargetPort)
			to := fmt.Sprintf("127.0.0.1:%d", pub.PublishedPort)
			pat[from] = to
		}
	}

	envMap, err := readEnvFile(envFile, pat)
	if err != nil {
		return err
	}

	for key, val := range envMap {
		os.Setenv(key, val)
	}

	return nil
}

func readEnvFile(envFile string, pat map[string]string) (map[string]string, error) {
	envMap := make(map[string]string)

	if envFile == "" {
		return envMap, nil
	}

	f, err := os.Open(envFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			continue
		}
		if len(line) >= 1 {
			if bytes.HasPrefix(line, []byte{'/', '/'}) || line[0] == '#' {
				continue
			}
		}

		kv := bytes.SplitN(line, []byte{'='}, 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("malformed file: %s", line)
		}
		key, val := kv[0], string(kv[1])

		for from, to := range pat {
			val = strings.ReplaceAll(val, from, to)
		}

		envMap[string(key)] = val
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return envMap, nil
}

type Config struct {
	Services map[string]Service `json:"services"`
}

type Service struct {
	Image string `json:"image"`
	Ports []Port `json:"port"`
}

type Port struct {
	Mode     string `json:"mode"`
	Target   int    `json:"target"`
	Protocol string `json:"protocol"`
}

type Container struct {
	ID         string
	Name       string
	Project    string
	Service    string
	Publishers []Publisher
}

type Publisher struct {
	URL           string
	TargetPort    int
	PublishedPort int
	Protocol      string
}
