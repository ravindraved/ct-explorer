package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// FetchSCPs fetches all Service Control Policies, their policy documents,
// and attachment targets, then stores them. Mirrors the Python scp_fetcher.
func FetchSCPs(ctx context.Context, cfg aws.Config, s *store.Store) error {
	client := organizations.NewFromConfig(cfg)

	policies := listAllSCPs(ctx, client, s)
	if len(policies) == 0 {
		slog.Info("fetch_complete", "component", "fetcher", "scp_count", 0)
		return nil
	}

	total := 0
	for _, summary := range policies {
		scp := buildSCP(ctx, client, s, summary)
		if scp != nil {
			s.PutSCP(*scp)
			total++
		}
	}

	slog.Info("fetch_complete", "component", "fetcher", "scp_count", total)
	return nil
}

// listAllSCPs lists all SCP policy summaries with pagination.
func listAllSCPs(ctx context.Context, client *organizations.Client, s *store.Store) []types.PolicySummary {
	results := make([]types.PolicySummary, 0)

	paginator := organizations.NewListPoliciesPaginator(client,
		&organizations.ListPoliciesInput{
			Filter: types.PolicyTypeServiceControlPolicy,
		},
	)

	cliCmd := "aws organizations list-policies --filter SERVICE_CONTROL_POLICY"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordSCPError(s, "organizations:list_policies", err)
			break
		}

		results = append(results, page.Policies...)
	}

	return results
}

// buildSCP describes a single SCP, fetches its document and targets.
func buildSCP(ctx context.Context, client *organizations.Client, s *store.Store, summary types.PolicySummary) *models.SCP {
	policyID := aws.ToString(summary.Id)

	// Fetch full policy detail
	cliCmd := fmt.Sprintf("aws organizations describe-policy --policy-id %s", policyID)
	LogAPICall(cliCmd, "ok")

	docRaw := ""
	summaryData := summary

	resp, err := client.DescribePolicy(ctx, &organizations.DescribePolicyInput{
		PolicyId: aws.String(policyID),
	})
	if err != nil {
		LogAPICallError(cliCmd, err.Error())
		recordSCPError(s, "organizations:describe_policy", err)
	} else if resp.Policy != nil {
		docRaw = aws.ToString(resp.Policy.Content)
		if resp.Policy.PolicySummary != nil {
			summaryData = *resp.Policy.PolicySummary
		}
	}

	// Parse policy document JSON
	var parsed *models.Policy
	if docRaw != "" {
		p, parseErr := parsePolicyDocument(docRaw)
		if parseErr != nil {
			slog.Warn("policy_parse_failed", "component", "fetcher", "scp_id", policyID, "error", parseErr)
		} else {
			parsed = p
		}
	}

	// Fetch targets
	targetIDs := listTargets(ctx, client, s, policyID)

	scp := models.NewSCP()
	scp.ID = policyID
	scp.ARN = aws.ToString(summaryData.Arn)
	scp.Name = aws.ToString(summaryData.Name)
	scp.Description = aws.ToString(summaryData.Description)
	scp.PolicyDocument = docRaw
	scp.Policy = parsed
	scp.TargetIDs = targetIDs

	return &scp
}

// listTargets lists all targets (OUs/accounts) a policy is attached to.
func listTargets(ctx context.Context, client *organizations.Client, s *store.Store, policyID string) []string {
	targets := make([]string, 0)

	paginator := organizations.NewListTargetsForPolicyPaginator(client,
		&organizations.ListTargetsForPolicyInput{
			PolicyId: aws.String(policyID),
		},
	)

	cliCmd := fmt.Sprintf("aws organizations list-targets-for-policy --policy-id %s", policyID)

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordSCPError(s, "organizations:list_targets_for_policy", err)
			break
		}

		for _, target := range page.Targets {
			targets = append(targets, aws.ToString(target.TargetId))
		}
	}

	return targets
}

// parsePolicyDocument parses a raw JSON policy document into a models.Policy.
func parsePolicyDocument(raw string) (*models.Policy, error) {
	// Parse into a generic map first to handle case-insensitive keys
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &rawMap); err != nil {
		return nil, fmt.Errorf("unmarshal policy document: %w", err)
	}

	p := models.NewPolicy()

	// Extract Version (try both cases)
	if v, ok := rawMap["Version"]; ok {
		json.Unmarshal(v, &p.Version)
	} else if v, ok := rawMap["version"]; ok {
		json.Unmarshal(v, &p.Version)
	}

	// Extract Statement (try both cases)
	if v, ok := rawMap["Statement"]; ok {
		json.Unmarshal(v, &p.Statement)
	} else if v, ok := rawMap["statement"]; ok {
		json.Unmarshal(v, &p.Statement)
	}

	return &p, nil
}

// recordSCPError adds an ErrorInfo to the store for an Organizations API failure.
func recordSCPError(s *store.Store, apiCall string, err error) {
	s.AddError(models.ErrorInfo{
		Source:      fmt.Sprintf("Organizations API (%s)", apiCall),
		Message:     err.Error(),
		Recoverable: false,
	})
}
