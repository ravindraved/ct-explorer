package fetchers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"golang.org/x/sync/errgroup"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// FetchOrganization fetches the full AWS Organization tree (roots → OUs → accounts)
// and populates the store. It mirrors the Python org_fetcher.fetch_organization behavior.
func FetchOrganization(ctx context.Context, cfg aws.Config, s *store.Store) error {
	client := organizations.NewFromConfig(cfg)

	// 1. Get the root OU
	cliCmd := "aws organizations list-roots"
	LogAPICall(cliCmd, "ok")

	rootsOut, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		LogAPICallError(cliCmd, err.Error())
		recordOrgError(s, "organizations:list_roots", err)
		return fmt.Errorf("list roots: %w", err)
	}

	if len(rootsOut.Roots) == 0 {
		slog.Warn("no_roots_found", "component", "fetcher")
		return nil
	}

	root := rootsOut.Roots[0]
	rootID := aws.ToString(root.Id)
	s.SetRootOUID(rootID)

	rootOU := models.NewOrgUnit()
	rootOU.ID = rootID
	rootOU.Name = aws.ToString(root.Name)
	rootOU.ARN = aws.ToString(root.Arn)
	s.PutOU(rootOU)

	// 2. Recursive OU traversal
	slog.Info("fetching_ous", "component", "fetcher")

	ouCount, accountCount, err := traverse(ctx, client, s, rootID)
	if err != nil {
		return err
	}

	slog.Info("fetch_complete",
		"component", "fetcher",
		"ou_count", ouCount+1,
		"account_count", accountCount,
	)

	return nil
}

// traverse recursively walks the OU tree starting from parentID.
// It fetches child OUs and accounts concurrently using errgroup.
// Returns (ouCount, accountCount, error).
func traverse(ctx context.Context, client *organizations.Client, s *store.Store, parentID string) (int, int, error) {
	childOUIDs, err := listChildOUs(ctx, client, s, parentID)
	if err != nil {
		// Non-fatal: logged and recorded in store already
		slog.Warn("list_child_ous_partial_failure", "component", "fetcher", "parent_id", parentID)
	}

	// Update parent OU with children IDs
	if parentOU, ok := s.GetOU(parentID); ok {
		parentOU.ChildrenIDs = childOUIDs
		s.PutOU(parentOU)
	}

	ouCount := 0
	accountCount := 0

	// Use errgroup for concurrent traversal of child OUs
	g, gctx := errgroup.WithContext(ctx)

	type traverseResult struct {
		ous      int
		accounts int
	}
	results := make([]traverseResult, len(childOUIDs))

	for i, childID := range childOUIDs {
		i, childID := i, childID
		ouCount++
		g.Go(func() error {
			childOUs, childAccts, err := traverse(gctx, client, s, childID)
			results[i] = traverseResult{ous: childOUs, accounts: childAccts}
			return err
		})
	}

	if err := g.Wait(); err != nil {
		// Log but don't fail the whole operation
		slog.Warn("traverse_partial_failure", "component", "fetcher", "parent_id", parentID, "error", err)
	}

	for _, r := range results {
		ouCount += r.ous
		accountCount += r.accounts
	}

	// Fetch accounts for this OU
	acctIDs, err := listAccounts(ctx, client, s, parentID)
	if err != nil {
		slog.Warn("list_accounts_partial_failure", "component", "fetcher", "parent_id", parentID)
	}

	// Update parent OU with account IDs
	if parentOU, ok := s.GetOU(parentID); ok {
		parentOU.AccountIDs = acctIDs
		s.PutOU(parentOU)
	}

	accountCount += len(acctIDs)

	return ouCount, accountCount, nil
}

// listChildOUs lists child OUs under parentID using pagination.
// For each child OU found, it fetches the OU details and stores them.
func listChildOUs(ctx context.Context, client *organizations.Client, s *store.Store, parentID string) ([]string, error) {
	childIDs := make([]string, 0)

	paginator := organizations.NewListOrganizationalUnitsForParentPaginator(client,
		&organizations.ListOrganizationalUnitsForParentInput{
			ParentId: aws.String(parentID),
		},
	)

	cliCmd := fmt.Sprintf("aws organizations list-organizational-units-for-parent --parent-id %s", parentID)

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordOrgError(s, "organizations:list_organizational_units_for_parent", err)
			break
		}

		for _, ouItem := range page.OrganizationalUnits {
			childID := aws.ToString(ouItem.Id)

			ou := models.NewOrgUnit()
			ou.ID = childID
			ou.Name = aws.ToString(ouItem.Name)
			ou.ARN = aws.ToString(ouItem.Arn)
			ou.ParentID = parentID
			s.PutOU(ou)
			childIDs = append(childIDs, childID)
		}
	}

	return childIDs, nil
}

// listAccounts lists accounts under parentID using pagination.
func listAccounts(ctx context.Context, client *organizations.Client, s *store.Store, parentID string) ([]string, error) {
	accountIDs := make([]string, 0)

	paginator := organizations.NewListAccountsForParentPaginator(client,
		&organizations.ListAccountsForParentInput{
			ParentId: aws.String(parentID),
		},
	)

	cliCmd := fmt.Sprintf("aws organizations list-accounts-for-parent --parent-id %s", parentID)

	for paginator.HasMorePages() {
		LogAPICall(cliCmd, "ok")

		page, err := paginator.NextPage(ctx)
		if err != nil {
			LogAPICallError(cliCmd, err.Error())
			recordOrgError(s, "organizations:list_accounts_for_parent", err)
			break
		}

		for _, acct := range page.Accounts {
			account := models.Account{
				ID:     aws.ToString(acct.Id),
				Name:   aws.ToString(acct.Name),
				ARN:    aws.ToString(acct.Arn),
				Email:  aws.ToString(acct.Email),
				Status: string(acct.Status),
				OUID:   parentID,
			}
			s.PutAccount(account)
			accountIDs = append(accountIDs, account.ID)
		}
	}

	return accountIDs, nil
}

// recordOrgError adds an ErrorInfo to the store for an Organizations API failure.
func recordOrgError(s *store.Store, apiCall string, err error) {
	s.AddError(models.ErrorInfo{
		Source:      fmt.Sprintf("Organizations API (%s)", apiCall),
		Message:     err.Error(),
		Recoverable: true,
	})
}
