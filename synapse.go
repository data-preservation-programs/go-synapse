package synapse

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/costs"
	"github.com/data-preservation-programs/go-synapse/pdp"
	"github.com/data-preservation-programs/go-synapse/storage"
	"github.com/data-preservation-programs/go-synapse/warmstorage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Options struct {
	PrivateKey *ecdsa.PrivateKey

	RPCURL string

	WarmStorageAddress common.Address

	ProviderURL string

	DataSetID int
}

type Client struct {
	network            Network
	chainID            int64
	ethClient          *ethclient.Client
	privateKey         *ecdsa.PrivateKey
	address            common.Address
	warmStorageAddress common.Address
	storageManager     *storage.Manager
	costsService       *costs.Service
	providerURL        string
	dataSetID          int
}

func New(ctx context.Context, opts Options) (*Client, error) {
	if opts.PrivateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}
	if opts.RPCURL == "" {
		return nil, fmt.Errorf("RPC URL is required")
	}

	ethClient, err := ethclient.DialContext(ctx, opts.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	network, chainID, err := DetectNetwork(ctx, ethClient)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("failed to detect network: %w", err)
	}

	warmStorageAddr := opts.WarmStorageAddress
	if warmStorageAddr == (common.Address{}) {
		warmStorageAddr = WarmStorageAddresses[network]
	}

	// for networks without static addresses (e.g. devnet), resolve from FWSS at runtime
	if _, ok := constants.PDPVerifierAddresses[constants.Network(network)]; !ok {
		if warmStorageAddr == (common.Address{}) {
			ethClient.Close()
			return nil, fmt.Errorf("network %s has no built-in addresses; set WarmStorageAddress (FWSS) to resolve at runtime", network)
		}
		addrs, err := constants.ResolveFromFWSS(ctx, ethClient, warmStorageAddr)
		if err != nil {
			ethClient.Close()
			return nil, fmt.Errorf("failed to resolve addresses from FWSS on %s: %w", network, err)
		}
		constants.RegisterNetwork(constants.Network(network), addrs)
	}

	address := crypto.PubkeyToAddress(opts.PrivateKey.PublicKey)

	client := &Client{
		network:            network,
		chainID:            chainID,
		ethClient:          ethClient,
		privateKey:         opts.PrivateKey,
		address:            address,
		warmStorageAddress: warmStorageAddr,
		providerURL:        opts.ProviderURL,
		dataSetID:          opts.DataSetID,
	}

	return client, nil
}

func (c *Client) Network() Network {
	return c.network
}

func (c *Client) ChainID() int64 {
	return c.chainID
}

func (c *Client) Address() common.Address {
	return c.address
}

func (c *Client) WarmStorageAddress() common.Address {
	return c.warmStorageAddress
}

func (c *Client) EthClient() *ethclient.Client {
	return c.ethClient
}

func (c *Client) Storage() (*storage.Manager, error) {
	if c.storageManager != nil {
		return c.storageManager, nil
	}

	if c.providerURL == "" {
		return nil, fmt.Errorf("provider URL is required for storage operations")
	}

	authHelper := pdp.NewAuthHelperFromKey(c.privateKey, c.warmStorageAddress, big.NewInt(c.chainID))
	pdpServer := pdp.NewServer(c.providerURL)

	var opts []storage.ManagerOption
	if c.dataSetID != 0 {
		stateViewAddr := constants.WarmStorageStateViewAddresses[constants.Network(c.network)]
		stateView, err := warmstorage.NewStateViewContract(stateViewAddr, c.ethClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create state view contract: %w", err)
		}
		opts = append(opts, storage.WithDataSetInfoFetcher(stateView))
	}

	c.storageManager = storage.NewManager(
		c.address,
		c.warmStorageAddress,
		authHelper,
		pdpServer,
		c.dataSetID,
		opts...,
	)

	return c.storageManager, nil
}

// Costs returns a lazily-initialized costs service for computing storage
// costs and deposit requirements.
func (c *Client) Costs() (*costs.Service, error) {
	if c.costsService != nil {
		return c.costsService, nil
	}

	config, err := costs.NetworkConfig(constants.Network(c.network))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve costs config: %w", err)
	}

	svc, err := costs.NewService(c.ethClient, c.chainID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create costs service: %w", err)
	}

	c.costsService = svc
	return c.costsService, nil
}

// GetUploadCosts is a convenience wrapper that computes the cost summary for
// uploading data using the caller's address as payer.
func (c *Client) GetUploadCosts(
	ctx context.Context,
	dataSetSizeBytes *big.Int,
	uploadSizeBytes *big.Int,
	opts *costs.UploadCostOptions,
) (*costs.UploadCosts, error) {
	svc, err := c.Costs()
	if err != nil {
		return nil, err
	}
	return svc.GetUploadCosts(ctx, c.address, dataSetSizeBytes, uploadSizeBytes, opts)
}

func (c *Client) Close() {
	if c.ethClient != nil {
		c.ethClient.Close()
	}
}

func (c *Client) NewAuthHelper() *pdp.AuthHelper {
	return pdp.NewAuthHelperFromKey(c.privateKey, c.warmStorageAddress, big.NewInt(c.chainID))
}

func (c *Client) NewPDPServer(providerURL string) *pdp.Server {
	return pdp.NewServer(providerURL)
}
