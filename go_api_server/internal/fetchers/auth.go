package fetchers

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const defaultRegion = "ap-south-1"

// AuthStateSnapshot holds a point-in-time copy of auth state (safe for JSON serialization).
type AuthStateSnapshot struct {
	Authenticated bool       `json:"authenticated"`
	AccountID     string     `json:"account_id"`
	Region        string     `json:"region"`
	Error         string     `json:"error"`
	AuthMode      string     `json:"auth_mode"`
	AWSConfig     *aws.Config `json:"-"`
}

// AuthState is a thread-safe wrapper around authentication state.
type AuthState struct {
	mu       sync.Mutex
	snapshot AuthStateSnapshot
}

// Get returns a copy of the current auth state.
func (a *AuthState) Get() AuthStateSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshot
}

// Set updates the auth state under lock.
func (a *AuthState) Set(s AuthStateSnapshot) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.snapshot = s
}

// IMDSMetadata holds EC2 instance metadata retrieved via IMDSv2.
type IMDSMetadata struct {
	Available        bool   `json:"available"`
	InstanceID       string `json:"instance_id"`
	InstanceType     string `json:"instance_type"`
	AvailabilityZone string `json:"availability_zone"`
	Region           string `json:"region"`
	IAMRole          string `json:"iam_role"`
	AccountID        string `json:"account_id"`
	AuthMode         string `json:"auth_mode"`
}

// ManualCredentialOpts holds options for manual credential configuration.
type ManualCredentialOpts struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	ProfileName     string `json:"profile_name"`
	Region          string `json:"region"`
}

// ValidateCredentials validates the default AWS credential chain via STS GetCallerIdentity.
// Returns the aws.Config, account ID, and any error.
func ValidateCredentials(ctx context.Context, region string) (*aws.Config, string, error) {
	if region == "" {
		region = defaultRegion
	}

	slog.Info("aws_api_call",
		"component", "fetcher",
		"cli_command", "aws sts get-caller-identity --region "+region,
	)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, "", err
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, "", err
	}

	accountID := aws.ToString(identity.Account)
	return &cfg, accountID, nil
}

// ValidateManualCredentials validates manually-provided AWS credentials via STS GetCallerIdentity.
// Supports either a named profile or static access key credentials.
func ValidateManualCredentials(ctx context.Context, opts ManualCredentialOpts) (*aws.Config, string, error) {
	region := opts.Region
	if region == "" {
		region = defaultRegion
	}

	slog.Info("aws_api_call",
		"component", "fetcher",
		"cli_command", "aws sts get-caller-identity --region "+region,
	)

	var cfg aws.Config
	var err error

	if opts.ProfileName != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(opts.ProfileName),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					opts.AccessKeyID,
					opts.SecretAccessKey,
					opts.SessionToken,
				),
			),
		)
	}
	if err != nil {
		return nil, "", err
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, "", err
	}

	accountID := aws.ToString(identity.Account)
	return &cfg, accountID, nil
}

// iamInfo is used to parse the JSON response from IMDS iam/info endpoint.
type iamInfo struct {
	InstanceProfileARN string `json:"InstanceProfileArn"`
}

// GetIMDSMetadata fetches EC2 instance metadata via IMDSv2.
// If IMDS is not available (e.g., not running on EC2), returns &IMDSMetadata{Available: false} with nil error.
func GetIMDSMetadata(ctx context.Context) (*IMDSMetadata, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Info("imds_unavailable", "component", "fetcher", "reason", "config_load_failed")
		return &IMDSMetadata{Available: false}, nil
	}

	client := imds.NewFromConfig(cfg)

	// Fetch instance-id
	instanceIDOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "instance-id"})
	if err != nil {
		slog.Info("imds_unavailable", "component", "fetcher")
		return &IMDSMetadata{Available: false}, nil
	}
	instanceID := readIMDSOutput(instanceIDOut.Content)

	// Fetch instance-type
	instanceTypeOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "instance-type"})
	if err != nil {
		slog.Info("imds_unavailable", "component", "fetcher")
		return &IMDSMetadata{Available: false}, nil
	}
	instanceType := readIMDSOutput(instanceTypeOut.Content)

	// Fetch availability-zone
	azOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "placement/availability-zone"})
	if err != nil {
		slog.Info("imds_unavailable", "component", "fetcher")
		return &IMDSMetadata{Available: false}, nil
	}
	az := readIMDSOutput(azOut.Content)

	// Parse region from availability zone (strip last char)
	region := ""
	if len(az) > 1 {
		region = az[:len(az)-1]
	}

	// Fetch IAM info
	iamInfoOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/info"})
	var iamRole string
	var accountID string
	if err == nil {
		iamInfoRaw := readIMDSOutput(iamInfoOut.Content)
		var info iamInfo
		if jsonErr := json.Unmarshal([]byte(iamInfoRaw), &info); jsonErr == nil {
			iamRole = info.InstanceProfileARN
		}
	}

	// Try to get account ID from STS if we have credentials
	if accountID == "" {
		stsCfg, loadErr := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if loadErr == nil {
			stsClient := sts.NewFromConfig(stsCfg)
			identity, stsErr := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if stsErr == nil {
				accountID = aws.ToString(identity.Account)
			}
		}
	}

	metadata := &IMDSMetadata{
		Available:        true,
		InstanceID:       instanceID,
		InstanceType:     instanceType,
		AvailabilityZone: az,
		Region:           region,
		IAMRole:          iamRole,
		AccountID:        accountID,
		AuthMode:         "instance_metadata",
	}

	slog.Info("imds_metadata_fetched",
		"component", "fetcher",
		"region", region,
		"iam_role", iamRole,
	)

	return metadata, nil
}

// readIMDSOutput reads all bytes from an io.ReadCloser and returns as string.
func readIMDSOutput(r interface{ Read([]byte) (int, error) }) string {
	buf := make([]byte, 0, 256)
	tmp := make([]byte, 256)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return string(buf)
}
