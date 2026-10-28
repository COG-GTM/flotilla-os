package services

import (
	"fmt"
	"strings"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

// serviceAccountPolicy decides which Kubernetes service accounts (and, for
// eks-spark, which emr_job_role_arn entries) a launch request may select.
// Callers may only name the engine default, an account in
// service_account_allowlist, or an account mapped to the request's team label
// in service_account_allowlist_by_team. Everything else is rejected.
type serviceAccountPolicy struct {
	eksDefault string
	emrDefault string
	global     map[string]bool
	byTeam     map[string]map[string]bool
	emrRoleArn map[string]string
}

func newServiceAccountPolicy(conf config.Config) serviceAccountPolicy {
	p := serviceAccountPolicy{
		eksDefault: strings.TrimSpace(conf.GetString("eks_default_service_account")),
		emrDefault: strings.TrimSpace(conf.GetString("emr_default_service_account")),
		global:     toSet(conf.GetStringSlice("service_account_allowlist")),
		byTeam:     map[string]map[string]bool{},
		emrRoleArn: conf.GetStringMapString("emr_job_role_arn"),
	}
	// Values are comma-separated so the map can also be supplied via env/flat config.
	for team, accounts := range conf.GetStringMapString("service_account_allowlist_by_team") {
		p.byTeam[strings.ToLower(strings.TrimSpace(team))] = toSet(strings.Split(accounts, ","))
	}
	return p
}

func toSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			set[v] = true
		}
	}
	return set
}

func (p serviceAccountPolicy) allowed(sa string, labels *state.Labels) bool {
	if p.global[sa] {
		return true
	}
	if labels == nil {
		return false
	}
	team := strings.ToLower(strings.TrimSpace((*labels)["team"]))
	return team != "" && p.byTeam[team][sa]
}

// authorize validates fields.ServiceAccount in place. A blank value is treated
// as absent so the engine default applies.
func (p serviceAccountPolicy) authorize(fields *state.ExecutionRequestCommon) error {
	if fields.ServiceAccount != nil && strings.TrimSpace(*fields.ServiceAccount) == "" {
		fields.ServiceAccount = nil
	}

	isSpark := fields.Engine != nil && *fields.Engine == state.EKSSparkEngine
	defaultSA := p.eksDefault
	if isSpark {
		defaultSA = p.emrDefault
	}

	if fields.ServiceAccount != nil {
		sa := *fields.ServiceAccount
		if sa != defaultSA && !p.allowed(sa, fields.Labels) {
			return exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("service_account [%s] is not permitted for this request", sa),
			}
		}
	}

	if isSpark {
		sa := defaultSA
		if fields.ServiceAccount != nil {
			sa = *fields.ServiceAccount
		}
		if p.emrRoleArn[sa] == "" {
			return exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("service_account [%s] has no EMR execution role configured", sa),
			}
		}
	}
	return nil
}
