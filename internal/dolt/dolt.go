package dolt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Dolt struct {
	Dir string
}

func New(dir string) *Dolt {
	return &Dolt{Dir: dir}
}

func (d *Dolt) IsInitialized() bool {
	info, err := os.Stat(filepath.Join(d.Dir, ".dolt"))
	return err == nil && info.IsDir()
}

func (d *Dolt) Init() error {
	if d.IsInitialized() {
		return nil
	}
	if err := os.MkdirAll(d.Dir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}
	return d.run("init")
}

// SQL executes a SQL query and returns raw output.
func (d *Dolt) SQL(query string) ([]byte, error) {
	return d.output("sql", "-q", query, "-r", "json")
}

// SQLExec executes a SQL statement (no output expected).
func (d *Dolt) SQLExec(query string) error {
	return d.run("sql", "-q", query)
}

// SQLBatch executes multiple SQL statements from a string.
func (d *Dolt) SQLBatch(statements string) error {
	cmd := exec.Command("dolt", "sql")
	cmd.Dir = d.Dir
	cmd.Stdin = strings.NewReader(statements)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dolt sql batch: %s: %w", stderr.String(), err)
	}
	return nil
}

// QueryRows executes a query and returns the rows as a slice of maps.
func (d *Dolt) QueryRows(query string) ([]map[string]interface{}, error) {
	out, err := d.SQL(query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []map[string]interface{} `json:"rows"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing dolt json output: %w", err)
	}
	return result.Rows, nil
}

func (d *Dolt) Add() error {
	return d.run("add", ".")
}

func (d *Dolt) Commit(message string) error {
	err := d.run("commit", "-m", message, "--allow-empty")
	if err != nil && strings.Contains(err.Error(), "nothing to commit") {
		return nil
	}
	return err
}

func (d *Dolt) run(args ...string) error {
	cmd := exec.Command("dolt", args...)
	cmd.Dir = d.Dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dolt %s: %s: %w", args[0], stderr.String(), err)
	}
	return nil
}

func (d *Dolt) output(args ...string) ([]byte, error) {
	cmd := exec.Command("dolt", args...)
	cmd.Dir = d.Dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dolt %s: %s: %w", args[0], stderr.String(), err)
	}
	return out, nil
}
