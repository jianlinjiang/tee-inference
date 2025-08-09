package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

const AizelModelPrefix = "aizel/"

var Backend *ModelBackend

type ModelBackend struct {
	mu           sync.Mutex
	currentCmd   *exec.Cmd
	currentModel string
	StartupDone  context.Context
	cancelFunc   context.CancelFunc
}

func NewModelBackend() *ModelBackend {
	return &ModelBackend{}
}

func (m *ModelBackend) RunModel(ctx context.Context, modelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if model not equal to current model
	if m.currentCmd != nil && m.currentModel == modelName {
		return nil
	}

	if m.currentCmd != nil {
		if err := m.currentCmd.Process.Kill(); err != nil {
			return errors.New("failed to stop current model process: " + err.Error())
		}
		if err := m.currentCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == -1 {
				} else {
					return errors.New("failed to wait for process termination: " + err.Error())
				}
			} else {
				return errors.New("failed to wait for process termination: " + err.Error())
			}
		}
		m.currentCmd = nil
	}

	startupCtx, cancel := context.WithCancel(ctx)
	m.StartupDone = startupCtx
	m.cancelFunc = cancel

	modelPath := fmt.Sprintf("%s/%s", "/app/models", modelName)

	// launch a new model process
	cmdArgs := []string{
		"-c", "4096", "-b", "4096", "-t", "12",
		"-m", modelPath, "--host", "0.0.0.0", "--port", "80", "--jinja",
	}

	cmd := exec.Command("/app/llama-server", cmdArgs...)

	if err := cmd.Start(); err != nil {
		return errors.New("failed to start model backend process: " + err.Error())
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := http.Get(fmt.Sprintf("http://%s/health", "0.0.0.0"))
				if err == nil && resp.StatusCode == http.StatusOK {
					m.cancelFunc()
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	m.currentCmd = cmd
	m.currentModel = modelName

	return nil
}

func (m *ModelBackend) StopModel() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentCmd == nil {
		return errors.New("no model process is running")
	}

	if err := m.currentCmd.Process.Kill(); err != nil {
		return errors.New("failed to stop model process: " + err.Error())
	}
	m.currentCmd = nil
	return nil
}
