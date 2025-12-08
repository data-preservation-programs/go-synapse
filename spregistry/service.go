package spregistry

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Service struct {
	client     *ethclient.Client
	contract   *Contract
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int
}

func NewService(client *ethclient.Client, registryAddress common.Address, privateKey *ecdsa.PrivateKey, chainID *big.Int) (*Service, error) {
	contract, err := NewContract(registryAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	var address common.Address
	if privateKey != nil {
		address = crypto.PubkeyToAddress(privateKey.PublicKey)
	}

	return &Service{
		client:     client,
		contract:   contract,
		privateKey: privateKey,
		address:    address,
		chainID:    chainID,
	}, nil
}


func (s *Service) GetProvider(ctx context.Context, providerID int) (*ProviderInfo, error) {
	result, err := s.contract.GetProviderWithProduct(ctx, big.NewInt(int64(providerID)), uint8(ProductTypePDP))
	if err != nil {
		return nil, err
	}

	if result.ProviderInfo.ServiceProvider == (common.Address{}) {
		return nil, nil
	}

	return s.convertToProviderInfo(providerID, result), nil
}

func (s *Service) GetProviderByAddress(ctx context.Context, addr common.Address) (*ProviderInfo, error) {
	result, err := s.contract.GetProviderByAddress(ctx, addr)
	if err != nil {
		return nil, err
	}

	if result.Info.ServiceProvider == (common.Address{}) {
		return nil, nil
	}

	return s.GetProvider(ctx, int(result.ProviderID.Int64()))
}

func (s *Service) GetProviderIDByAddress(ctx context.Context, addr common.Address) (int, error) {
	id, err := s.contract.GetProviderIDByAddress(ctx, addr)
	if err != nil {
		return 0, err
	}
	return int(id.Int64()), nil
}

func (s *Service) GetAllActiveProviders(ctx context.Context) ([]*ProviderInfo, error) {
	var allProviders []*ProviderInfo
	pageSize := big.NewInt(50)
	offset := big.NewInt(0)

	for {
		providerIDs, hasMore, err := s.contract.GetAllActiveProviders(ctx, offset, pageSize)
		if err != nil {
			return nil, err
		}

		if len(providerIDs) > 0 {
			for _, id := range providerIDs {
				provider, err := s.GetProvider(ctx, int(id.Int64()))
				if err != nil {
					continue
				}
				if provider != nil {
					allProviders = append(allProviders, provider)
				}
			}
		}

		if !hasMore {
			break
		}
		offset = new(big.Int).Add(offset, pageSize)
	}

	return allProviders, nil
}

func (s *Service) GetProviders(ctx context.Context, providerIDs []int) ([]*ProviderInfo, error) {
	if len(providerIDs) == 0 {
		return nil, nil
	}

	providers := make([]*ProviderInfo, 0, len(providerIDs))
	for _, id := range providerIDs {
		provider, err := s.GetProvider(ctx, id)
		if err != nil {
			continue
		}
		if provider != nil {
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

func (s *Service) IsProviderActive(ctx context.Context, providerID int) (bool, error) {
	return s.contract.IsProviderActive(ctx, big.NewInt(int64(providerID)))
}

func (s *Service) IsRegisteredProvider(ctx context.Context, addr common.Address) (bool, error) {
	return s.contract.IsRegisteredProvider(ctx, addr)
}

func (s *Service) GetProviderCount(ctx context.Context) (int, error) {
	count, err := s.contract.GetProviderCount(ctx)
	if err != nil {
		return 0, err
	}
	return int(count.Int64()), nil
}

func (s *Service) ActiveProviderCount(ctx context.Context) (int, error) {
	count, err := s.contract.ActiveProviderCount(ctx)
	if err != nil {
		return 0, err
	}
	return int(count.Int64()), nil
}

func (s *Service) GetPDPService(ctx context.Context, providerID int) (*PDPServiceInfo, error) {
	result, err := s.contract.GetProviderWithProduct(ctx, big.NewInt(int64(providerID)), uint8(ProductTypePDP))
	if err != nil {
		return nil, err
	}

	if !result.Product.IsActive {
		return nil, nil
	}

	capabilities := CapabilitiesListToMap(result.Product.CapabilityKeys, result.ProductCapabilityValues)

	return &PDPServiceInfo{
		Offering:     *DecodePDPCapabilities(capabilities),
		Capabilities: capabilities,
		IsActive:     result.Product.IsActive,
	}, nil
}

func (s *Service) ProviderHasProduct(ctx context.Context, providerID int, productType ProductType) (bool, error) {
	return s.contract.ProviderHasProduct(ctx, big.NewInt(int64(providerID)), uint8(productType))
}


func (s *Service) RegisterProvider(ctx context.Context, info ProviderRegistrationInfo) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	fee, err := s.contract.RegistrationFee(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get registration fee: %w", err)
	}

	capabilityKeys, capabilityValues, err := EncodePDPCapabilities(&info.PDPOffering, info.Capabilities)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to encode capabilities: %w", err)
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	opts.Value = fee

	tx, err := s.contract.RegisterProvider(opts, info.Payee, info.Name, info.Description, uint8(ProductTypePDP), capabilityKeys, capabilityValues)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to register provider: %w", err)
	}

	return tx.Hash(), nil
}

func (s *Service) UpdateProviderInfo(ctx context.Context, name, description string) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.contract.UpdateProviderInfo(opts, name, description)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to update provider info: %w", err)
	}

	return tx.Hash(), nil
}

func (s *Service) RemoveProvider(ctx context.Context) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.contract.RemoveProvider(opts)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to remove provider: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) AddPDPProduct(ctx context.Context, offering PDPOffering, capabilities map[string]string) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	capabilityKeys, capabilityValues, err := EncodePDPCapabilities(&offering, capabilities)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to encode capabilities: %w", err)
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.contract.AddProduct(opts, uint8(ProductTypePDP), capabilityKeys, capabilityValues)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to add PDP product: %w", err)
	}

	return tx.Hash(), nil
}

func (s *Service) UpdatePDPProduct(ctx context.Context, offering PDPOffering, capabilities map[string]string) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	capabilityKeys, capabilityValues, err := EncodePDPCapabilities(&offering, capabilities)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to encode capabilities: %w", err)
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.contract.UpdateProduct(opts, uint8(ProductTypePDP), capabilityKeys, capabilityValues)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to update PDP product: %w", err)
	}

	return tx.Hash(), nil
}

func (s *Service) RemoveProduct(ctx context.Context, productType ProductType) (common.Hash, error) {
	if s.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key required for write operations")
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.contract.RemoveProduct(opts, uint8(productType))
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to remove product: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) convertToProviderInfo(providerID int, result *GetProviderWithProductResult) *ProviderInfo {
	products := make(map[string]*ServiceProduct)

	if result.Product.IsActive {
		capabilities := CapabilitiesListToMap(result.Product.CapabilityKeys, result.ProductCapabilityValues)
		products["PDP"] = &ServiceProduct{
			Type:         "PDP",
			IsActive:     result.Product.IsActive,
			Capabilities: capabilities,
			Data:         DecodePDPCapabilities(capabilities),
		}
	}

	return &ProviderInfo{
		ID:              providerID,
		ServiceProvider: result.ProviderInfo.ServiceProvider,
		Payee:           result.ProviderInfo.Payee,
		Name:            result.ProviderInfo.Name,
		Description:     result.ProviderInfo.Description,
		Active:          result.ProviderInfo.IsActive,
		Products:        products,
	}
}

func (s *Service) transactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}
