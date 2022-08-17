package network

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"
	campaigntypes "github.com/tendermint/spn/x/campaign/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"
	rewardtypes "github.com/tendermint/spn/x/reward/types"
	"golang.org/x/sync/errgroup"

	"github.com/ignite/cli/ignite/pkg/cosmoserror"
	"github.com/ignite/cli/ignite/pkg/events"
	"github.com/ignite/cli/ignite/services/network/networktypes"
)

// ErrObjectNotFound is returned when the query returns a not found error.
var ErrObjectNotFound = errors.New("query object not found")

// ChainLaunch fetches the chain launch from Network by launch id.
func (n Network) ChainLaunch(ctx context.Context, id uint64) (networktypes.ChainLaunch, error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching chain information"))

	res, err := n.launchQuery.
		Chain(ctx,
			&launchtypes.QueryGetChainRequest{
				LaunchID: id,
			},
		)
	if err != nil {
		return networktypes.ChainLaunch{}, err
	}

	return networktypes.ToChainLaunch(res.Chain), nil
}

// ChainLaunchesWithReward fetches the chain launches with rewards from Network
func (n Network) ChainLaunchesWithReward(ctx context.Context) ([]networktypes.ChainLaunch, error) {
	g, ctx := errgroup.WithContext(ctx)

	n.ev.Send(events.New(events.StatusOngoing, "Fetching chains information"))
	res, err := n.launchQuery.
		ChainAll(ctx, &launchtypes.QueryAllChainRequest{})
	if err != nil {
		return nil, err
	}

	n.ev.Send(events.New(events.StatusOngoing, "Fetching reward information"))
	var chainLaunches []networktypes.ChainLaunch
	var mu sync.Mutex

	// Parse fetched chains and fetch rewards
	for _, chain := range res.Chain {
		chain := chain
		g.Go(func() error {
			chainLaunch := networktypes.ToChainLaunch(chain)
			reward, err := n.ChainReward(ctx, chain.LaunchID)
			if err != nil && err != ErrObjectNotFound {
				return err
			}
			chainLaunch.Reward = reward.RemainingCoins.String()
			mu.Lock()
			chainLaunches = append(chainLaunches, chainLaunch)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// sort filenames by launch id
	sort.Slice(chainLaunches, func(i, j int) bool {
		return chainLaunches[i].ID > chainLaunches[j].ID
	})
	return chainLaunches, nil
}

// GenesisInformation returns all the information to construct the genesis from a chain ID
func (n Network) GenesisInformation(ctx context.Context, launchID uint64) (gi networktypes.GenesisInformation, err error) {
	genAccs, err := n.GenesisAccounts(ctx, launchID)
	if err != nil {
		return gi, errors.Wrap(err, "error querying genesis accounts")
	}

	vestingAccs, err := n.VestingAccounts(ctx, launchID)
	if err != nil {
		return gi, errors.Wrap(err, "error querying vesting accounts")
	}

	genVals, err := n.GenesisValidators(ctx, launchID)
	if err != nil {
		return gi, errors.Wrap(err, "error querying genesis validators")
	}

	return networktypes.NewGenesisInformation(genAccs, vestingAccs, genVals), nil
}

// GenesisAccounts returns the list of approved genesis accounts for a launch from SPN
func (n Network) GenesisAccounts(ctx context.Context, launchID uint64) (genAccs []networktypes.GenesisAccount, err error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching genesis accounts"))
	res, err := n.launchQuery.
		GenesisAccountAll(ctx,
			&launchtypes.QueryAllGenesisAccountRequest{
				LaunchID: launchID,
			},
		)
	if err != nil {
		return genAccs, err
	}

	for _, acc := range res.GenesisAccount {
		genAccs = append(genAccs, networktypes.ToGenesisAccount(acc))
	}

	return genAccs, nil
}

// VestingAccounts returns the list of approved genesis vesting accounts for a launch from SPN
func (n Network) VestingAccounts(ctx context.Context, launchID uint64) (vestingAccs []networktypes.VestingAccount, err error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching genesis vesting accounts"))
	res, err := n.launchQuery.
		VestingAccountAll(ctx,
			&launchtypes.QueryAllVestingAccountRequest{
				LaunchID: launchID,
			},
		)
	if err != nil {
		return vestingAccs, err
	}

	for i, acc := range res.VestingAccount {
		parsedAcc, err := networktypes.ToVestingAccount(acc)
		if err != nil {
			return vestingAccs, errors.Wrapf(err, "error parsing vesting account %d", i)
		}

		vestingAccs = append(vestingAccs, parsedAcc)
	}

	return vestingAccs, nil
}

// GenesisValidators returns the list of approved genesis validators for a launch from SPN
func (n Network) GenesisValidators(ctx context.Context, launchID uint64) (genVals []networktypes.GenesisValidator, err error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching genesis validators"))
	res, err := n.launchQuery.
		GenesisValidatorAll(ctx,
			&launchtypes.QueryAllGenesisValidatorRequest{
				LaunchID: launchID,
			},
		)
	if err != nil {
		return genVals, err
	}

	for _, acc := range res.GenesisValidator {
		genVals = append(genVals, networktypes.ToGenesisValidator(acc))
	}

	return genVals, nil
}

// MainnetAccount returns the campaign mainnet account for a launch from SPN
func (n Network) MainnetAccount(
	ctx context.Context,
	campaignID uint64,
	address string,
) (acc networktypes.MainnetAccount, err error) {
	n.ev.Send(events.New(events.StatusOngoing,
		fmt.Sprintf("Fetching campaign %d mainnet account %s", campaignID, address)),
	)
	res, err := n.campaignQuery.
		MainnetAccount(ctx,
			&campaigntypes.QueryGetMainnetAccountRequest{
				CampaignID: campaignID,
				Address:    address,
			},
		)
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return networktypes.MainnetAccount{}, ErrObjectNotFound
	} else if err != nil {
		return acc, err
	}

	return networktypes.ToMainnetAccount(res.MainnetAccount), nil
}

// MainnetAccounts returns the list of campaign mainnet accounts for a launch from SPN
func (n Network) MainnetAccounts(ctx context.Context, campaignID uint64) (genAccs []networktypes.MainnetAccount, err error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching campaign mainnet accounts"))
	res, err := n.campaignQuery.
		MainnetAccountAll(ctx,
			&campaigntypes.QueryAllMainnetAccountRequest{
				CampaignID: campaignID,
			},
		)
	if err != nil {
		return genAccs, err
	}

	for _, acc := range res.MainnetAccount {
		genAccs = append(genAccs, networktypes.ToMainnetAccount(acc))
	}

	return genAccs, nil
}

func (n Network) GenesisAccount(ctx context.Context, launchID uint64, address string) (networktypes.GenesisAccount, error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching genesis accounts"))
	res, err := n.launchQuery.GenesisAccount(ctx, &launchtypes.QueryGetGenesisAccountRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return networktypes.GenesisAccount{}, ErrObjectNotFound
	} else if err != nil {
		return networktypes.GenesisAccount{}, err
	}
	return networktypes.ToGenesisAccount(res.GenesisAccount), nil
}

func (n Network) VestingAccount(ctx context.Context, launchID uint64, address string) (networktypes.VestingAccount, error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching vesting accounts"))
	res, err := n.launchQuery.VestingAccount(ctx, &launchtypes.QueryGetVestingAccountRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return networktypes.VestingAccount{}, ErrObjectNotFound
	} else if err != nil {
		return networktypes.VestingAccount{}, err
	}
	return networktypes.ToVestingAccount(res.VestingAccount)
}

func (n Network) GenesisValidator(ctx context.Context, launchID uint64, address string) (networktypes.GenesisValidator, error) {
	n.ev.Send(events.New(events.StatusOngoing, "Fetching genesis validator"))
	res, err := n.launchQuery.GenesisValidator(ctx, &launchtypes.QueryGetGenesisValidatorRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return networktypes.GenesisValidator{}, ErrObjectNotFound
	} else if err != nil {
		return networktypes.GenesisValidator{}, err
	}
	return networktypes.ToGenesisValidator(res.GenesisValidator), nil
}

// ChainReward fetches the chain reward from SPN by launch id
func (n Network) ChainReward(ctx context.Context, launchID uint64) (rewardtypes.RewardPool, error) {
	res, err := n.rewardQuery.
		RewardPool(ctx,
			&rewardtypes.QueryGetRewardPoolRequest{
				LaunchID: launchID,
			},
		)

	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return rewardtypes.RewardPool{}, ErrObjectNotFound
	} else if err != nil {
		return rewardtypes.RewardPool{}, err
	}
	return res.RewardPool, nil
}

// ChainID fetches the network chain id
func (n Network) ChainID(ctx context.Context) (string, error) {
	status, err := n.cosmos.Status(ctx)
	if err != nil {
		return "", err
	}
	return status.NodeInfo.Network, nil
}
