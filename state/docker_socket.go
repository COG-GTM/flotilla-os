package state

import (
	"strings"

	"github.com/stitchfix/flotilla-os/config"
)

// DockerSocketAllowlistKey names the config list of definition ids whose runs
// may mount the node's Docker socket. Empty (the default) disables the mount.
const DockerSocketAllowlistKey = "eks_requires_docker_allowed_definition_ids"

// DockerSocketAllowlist is the server-side set of definition ids allowed to
// use requires_docker. Mounting /var/run/docker.sock gives the job root on the
// node, so the flag on a definition is never trusted by itself.
type DockerSocketAllowlist map[string]struct{}

// NewDockerSocketAllowlist reads DockerSocketAllowlistKey from conf. It accepts
// a YAML list or a comma-separated string (e.g. from an environment variable).
func NewDockerSocketAllowlist(conf config.Config) DockerSocketAllowlist {
	allowlist := DockerSocketAllowlist{}
	if conf == nil {
		return allowlist
	}
	for _, entry := range conf.GetStringSlice(DockerSocketAllowlistKey) {
		for _, id := range strings.Split(entry, ",") {
			if id = strings.TrimSpace(id); len(id) > 0 {
				allowlist[id] = struct{}{}
			}
		}
	}
	return allowlist
}

// Permits reports whether runs of definitionID may mount the Docker socket.
func (a DockerSocketAllowlist) Permits(definitionID string) bool {
	if len(definitionID) == 0 {
		return false
	}
	_, ok := a[definitionID]
	return ok
}
