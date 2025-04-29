package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/builder"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/node"
	"github.com/prysmaticlabs/prysm/v5/cmd"
	blockchaincmd "github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/execution"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/flags"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/storage"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/sync/backfill"
	bflags "github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/sync/backfill/flags"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/sync/checkpoint"
	"github.com/prysmaticlabs/prysm/v5/cmd/beacon-chain/sync/genesis"
	"github.com/prysmaticlabs/prysm/v5/config/features"
	"github.com/prysmaticlabs/prysm/v5/io/file"
	"github.com/prysmaticlabs/prysm/v5/io/logs"
	"github.com/prysmaticlabs/prysm/v5/runtime/debug"
	"github.com/prysmaticlabs/prysm/v5/runtime/fdlimits"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Args struct {
	DataDir           string
	ChainConfigFile   string
	ExecutionEndpoint string
	JWTSecret         string
	GenesisState      string
}

var log = logrus.WithField("prefix", "main")

var appFlags = []cli.Flag{
	flags.DepositContractFlag,
	flags.ExecutionEngineEndpoint,
	flags.ExecutionEngineHeaders,
	flags.ExecutionJWTSecretFlag,
	flags.RPCHost,
	flags.RPCPort,
	flags.CertFlag,
	flags.KeyFlag,
	flags.HTTPModules,
	flags.HTTPServerHost,
	flags.HTTPServerPort,
	flags.HTTPServerCorsDomain,
	flags.MinSyncPeers,
	flags.ContractDeploymentBlock,
	flags.SetGCPercent,
	flags.BlockBatchLimit,
	flags.BlockBatchLimitBurstFactor,
	flags.BlobBatchLimit,
	flags.BlobBatchLimitBurstFactor,
	flags.InteropMockEth1DataVotesFlag,
	flags.SlotsPerArchivedPoint,
	flags.DisableDebugRPCEndpoints,
	flags.SubscribeToAllSubnets,
	flags.HistoricalSlasherNode,
	flags.ChainID,
	flags.NetworkID,
	flags.WeakSubjectivityCheckpoint,
	flags.Eth1HeaderReqLimit,
	flags.MinPeersPerSubnet,
	flags.MaxConcurrentDials,
	flags.SuggestedFeeRecipient,
	flags.TerminalTotalDifficultyOverride,
	flags.TerminalBlockHashOverride,
	flags.TerminalBlockHashActivationEpochOverride,
	flags.MevRelayEndpoint,
	flags.MaxBuilderEpochMissedSlots,
	flags.MaxBuilderConsecutiveMissedSlots,
	flags.EngineEndpointTimeoutSeconds,
	flags.LocalBlockValueBoost,
	flags.MinBuilderBid,
	flags.MinBuilderDiff,
	flags.BeaconDBPruning,
	flags.PrunerRetentionEpochs,
	flags.EnableBuilderSSZ,
	cmd.BackupWebhookOutputDir,
	cmd.MinimalConfigFlag,
	cmd.E2EConfigFlag,
	cmd.RPCMaxPageSizeFlag,
	cmd.BootstrapNode,
	cmd.NoDiscovery,
	cmd.StaticPeers,
	cmd.RelayNode,
	cmd.P2PUDPPort,
	cmd.P2PQUICPort,
	cmd.P2PTCPPort,
	cmd.P2PIP,
	cmd.P2PHost,
	cmd.P2PHostDNS,
	cmd.P2PMaxPeers,
	cmd.P2PPrivKey,
	cmd.P2PStaticID,
	cmd.P2PMetadata,
	cmd.P2PAllowList,
	cmd.P2PDenyList,
	cmd.PubsubQueueSize,
	cmd.DataDirFlag,
	cmd.VerbosityFlag,
	cmd.EnableTracingFlag,
	cmd.TracingProcessNameFlag,
	cmd.TracingEndpointFlag,
	cmd.TraceSampleFractionFlag,
	cmd.MonitoringHostFlag,
	flags.MonitoringPortFlag,
	cmd.DisableMonitoringFlag,
	cmd.ClearDB,
	cmd.ForceClearDB,
	cmd.LogFormat,
	cmd.MaxGoroutines,
	debug.PProfFlag,
	debug.PProfAddrFlag,
	debug.PProfPortFlag,
	debug.MemProfileRateFlag,
	debug.CPUProfileFlag,
	debug.TraceFlag,
	debug.BlockProfileRateFlag,
	debug.MutexProfileFractionFlag,
	cmd.LogFileName,
	cmd.EnableUPnPFlag,
	cmd.ConfigFileFlag,
	cmd.ChainConfigFileFlag,
	cmd.GrpcMaxCallRecvMsgSizeFlag,
	cmd.AcceptTosFlag,
	cmd.RestoreSourceFileFlag,
	cmd.RestoreTargetDirFlag,
	cmd.ValidatorMonitorIndicesFlag,
	cmd.ApiTimeoutFlag,
	checkpoint.BlockPath,
	checkpoint.StatePath,
	checkpoint.RemoteURL,
	genesis.StatePath,
	genesis.BeaconAPIURL,
	flags.SlasherDirFlag,
	flags.SlasherFlag,
	flags.JwtId,
	storage.BlobStoragePathFlag,
	storage.BlobRetentionEpochFlag,
	storage.BlobStorageLayout,
	bflags.EnableExperimentalBackfill,
	bflags.BackfillBatchSize,
	bflags.BackfillWorkerCount,
	bflags.BackfillOldestSlot,
}

func before(ctx *cli.Context) error {
	// Load flags from config file, if specified.
	if err := cmd.LoadFlagsFromConfig(ctx, appFlags); err != nil {
		return errors.Wrap(err, "failed to load flags from config file")
	}

	// format := ctx.String(cmd.LogFormat.Name)

	// switch format {
	// case "text":
	// 	formatter := new(prefixed.TextFormatter)
	// 	formatter.TimestampFormat = time.DateTime
	// 	formatter.FullTimestamp = true

	// 	// If persistent log files are written - we disable the log messages coloring because
	// 	// the colors are ANSI codes and seen as gibberish in the log files.
	// 	formatter.DisableColors = ctx.String(cmd.LogFileName.Name) != ""
	// 	logrus.SetFormatter(formatter)
	// case "fluentd":
	// 	f := joonix.NewFormatter()

	// 	if err := joonix.DisableTimestampFormat(f); err != nil {
	// 		panic(err) // lint:nopanic -- This shouldn't happen, but crashing immediately at startup is OK.
	// 	}

	// 	logrus.SetFormatter(f)
	// case "json":
	// 	logrus.SetFormatter(&logrus.JSONFormatter{})
	// case "journald":
	// 	if err := journald.Enable(); err != nil {
	// 		return err
	// 	}
	// default:
	// 	return fmt.Errorf("unknown log format %s", format)
	// }

	logFileName := ctx.String(cmd.LogFileName.Name)
	if logFileName != "" {
		if err := logs.ConfigurePersistentLogging(logFileName); err != nil {
			log.WithError(err).Error("Failed to configuring logging to disk.")
		}
	}

	if err := cmd.ExpandSingleEndpointIfFile(ctx, flags.ExecutionEngineEndpoint); err != nil {
		return errors.Wrap(err, "failed to expand single endpoint")
	}

	// if ctx.IsSet(flags.SetGCPercent.Name) {
	// 	runtimeDebug.SetGCPercent(ctx.Int(flags.SetGCPercent.Name))
	// }

	if err := debug.Setup(ctx); err != nil {
		return errors.Wrap(err, "failed to setup debug")
	}

	if err := fdlimits.SetMaxFdLimits(); err != nil {
		return errors.Wrap(err, "failed to set max fd limits")
	}

	if err := features.ValidateNetworkFlags(ctx); err != nil {
		return errors.Wrap(err, "provided multiple network flags")
	}

	return cmd.ValidateNoArgs(ctx)
}

func startNode(ctx *cli.Context, cancel context.CancelFunc) error {
	// Fix data dir for Windows users.
	outdatedDataDir := filepath.Join(file.HomeDir(), "AppData", "Roaming", "Eth2")
	currentDataDir := ctx.String(cmd.DataDirFlag.Name)
	if err := cmd.FixDefaultDataDir(outdatedDataDir, currentDataDir); err != nil {
		return err
	}

	// verify if ToS accepted
	// if err := tos.VerifyTosAcceptedOrPrompt(ctx); err != nil {
	// 	return err
	// }

	blockchainFlagOpts, err := blockchaincmd.FlagOptions(ctx)
	if err != nil {
		return err
	}
	executionFlagOpts, err := execution.FlagOptions(ctx)
	if err != nil {
		return err
	}
	builderFlagOpts, err := builder.FlagOptions(ctx)
	if err != nil {
		return err
	}
	opts := []node.Option{
		node.WithBlockchainFlagOptions(blockchainFlagOpts),
		node.WithExecutionChainOptions(executionFlagOpts),
		node.WithBuilderFlagOptions(builderFlagOpts),
	}

	optFuncs := []func(*cli.Context) ([]node.Option, error){
		genesis.BeaconNodeOptions,
		checkpoint.BeaconNodeOptions,
		storage.BeaconNodeOptions,
		backfill.BeaconNodeOptions,
	}
	for _, of := range optFuncs {
		ofo, err := of(ctx)
		if err != nil {
			return err
		}
		if ofo != nil {
			opts = append(opts, ofo...)
		}
	}

	beacon, err := node.New(ctx, cancel, opts...)
	if err != nil {
		return fmt.Errorf("unable to start beacon node: %w", err)
	}
	beacon.Start()
	return nil
}

func runApp(args Args) {

	app := &cli.App{
		Flags: appFlags,
		Action: func(ctx *cli.Context) error {

			_, cancel := context.WithCancel(context.Background())
			if err := startNode(ctx, cancel); err != nil {
				log.Fatal(err.Error())
				return err
			}

			return nil
		},
		Before: before,
	}

	cliArgs := []string{
		"app",
		"--datadir", args.DataDir,
		"--jwt-secret", args.JWTSecret,
		"--genesis-state", args.GenesisState,
		"--chain-config-file", args.ChainConfigFile,
		"--execution-endpoint", args.ExecutionEndpoint,
	}

	err := app.Run(cliArgs)
	if err != nil {
		log.Fatalf("Failed to run CLI app: %v", err)
	}
}

func main() {

	args := &Args{
		DataDir:           "data",
		ChainConfigFile:   "config.yaml",
		ExecutionEndpoint: "http://127.0.0.1:8551",
		JWTSecret:         "jwtsecret",
		GenesisState:      "genesis.ssz",
	}

	runApp(*args)
}
