package fetchers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controlcatalog"
	cctypes "github.com/aws/aws-sdk-go-v2/service/controlcatalog/types"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// FetchCatalog fetches the full AWS Control Catalog (domains, objectives,
// common controls, controls, and control mappings) and populates the store.
func FetchCatalog(ctx context.Context, cfg aws.Config, s *store.Store) error {
	client := controlcatalog.NewFromConfig(cfg)

	fetchDomains(ctx, client, s)
	fetchObjectives(ctx, client, s)
	fetchCommonControls(ctx, client, s)
	fetchControls(ctx, client, s)
	fetchControlMappings(ctx, client, s)

	slog.Info("catalog_fetch_complete",
		"component", "fetcher",
		"domain_count", len(s.AllCatalogDomains()),
		"objective_count", len(s.AllCatalogObjectives()),
		"common_control_count", len(s.AllCatalogCommonControls()),
		"control_count", len(s.AllCatalogControls()),
	)

	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// fetchDomains paginates ListDomains and stores each CatalogDomain.
func fetchDomains(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	paginator := controlcatalog.NewListDomainsPaginator(client, &controlcatalog.ListDomainsInput{})
	cliCmd := "aws controlcatalog list-domains"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordCatalogError(s, "controlcatalog:ListDomains", err)
			break
		}

		for _, entry := range page.Domains {
			domain := models.CatalogDomain{
				ARN:         aws.ToString(entry.Arn),
				Name:        aws.ToString(entry.Name),
				Description: aws.ToString(entry.Description),
			}
			s.PutCatalogDomain(domain)
		}
	}
}

// fetchObjectives paginates ListObjectives per domain. If no domains exist,
// falls back to an unfiltered listing.
func fetchObjectives(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	domains := s.AllCatalogDomains()
	if len(domains) == 0 {
		fetchObjectivesUnfiltered(ctx, client, s)
		return
	}

	for _, domain := range domains {
		input := &controlcatalog.ListObjectivesInput{
			ObjectiveFilter: &cctypes.ObjectiveFilter{
				Domains: []cctypes.DomainResourceFilter{
					{Arn: aws.String(domain.ARN)},
				},
			},
		}
		paginator := controlcatalog.NewListObjectivesPaginator(client, input)
		cliCmd := fmt.Sprintf("aws controlcatalog list-objectives --objective-filter Domains=[{Arn=%s}]", domain.ARN)

		for paginator.HasMorePages() {
			LogAPICall(cliCmd, "ok")

			page, err := paginator.NextPage(ctx)
			if err != nil {
				LogAPICallError(cliCmd, err.Error())
				recordCatalogError(s, "controlcatalog:ListObjectives", err)
				break
			}

			for _, entry := range page.Objectives {
				objective := models.CatalogObjective{
					ARN:         aws.ToString(entry.Arn),
					Name:        aws.ToString(entry.Name),
					Description: aws.ToString(entry.Description),
					DomainARN:   domain.ARN,
				}
				s.PutCatalogObjective(objective)
			}
		}
	}
}

// fetchObjectivesUnfiltered lists all objectives without a domain filter.
func fetchObjectivesUnfiltered(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	paginator := controlcatalog.NewListObjectivesPaginator(client, &controlcatalog.ListObjectivesInput{})
	cliCmd := "aws controlcatalog list-objectives"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordCatalogError(s, "controlcatalog:ListObjectives", err)
			break
		}

		for _, entry := range page.Objectives {
			domainARN := ""
			if entry.Domain != nil {
				domainARN = aws.ToString(entry.Domain.Arn)
			}
			objective := models.CatalogObjective{
				ARN:         aws.ToString(entry.Arn),
				Name:        aws.ToString(entry.Name),
				Description: aws.ToString(entry.Description),
				DomainARN:   domainARN,
			}
			s.PutCatalogObjective(objective)
		}
	}
}

// fetchCommonControls paginates ListCommonControls per objective. If no
// objectives exist, falls back to an unfiltered listing.
func fetchCommonControls(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	objectives := s.AllCatalogObjectives()
	if len(objectives) == 0 {
		fetchCommonControlsUnfiltered(ctx, client, s)
		return
	}

	for _, objective := range objectives {
		input := &controlcatalog.ListCommonControlsInput{
			CommonControlFilter: &cctypes.CommonControlFilter{
				Objectives: []cctypes.ObjectiveResourceFilter{
					{Arn: aws.String(objective.ARN)},
				},
			},
		}
		paginator := controlcatalog.NewListCommonControlsPaginator(client, input)
		cliCmd := fmt.Sprintf("aws controlcatalog list-common-controls --common-control-filter Objectives=[{Arn=%s}]", objective.ARN)

		for paginator.HasMorePages() {
			LogAPICall(cliCmd, "ok")

			page, err := paginator.NextPage(ctx)
			if err != nil {
				LogAPICallError(cliCmd, err.Error())
				recordCatalogError(s, "controlcatalog:ListCommonControls", err)
				break
			}

			for _, entry := range page.CommonControls {
				cc := models.CatalogCommonControl{
					ARN:          aws.ToString(entry.Arn),
					Name:         aws.ToString(entry.Name),
					Description:  aws.ToString(entry.Description),
					ObjectiveARN: objective.ARN,
					DomainARN:    objective.DomainARN,
				}
				s.PutCatalogCommonControl(cc)
			}
		}
	}
}

// fetchCommonControlsUnfiltered lists all common controls without an objective filter.
func fetchCommonControlsUnfiltered(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	paginator := controlcatalog.NewListCommonControlsPaginator(client, &controlcatalog.ListCommonControlsInput{})
	cliCmd := "aws controlcatalog list-common-controls"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordCatalogError(s, "controlcatalog:ListCommonControls", err)
			break
		}

		for _, entry := range page.CommonControls {
			objectiveARN := ""
			if entry.Objective != nil {
				objectiveARN = aws.ToString(entry.Objective.Arn)
			}
			domainARN := ""
			if entry.Domain != nil {
				domainARN = aws.ToString(entry.Domain.Arn)
			}
			cc := models.CatalogCommonControl{
				ARN:          aws.ToString(entry.Arn),
				Name:         aws.ToString(entry.Name),
				Description:  aws.ToString(entry.Description),
				ObjectiveARN: objectiveARN,
				DomainARN:    domainARN,
			}
			s.PutCatalogCommonControl(cc)
		}
	}
}

// fetchControls paginates ListControls and stores each CatalogControl.
func fetchControls(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	paginator := controlcatalog.NewListControlsPaginator(client, &controlcatalog.ListControlsInput{})
	cliCmd := "aws controlcatalog list-controls"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordCatalogError(s, "controlcatalog:ListControls", err)
			break
		}

		for _, entry := range page.Controls {
			ctrl := mapCatalogControl(entry)
			s.PutCatalogControl(ctrl)
		}
	}
}

// mapCatalogControl maps a ControlSummary to a models.CatalogControl.
func mapCatalogControl(entry cctypes.ControlSummary) models.CatalogControl {
	ctrl := models.NewCatalogControl()
	ctrl.ARN = aws.ToString(entry.Arn)
	ctrl.Name = aws.ToString(entry.Name)
	ctrl.Description = aws.ToString(entry.Description)

	// Behavior
	ctrl.Behavior = mapBehavior(entry.Behavior)

	// Severity
	ctrl.Severity = mapSeverity(entry.Severity)

	// Aliases
	if len(entry.Aliases) > 0 {
		ctrl.Aliases = make([]string, len(entry.Aliases))
		copy(ctrl.Aliases, entry.Aliases)
	}

	// GovernedResources
	if len(entry.GovernedResources) > 0 {
		ctrl.GovernedResources = make([]string, len(entry.GovernedResources))
		copy(ctrl.GovernedResources, entry.GovernedResources)
	}

	// Implementation
	if entry.Implementation != nil {
		ctrl.ImplementationType = aws.ToString(entry.Implementation.Type)
		ctrl.ImplementationIdentifier = aws.ToString(entry.Implementation.Identifier)
	}

	// CreateTime
	if entry.CreateTime != nil {
		ctrl.CreateTime = entry.CreateTime.Format("2006-01-02T15:04:05Z07:00")
	}

	return ctrl
}

// mapBehavior converts the SDK ControlBehavior enum to our model enum.
func mapBehavior(b cctypes.ControlBehavior) models.CatalogBehavior {
	switch b {
	case cctypes.ControlBehaviorPreventive:
		return models.BehaviorPreventive
	case cctypes.ControlBehaviorDetective:
		return models.BehaviorDetective
	case cctypes.ControlBehaviorProactive:
		return models.BehaviorProactive
	default:
		return models.BehaviorPreventive
	}
}

// mapSeverity converts the SDK ControlSeverity enum to our model enum.
func mapSeverity(s cctypes.ControlSeverity) models.CatalogSeverity {
	switch s {
	case cctypes.ControlSeverityLow:
		return models.SeverityLow
	case cctypes.ControlSeverityMedium:
		return models.SeverityMedium
	case cctypes.ControlSeverityHigh:
		return models.SeverityHigh
	case cctypes.ControlSeverityCritical:
		return models.SeverityCritical
	default:
		return models.SeverityMedium
	}
}

// fetchControlMappings paginates ListControlMappings for COMMON_CONTROL type
// and stores the mapping: control_arn → []common_control_arn.
func fetchControlMappings(ctx context.Context, client *controlcatalog.Client, s *store.Store) {
	mappings := make(map[string][]string)

	input := &controlcatalog.ListControlMappingsInput{
		Filter: &cctypes.ControlMappingFilter{
			MappingTypes: []cctypes.MappingType{cctypes.MappingTypeCommonControl},
		},
	}
	paginator := controlcatalog.NewListControlMappingsPaginator(client, input)
	cliCmd := "aws controlcatalog list-control-mappings --filter MappingTypes=[COMMON_CONTROL]"

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordCatalogError(s, "controlcatalog:ListControlMappings", err)
			break
		}

		for _, entry := range page.ControlMappings {
			if entry.MappingType != cctypes.MappingTypeCommonControl {
				continue
			}
			ctrlARN := aws.ToString(entry.ControlArn)
			if ctrlARN == "" {
				continue
			}
			// Extract common control ARN from the union type
			if ccMapping, ok := entry.Mapping.(*cctypes.MappingMemberCommonControl); ok {
				ccARN := aws.ToString(ccMapping.Value.CommonControlArn)
				if ccARN != "" {
					mappings[ctrlARN] = append(mappings[ctrlARN], ccARN)
				}
			}
		}
	}

	for ctrlARN, ccARNs := range mappings {
		s.PutControlMapping(ctrlARN, ccARNs)
	}

	slog.Info("control_mappings_fetched", "component", "fetcher", "mapping_count", len(mappings))
}

// recordCatalogError adds an ErrorInfo to the store for a Control Catalog API failure.
func recordCatalogError(s *store.Store, apiCall string, err error) {
	s.AddError(models.ErrorInfo{
		Source:      fmt.Sprintf("Control Catalog API (%s)", apiCall),
		Message:     err.Error(),
		Recoverable: false,
	})
}
