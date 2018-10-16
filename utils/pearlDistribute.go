// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package oyster_utils

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// PearlDistributeOysterbyABI is the input ABI used to generate the binding from.
const PearlDistributeOysterbyABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"multi\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"complete\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"director\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pearlContract\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"price\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_send\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newPrice\",\"type\":\"uint256\"}],\"name\":\"calculate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"calcMode\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"calcAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"distribute\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"pearlSend\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// PearlDistributeOysterbyBin is the compiled bytecode used for deploying new contracts.
const PearlDistributeOysterbyBin = `0x608060405260048054600160a060020a031990811673317ce293f9c6a7070bf474917fe6f37e21fbcbbe179182905560058054909116600160a060020a039290921691909117905534801561005357600080fd5b5060006002819055808055670de0b6b3a764000060019081556003805460ff191690911761010060b060020a0319166201000033600160a060020a0316021790556106039081906100a490396000f3006080604052600436106100ae5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631b8f5d5081146100b3578063522e1177146100da5780635af82abf1461010357806371c6d4dc14610134578063a035b1fe14610149578063adc9772e1461015e578063b9d92de814610184578063d402be571461019c578063dfcdde5e146101b1578063e4fc6b6d146101c6578063e75f1634146101db575b600080fd5b3480156100bf57600080fd5b506100c86101fc565b60408051918252519081900360200190f35b3480156100e657600080fd5b506100ef610202565b604080519115158252519081900360200190f35b34801561010f57600080fd5b50610118610210565b60408051600160a060020a039092168252519081900360200190f35b34801561014057600080fd5b50610118610225565b34801561015557600080fd5b506100c8610234565b34801561016a57600080fd5b50610182600160a060020a036004351660243561023a565b005b34801561019057600080fd5b506101826004356102c4565b3480156101a857600080fd5b506100ef610327565b3480156101bd57600080fd5b506100c8610330565b3480156101d257600080fd5b50610182610336565b3480156101e757600080fd5b506100c8600160a060020a0360043516610451565b60015481565b600354610100900460ff1681565b600354620100009004600160a060020a031681565b600454600160a060020a031681565b60005481565b600554604080517fa9059cbb000000000000000000000000000000000000000000000000000000008152600160a060020a038581166004830152602482018590529151919092169163a9059cbb91604480830192600092919082900301818387803b1580156102a857600080fd5b505af11580156102bc573d6000803e3d6000fd5b505050505050565b60035433600160a060020a039081166201000090920416146102e557600080fd5b600354610100900460ff16156102fa57600080fd5b6000811161030757600080fd5b60008181556003805460ff19166001179055600255610324610463565b50565b60035460ff1681565b60025481565b60035433600160a060020a0390811662010000909204161461035757600080fd5b600354610100900460ff161561036c57600080fd5b60035460ff16151561037d57600080fd5b60025460001061038c57600080fd5b600554604080517f70a08231000000000000000000000000000000000000000000000000000000008152600160a060020a033081166004830152915191909216916370a082319160248083019260209291908290030181600087803b1580156103f457600080fd5b505af1158015610408573d6000803e3d6000fd5b505050506040513d602081101561041e57600080fd5b5051600254111561042e57600080fd5b6003805460ff19169055610440610463565b6003805461ff001916610100179055565b60066020526000908152604090205481565b610483731be77862769ab791c4f95f8a2cbd0d3e07a3fd1f6103846104c5565b6104a37374ad69b41e71e311304564611434ddd59ee5d1f86103846104c5565b6104c37373da066d94fc41f11c2672ed9ecd39127da309766103846104c5565b565b60035460009060ff16151560011415610526576000546001548302611388028115156104ed57fe5b0490506104fc816001546105bb565b6002805482019055600160a060020a038416600090815260066020526040902081905590506105b6565b600554600160a060020a038481166000818152600660205260408082205481517fa9059cbb0000000000000000000000000000000000000000000000000000000081526004810194909452602484015251929093169263a9059cbb9260448084019382900301818387803b15801561059d57600080fd5b505af11580156105b1573d6000803e3d6000fd5b505050505b505050565b600081826001848601038115156105ce57fe5b040293925050505600a165627a7a723058201747c2a96ff4991a9e48f137a8f4ec4bb843f47239c8873f1835aa758970154c0029`

// DeployPearlDistributeOysterby deploys a new Ethereum contract, binding an instance of PearlDistributeOysterby to it.
func DeployPearlDistributeOysterby(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PearlDistributeOysterby, error) {
	parsed, err := abi.JSON(strings.NewReader(PearlDistributeOysterbyABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PearlDistributeOysterbyBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PearlDistributeOysterby{PearlDistributeOysterbyCaller: PearlDistributeOysterbyCaller{contract: contract}, PearlDistributeOysterbyTransactor: PearlDistributeOysterbyTransactor{contract: contract}, PearlDistributeOysterbyFilterer: PearlDistributeOysterbyFilterer{contract: contract}}, nil
}

// PearlDistributeOysterby is an auto generated Go binding around an Ethereum contract.
type PearlDistributeOysterby struct {
	PearlDistributeOysterbyCaller     // Read-only binding to the contract
	PearlDistributeOysterbyTransactor // Write-only binding to the contract
	PearlDistributeOysterbyFilterer   // Log filterer for contract events
}

// PearlDistributeOysterbyCaller is an auto generated read-only Go binding around an Ethereum contract.
type PearlDistributeOysterbyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PearlDistributeOysterbyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PearlDistributeOysterbyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PearlDistributeOysterbyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PearlDistributeOysterbyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PearlDistributeOysterbySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PearlDistributeOysterbySession struct {
	Contract     *PearlDistributeOysterby // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// PearlDistributeOysterbyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PearlDistributeOysterbyCallerSession struct {
	Contract *PearlDistributeOysterbyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// PearlDistributeOysterbyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PearlDistributeOysterbyTransactorSession struct {
	Contract     *PearlDistributeOysterbyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// PearlDistributeOysterbyRaw is an auto generated low-level Go binding around an Ethereum contract.
type PearlDistributeOysterbyRaw struct {
	Contract *PearlDistributeOysterby // Generic contract binding to access the raw methods on
}

// PearlDistributeOysterbyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PearlDistributeOysterbyCallerRaw struct {
	Contract *PearlDistributeOysterbyCaller // Generic read-only contract binding to access the raw methods on
}

// PearlDistributeOysterbyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PearlDistributeOysterbyTransactorRaw struct {
	Contract *PearlDistributeOysterbyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPearlDistributeOysterby creates a new instance of PearlDistributeOysterby, bound to a specific deployed contract.
func NewPearlDistributeOysterby(address common.Address, backend bind.ContractBackend) (*PearlDistributeOysterby, error) {
	contract, err := bindPearlDistributeOysterby(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PearlDistributeOysterby{PearlDistributeOysterbyCaller: PearlDistributeOysterbyCaller{contract: contract}, PearlDistributeOysterbyTransactor: PearlDistributeOysterbyTransactor{contract: contract}, PearlDistributeOysterbyFilterer: PearlDistributeOysterbyFilterer{contract: contract}}, nil
}

// NewPearlDistributeOysterbyCaller creates a new read-only instance of PearlDistributeOysterby, bound to a specific deployed contract.
func NewPearlDistributeOysterbyCaller(address common.Address, caller bind.ContractCaller) (*PearlDistributeOysterbyCaller, error) {
	contract, err := bindPearlDistributeOysterby(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PearlDistributeOysterbyCaller{contract: contract}, nil
}

// NewPearlDistributeOysterbyTransactor creates a new write-only instance of PearlDistributeOysterby, bound to a specific deployed contract.
func NewPearlDistributeOysterbyTransactor(address common.Address, transactor bind.ContractTransactor) (*PearlDistributeOysterbyTransactor, error) {
	contract, err := bindPearlDistributeOysterby(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PearlDistributeOysterbyTransactor{contract: contract}, nil
}

// NewPearlDistributeOysterbyFilterer creates a new log filterer instance of PearlDistributeOysterby, bound to a specific deployed contract.
func NewPearlDistributeOysterbyFilterer(address common.Address, filterer bind.ContractFilterer) (*PearlDistributeOysterbyFilterer, error) {
	contract, err := bindPearlDistributeOysterby(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PearlDistributeOysterbyFilterer{contract: contract}, nil
}

// bindPearlDistributeOysterby binds a generic wrapper to an already deployed contract.
func bindPearlDistributeOysterby(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PearlDistributeOysterbyABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PearlDistributeOysterby *PearlDistributeOysterbyRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PearlDistributeOysterby.Contract.PearlDistributeOysterbyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PearlDistributeOysterby *PearlDistributeOysterbyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.PearlDistributeOysterbyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PearlDistributeOysterby *PearlDistributeOysterbyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.PearlDistributeOysterbyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PearlDistributeOysterby.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.contract.Transact(opts, method, params...)
}

// CalcAmount is a free data retrieval call binding the contract method 0xdfcdde5e.
//
// Solidity: function calcAmount() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) CalcAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "calcAmount")
	return *ret0, err
}

// CalcAmount is a free data retrieval call binding the contract method 0xdfcdde5e.
//
// Solidity: function calcAmount() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) CalcAmount() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.CalcAmount(&_PearlDistributeOysterby.CallOpts)
}

// CalcAmount is a free data retrieval call binding the contract method 0xdfcdde5e.
//
// Solidity: function calcAmount() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) CalcAmount() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.CalcAmount(&_PearlDistributeOysterby.CallOpts)
}

// CalcMode is a free data retrieval call binding the contract method 0xd402be57.
//
// Solidity: function calcMode() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) CalcMode(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "calcMode")
	return *ret0, err
}

// CalcMode is a free data retrieval call binding the contract method 0xd402be57.
//
// Solidity: function calcMode() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) CalcMode() (bool, error) {
	return _PearlDistributeOysterby.Contract.CalcMode(&_PearlDistributeOysterby.CallOpts)
}

// CalcMode is a free data retrieval call binding the contract method 0xd402be57.
//
// Solidity: function calcMode() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) CalcMode() (bool, error) {
	return _PearlDistributeOysterby.Contract.CalcMode(&_PearlDistributeOysterby.CallOpts)
}

// Complete is a free data retrieval call binding the contract method 0x522e1177.
//
// Solidity: function complete() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) Complete(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "complete")
	return *ret0, err
}

// Complete is a free data retrieval call binding the contract method 0x522e1177.
//
// Solidity: function complete() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Complete() (bool, error) {
	return _PearlDistributeOysterby.Contract.Complete(&_PearlDistributeOysterby.CallOpts)
}

// Complete is a free data retrieval call binding the contract method 0x522e1177.
//
// Solidity: function complete() constant returns(bool)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) Complete() (bool, error) {
	return _PearlDistributeOysterby.Contract.Complete(&_PearlDistributeOysterby.CallOpts)
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) Director(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "director")
	return *ret0, err
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Director() (common.Address, error) {
	return _PearlDistributeOysterby.Contract.Director(&_PearlDistributeOysterby.CallOpts)
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) Director() (common.Address, error) {
	return _PearlDistributeOysterby.Contract.Director(&_PearlDistributeOysterby.CallOpts)
}

// Multi is a free data retrieval call binding the contract method 0x1b8f5d50.
//
// Solidity: function multi() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) Multi(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "multi")
	return *ret0, err
}

// Multi is a free data retrieval call binding the contract method 0x1b8f5d50.
//
// Solidity: function multi() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Multi() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.Multi(&_PearlDistributeOysterby.CallOpts)
}

// Multi is a free data retrieval call binding the contract method 0x1b8f5d50.
//
// Solidity: function multi() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) Multi() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.Multi(&_PearlDistributeOysterby.CallOpts)
}

// PearlContract is a free data retrieval call binding the contract method 0x71c6d4dc.
//
// Solidity: function pearlContract() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) PearlContract(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "pearlContract")
	return *ret0, err
}

// PearlContract is a free data retrieval call binding the contract method 0x71c6d4dc.
//
// Solidity: function pearlContract() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) PearlContract() (common.Address, error) {
	return _PearlDistributeOysterby.Contract.PearlContract(&_PearlDistributeOysterby.CallOpts)
}

// PearlContract is a free data retrieval call binding the contract method 0x71c6d4dc.
//
// Solidity: function pearlContract() constant returns(address)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) PearlContract() (common.Address, error) {
	return _PearlDistributeOysterby.Contract.PearlContract(&_PearlDistributeOysterby.CallOpts)
}

// PearlSend is a free data retrieval call binding the contract method 0xe75f1634.
//
// Solidity: function pearlSend( address) constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) PearlSend(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "pearlSend", arg0)
	return *ret0, err
}

// PearlSend is a free data retrieval call binding the contract method 0xe75f1634.
//
// Solidity: function pearlSend( address) constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) PearlSend(arg0 common.Address) (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.PearlSend(&_PearlDistributeOysterby.CallOpts, arg0)
}

// PearlSend is a free data retrieval call binding the contract method 0xe75f1634.
//
// Solidity: function pearlSend( address) constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) PearlSend(arg0 common.Address) (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.PearlSend(&_PearlDistributeOysterby.CallOpts, arg0)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCaller) Price(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PearlDistributeOysterby.contract.Call(opts, out, "price")
	return *ret0, err
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Price() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.Price(&_PearlDistributeOysterby.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() constant returns(uint256)
func (_PearlDistributeOysterby *PearlDistributeOysterbyCallerSession) Price() (*big.Int, error) {
	return _PearlDistributeOysterby.Contract.Price(&_PearlDistributeOysterby.CallOpts)
}

// Calculate is a paid mutator transaction binding the contract method 0xb9d92de8.
//
// Solidity: function calculate(newPrice uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactor) Calculate(opts *bind.TransactOpts, newPrice *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.contract.Transact(opts, "calculate", newPrice)
}

// Calculate is a paid mutator transaction binding the contract method 0xb9d92de8.
//
// Solidity: function calculate(newPrice uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Calculate(newPrice *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Calculate(&_PearlDistributeOysterby.TransactOpts, newPrice)
}

// Calculate is a paid mutator transaction binding the contract method 0xb9d92de8.
//
// Solidity: function calculate(newPrice uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactorSession) Calculate(newPrice *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Calculate(&_PearlDistributeOysterby.TransactOpts, newPrice)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactor) Distribute(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PearlDistributeOysterby.contract.Transact(opts, "distribute")
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Distribute() (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Distribute(&_PearlDistributeOysterby.TransactOpts)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactorSession) Distribute() (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Distribute(&_PearlDistributeOysterby.TransactOpts)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(_send address, _amount uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactor) Stake(opts *bind.TransactOpts, _send common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.contract.Transact(opts, "stake", _send, _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(_send address, _amount uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbySession) Stake(_send common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Stake(&_PearlDistributeOysterby.TransactOpts, _send, _amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(_send address, _amount uint256) returns()
func (_PearlDistributeOysterby *PearlDistributeOysterbyTransactorSession) Stake(_send common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _PearlDistributeOysterby.Contract.Stake(&_PearlDistributeOysterby.TransactOpts, _send, _amount)
}
