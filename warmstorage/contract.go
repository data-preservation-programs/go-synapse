package warmstorage

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const StateViewABIJSON = `[
	{
		"type": "function",
		"name": "getDataSet",
		"inputs": [{"name": "dataSetId", "type": "uint256"}],
		"outputs": [
			{
				"name": "info",
				"type": "tuple",
				"components": [
					{"name": "pdpRailId", "type": "uint256"},
					{"name": "cacheMissRailId", "type": "uint256"},
					{"name": "cdnRailId", "type": "uint256"},
					{"name": "payer", "type": "address"},
					{"name": "payee", "type": "address"},
					{"name": "serviceProvider", "type": "address"},
					{"name": "commissionBps", "type": "uint256"},
					{"name": "clientDataSetId", "type": "uint256"},
					{"name": "pdpEndEpoch", "type": "uint256"},
					{"name": "providerId", "type": "uint256"},
					{"name": "dataSetId", "type": "uint256"}
				]
			}
		],
		"stateMutability": "view"
	}
]`

type StateViewContract struct {
	address common.Address
	abi     abi.ABI
	client  *ethclient.Client
}

func NewStateViewContract(address common.Address, client *ethclient.Client) (*StateViewContract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(StateViewABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse StateView ABI: %w", err)
	}

	return &StateViewContract{
		address: address,
		abi:     parsedABI,
		client:  client,
	}, nil
}

func (c *StateViewContract) GetDataSet(ctx context.Context, dataSetID int) (*DataSetInfo, error) {
	data, err := c.abi.Pack("getDataSet", big.NewInt(int64(dataSetID)))
	if err != nil {
		return nil, fmt.Errorf("failed to pack getDataSet call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getDataSet: %w", err)
	}

	values, err := c.abi.Unpack("getDataSet", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getDataSet result: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("empty result from getDataSet")
	}

	infoStruct, ok := values[0].(struct {
		PdpRailId       *big.Int       `abi:"pdpRailId"`
		CacheMissRailId *big.Int       `abi:"cacheMissRailId"`
		CdnRailId       *big.Int       `abi:"cdnRailId"`
		Payer           common.Address `abi:"payer"`
		Payee           common.Address `abi:"payee"`
		ServiceProvider common.Address `abi:"serviceProvider"`
		CommissionBps   *big.Int       `abi:"commissionBps"`
		ClientDataSetId *big.Int       `abi:"clientDataSetId"`
		PdpEndEpoch     *big.Int       `abi:"pdpEndEpoch"`
		ProviderId      *big.Int       `abi:"providerId"`
		DataSetId       *big.Int       `abi:"dataSetId"`
	})
	if !ok {
		return nil, fmt.Errorf("unexpected type for getDataSet result: %T", values[0])
	}

	if infoStruct.PdpRailId.Sign() == 0 {
		return nil, fmt.Errorf("data set %d does not exist", dataSetID)
	}

	return &DataSetInfo{
		PDPRailID:       infoStruct.PdpRailId,
		CacheMissRailID: infoStruct.CacheMissRailId,
		CDNRailID:       infoStruct.CdnRailId,
		Payer:           infoStruct.Payer,
		Payee:           infoStruct.Payee,
		ServiceProvider: infoStruct.ServiceProvider,
		CommissionBps:   infoStruct.CommissionBps,
		ClientDataSetID: infoStruct.ClientDataSetId,
		PDPEndEpoch:     infoStruct.PdpEndEpoch,
		ProviderID:      infoStruct.ProviderId,
		DataSetID:       infoStruct.DataSetId,
	}, nil
}
