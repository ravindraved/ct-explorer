package fetchers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// FetchControls fetches enabled Control Tower controls for every OU and account
// in the store and populates the store with Control entries.
func FetchControls(ctx context.Context, cfg aws.Config, s *store.Store) error {
	client := controltower.NewFromConfig(cfg)

	ous := s.AllOUs()
	if len(ous) == 0 {
		slog.Warn("no_ous_found", "component", "fetcher", "message", "No OUs in store; skipping control fetch")
		return nil
	}

	total := 0
	for _, ou := range ous {
		count := fetchControlsForTarget(ctx, client, s, ou.ARN, ou.ID)
		total += count
	}

	accounts := s.AllAccounts()
	for _, acct := range accounts {
		count := fetchControlsForTarget(ctx, client, s, acct.ARN, acct.ID)
		total += count
	}

	slog.Info("fetch_complete", "component", "fetcher", "control_count", total)
	return nil
}

// fetchControlsForTarget fetches enabled controls for a single target (OU or account).
// Returns the count of controls fetched.
func fetchControlsForTarget(
	ctx context.Context,
	client *controltower.Client,
	s *store.Store,
	targetARN string,
	targetID string,
) int {
	count := 0

	paginator := controltower.NewListEnabledControlsPaginator(client,
		&controltower.ListEnabledControlsInput{
			TargetIdentifier: aws.String(targetARN),
		},
	)

	cliCmd := fmt.Sprintf("aws controltower list-enabled-controls --target-identifier %s", targetARN)

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordControlError(s, "controltower:list_enabled_controls", err)
			break
		}

		for _, entry := range page.EnabledControls {
			ctrl := mapControl(entry, targetID)
			s.PutControl(ctrl)
			count++
		}
	}

	return count
}

// mapControl maps an EnabledControlSummary to a models.Control.
func mapControl(entry types.EnabledControlSummary, targetID string) models.Control {
	controlARN := aws.ToString(entry.ControlIdentifier)
	controlID := controlARN
	if idx := strings.LastIndex(controlARN, "/"); idx >= 0 {
		controlID = controlARN[idx+1:]
	}

	controlType := inferControlType(controlID)

	enforcement := models.EnforcementEnabled
	if entry.StatusSummary != nil && entry.StatusSummary.Status != types.EnablementStatusSucceeded {
		enforcement = models.EnforcementFailed
	}

	return models.Control{
		ARN:         controlARN,
		ControlID:   controlID,
		Name:        controlID,
		ControlType: controlType,
		Enforcement: enforcement,
		TargetID:    targetID,
	}
}

// inferControlType does a best-effort control type inference from the control ID.
func inferControlType(controlID string) models.ControlType {
	upper := strings.ToUpper(controlID)
	if strings.Contains(upper, "CT_") {
		return models.ControlTypeDetective
	}
	if strings.Contains(upper, "GR_") {
		return models.ControlTypePreventive
	}
	return models.ControlTypePreventive
}

// recordControlError adds an ErrorInfo to the store for a Control Tower API failure.
func recordControlError(s *store.Store, apiCall string, err error) {
	s.AddError(models.ErrorInfo{
		Source:      fmt.Sprintf("Control Tower API (%s)", apiCall),
		Message:     err.Error(),
		Recoverable: false,
	})
}
