package docker

import (
	"bytes"
	"context"
	"strconv"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/exec"
)

// CurrentContext returns the name of the current Docker context (not to be
// confused with a Go context).
func CurrentContext(ctx context.Context) (string, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()

	args := []string{"info", "--format", "{{ .Host }}"}

	if err := exec.CommandContext(dctx, "docker", args...).Run(); err != nil {
		args := []string{"context", "inspect", "--format", "{{ .Name }}"}
		out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
		if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
			return "", newFastCommandTimeoutError(dctx, args...)
		} else if err != nil {
			return "", err
		}
		name := string(bytes.TrimSpace(out))
		if name == "" {
			return "", errors.New("no context returned from Docker")
		} else {
			return name, nil
		}
	}

	return "", nil
}

// NCPU returns the number of CPU cores available to Docker.
func NCPU(ctx context.Context) (int, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	docker_format := "{{ .NPCU }}"
	podman_format := "{{ .Host.CPUs }}"

	args := []string{"info", "--format", "{{ .Host }}"}
	if err := exec.CommandContext(dctx, "docker", args...).Run(); err != nil {
		args = []string{"info", "--format", docker_format}
	} else {
		args = []string{"info", "--format", podman_format}
	}

	out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
	if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
		return 0, newFastCommandTimeoutError(dctx, args...)
	} else if err != nil {
		return 0, err
	}

	dcpu, err := strconv.Atoi(string(bytes.TrimSpace(out)))
	if err != nil {
		return 0, errors.Wrap(err, "parsing docker cpu count")
	}

	return dcpu, nil
}
