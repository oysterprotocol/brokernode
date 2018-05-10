package services_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"math/big"
	abi2 "github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/oysterprotocol/brokernode/models"
	"io/ioutil"
	"time"
)

const (
	localNetworkUrl = "http://127.0.0.1:7545"
	oysterbyNetworkUrl = "http://54.197.3.171:8080"
    oysterPearlABI = "[ { \"constant\": true, \"inputs\": [], \"name\": \"name\", \"outputs\": [ { \"name\": \"\", \"type\": \"string\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_spender\", \"type\": \"address\" }, { \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"approve\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [], \"name\": \"openSale\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"totalSupply\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_payout\", \"type\": \"address\" }, { \"name\": \"_fee\", \"type\": \"address\" } ], \"name\": \"claim\", \"outputs\": [ {\"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"retentionMax\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_from\", \"type\": \"address\" }, { \"name\": \"_to\", \"type\": \"address\" }, { \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"transferFrom\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [], \"name\": \"withdrawFunds\", \"outputs\": [], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [ { \"name\": \"\", \"type\": \"address\" } ], \"name\": \"balances\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"decimals\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint8\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [ { \"name\": \"\", \"type\": \"address\" } ], \"name\": \"buried\", \"outputs\": [ { \"name\": \"\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"burn\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayabl\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"epochSet\", \"type\": \"uint256\" } ], \"name\": \"amendEpoch\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"director\", \"outputs\": [ { \"name\": \"\", \"type\": \"address\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"retentionSet\", \"type\": \"uint8\" }, { \"name\": \"accuracy\", \"type\": \"uint8\" } ], \"name\": \"amendRetention\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [], \"name\": \"bury\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"feeAmount\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [ { \"name\": \"_owner\", \"type\": \"address\" } ], \"name\": \"balanceOf\", \"outputs\": [ { \"name\": \"balance\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_from\", \"type\": \"address\" }, { \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"burnFrom\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"claimAmountSet\", \"type\": \"uint8\" }, { \"name\": \"payAmountSet\", \"type\": \"uint8\" }, { \"name\": \"feeAmountSet\", \"type\": \"uint8\" }, { \"name\": \"accuracy\", \"type\": \"uint8\" } ], \"name\": \"amendClaim\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"claimAmount\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"epoch\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"symbol\", \"outputs\": [ { \"name\": \"\", \"type\": \"string\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_to\", \"type\": \"address\" }, { \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"transfer\", \"outputs\": [], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"saleClosed\", \"outputs\": [ { \"name\": \"\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"payAmount\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [ { \"name\": \"\", \"type\": \"address\" } ], \"name\": \"claimed\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"funds\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"_spender\", \"type\": \"address\" }, { \"name\": \"_value\", \"type\": \"uint256\" }, { \"name\": \"_extraData\", \"type\": \"bytes\" } ], \"name\": \"approveAndCall\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [], \"name\": \"selfLock\", \"outputs\": [], \"payable\": true, \"stateMutability\": \"payable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [ { \"name\": \"\", \"type\": \"address\" }, { \"name\": \"\", \"type\": \"address\" } ], \"name\": \"allowance\", \"outputs\": [ { \"name\": \"\", \"type\": \"uint256\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [ { \"name\": \"newDirector\", \"type\": \"address\" } ], \"name\": \"transferDirector\", \"outputs\": [], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": false, \"inputs\": [], \"name\": \"closeSale\", \"outputs\": [ { \"name\": \"success\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"nonpayable\", \"type\": \"function\" }, { \"constant\": true, \"inputs\": [], \"name\": \"directorLock\", \"outputs\": [ { \"name\": \"\", \"type\": \"bool\" } ], \"payable\": false, \"stateMutability\": \"view\", \"type\": \"function\" }, { \"inputs\": [], \"payable\": true, \"stateMutability\": \"payable\", \"type\": \"constructor\" }, { \"payable\": true, \"stateMutability\": \"payable\", \"type\": \"fallback\" }, { \"anonymous\": false, \"inputs\": [ { \"indexed\": true, \"name\": \"_from\", \"type\": \"address\" }, { \"indexed\": true, \"name\": \"_to\", \"type\": \"address\" }, { \"indexed\": false, \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"Transfer\", \"type\": \"event\" }, { \"anonymous\": false, \"inputs\": [ { \"indexed\": true, \"name\": \"_owner\", \"type\": \"address\" }, { \"indexed\": true, \"name\": \"_spender\", \"type\": \"address\" }, { \"indexed\": false, \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"Approval\", \"type\": \"event\" }, { \"anonymous\": false, \"inputs\": [ { \"indexed\": true, \"name\": \"_from\", \"type\": \"address\" }, { \"indexed\": false, \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"Burn\", \"type\": \"event\" }, { \"anonymous\": false, \"inputs\": [ { \"indexed\": true, \"name\": \"_target\", \"type\": \"address\" }, { \"indexed\": false, \"name\": \"_value\", \"type\": \"uint256\" } ], \"name\": \"Bury\", \"type\": \"event\" }, { \"anonymous\": false, \"inputs\": [ { \"indexed\": true, \"name\": \"_target\", \"type\": \"address\" }, { \"indexed\": true, \"name\": \"_payout\", \"type\": \"address\" }, { \"indexed\": true, \"name\": \"_fee\", \"type\": \"address\" } ], \"name\": \"Claim\", \"type\": \"event\" } ]"
    oysterPearlByteCode = "0x606060405233600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506040805190810160405280600c81526020017f4f797374657220506561726c00000000000000000000000000000000000000008152506000908051906020019062000092929190620002c1565b506040805190810160405280600381526020017f50524c000000000000000000000000000000000000000000000000000000000081525060019080519060200190620000e0929190620002c1565b506012600260006101000a81548160ff021916908360ff1602179055506001600560146101000a81548160ff0219169083151502179055506000600560156101000a81548160ff02191690831515021790555060006004819055506000600381905550600260009054906101000a900460ff1660ff16600a0a63017d784002600360008282540192505081905550600260009054906101000a900460ff1660ff16600a0a63047868c002600360008282540192505081905550600260009054906101000a900460ff1660ff16600a0a627a120002600360008282540192505081905550600354600b6000600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506001600260009054906101000a900460ff1660ff1603600a0a6005026006819055506001600260009054906101000a900460ff1660ff1603600a0a6004026007819055506001600260009054906101000a900460ff1660ff1603600a0a6001026008819055506301e13380600981905550600260009054906101000a900460ff1660ff16600a0a602802600a8190555062000370565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106200030457805160ff191683800117855562000335565b8280016001018555821562000335579182015b828111156200033457825182559160200191906001019062000317565b5b50905062000344919062000348565b5090565b6200036d91905b80821115620003695760008160009055506001016200034f565b5090565b90565b6128da80620003806000396000f3006060604052600436106101ac576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde03146102ea578063095ea7b314610378578063167ff46f146103d257806318160ddd146103ff57806321c0b3421461042857806322bb4f531461049857806323b872dd146104c157806324600fc31461053a57806327e235e31461054f578063313ce5671461059c5780633f1199e6146105cb57806342966c681461061c578063549215a3146106575780635af82abf146106925780635f5f2aef146106e757806361161aae1461073157806369e154041461075e57806370a082311461078757806379cc6790146107d45780637dbc9fba1461082e578063830953ab14610890578063900cf0cf146108b957806395d89b41146108e2578063a9059cbb14610970578063b8c766b8146109b2578063c8705544146109df578063c884ef8314610a08578063c89f2ce414610a55578063cae9ca5114610a7e578063d1e7e81f14610b1b578063dd62ed3e14610b25578063ddd41ef614610b91578063ee55efee14610bca578063ffe2d77e14610bf7575b6000600560149054906101000a900460ff161515156101ca57600080fd5b66038d7ea4c6800034101515156101e057600080fd5b61138834029050600260009054906101000a900460ff1660ff16600a0a631dcd65000281600354011115151561021557600080fd5b8060036000828254019250508190555080600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550346004600082825401925050819055503373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a350005b34156102f557600080fd5b6102fd610c24565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561033d578082015181840152602081019050610322565b50505050905090810190601f16801561036a5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b341561038357600080fd5b6103b8600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091908035906020019091905050610cc2565b604051808215151515815260200191505060405180910390f35b34156103dd57600080fd5b6103e5610e0d565b604051808215151515815260200191505060405180910390f35b341561040a57600080fd5b610412610ec4565b6040518082815260200191505060405180910390f35b341561043357600080fd5b61047e600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050610eca565b604051808215151515815260200191505060405180910390f35b34156104a357600080fd5b6104ab6114cd565b6040518082815260200191505060405180910390f35b34156104cc57600080fd5b610520600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190803573ffffffffffffffffffffffffffffffffffffffff169060200190919080359060200190919050506114d3565b604051808215151515815260200191505060405180910390f35b341561054557600080fd5b61054d611600565b005b341561055a57600080fd5b610586600480803573ffffffffffffffffffffffffffffffffffffffff169060200190919050506116d7565b6040518082815260200191505060405180910390f35b34156105a757600080fd5b6105af6116ef565b604051808260ff1660ff16815260200191505060405180910390f35b34156105d657600080fd5b610602600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050611702565b604051808215151515815260200191505060405180910390f35b341561062757600080fd5b61063d6004808035906020019091905050611722565b604051808215151515815260200191505060405180910390f35b341561066257600080fd5b610678600480803590602001909190505061187f565b604051808215151515815260200191505060405180910390f35b341561069d57600080fd5b6106a5611909565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156106f257600080fd5b610717600480803560ff1690602001909190803560ff1690602001909190505061192f565b604051808215151515815260200191505060405180910390f35b341561073c57600080fd5b6107446119d9565b604051808215151515815260200191505060405180910390f35b341561076957600080fd5b610771611c05565b6040518082815260200191505060405180910390f35b341561079257600080fd5b6107be600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050611c0b565b6040518082815260200191505060405180910390f35b34156107df57600080fd5b610814600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091908035906020019091905050611c54565b604051808215151515815260200191505060405180910390f35b341561083957600080fd5b610876600480803560ff1690602001909190803560ff1690602001909190803560ff1690602001909190803560ff16906020019091905050611ec7565b604051808215151515815260200191505060405180910390f35b341561089b57600080fd5b6108a3611fd5565b6040518082815260200191505060405180910390f35b34156108c457600080fd5b6108cc611fdb565b6040518082815260200191505060405180910390f35b34156108ed57600080fd5b6108f5611fe1565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561093557808201518184015260208101905061091a565b50505050905090810190601f1680156109625780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b341561097b57600080fd5b6109b0600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190803590602001909190505061207f565b005b34156109bd57600080fd5b6109c561208e565b604051808215151515815260200191505060405180910390f35b34156109ea57600080fd5b6109f26120a1565b6040518082815260200191505060405180910390f35b3415610a1357600080fd5b610a3f600480803573ffffffffffffffffffffffffffffffffffffffff169060200190919050506120a7565b6040518082815260200191505060405180910390f35b3415610a6057600080fd5b610a686120bf565b6040518082815260200191505060405180910390f35b3415610a8957600080fd5b610b01600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190803590602001909190803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919050506120c5565b604051808215151515815260200191505060405180910390f35b610b23612243565b005b3415610b3057600080fd5b610b7b600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050612309565b6040518082815260200191505060405180910390f35b3415610b9c57600080fd5b610bc8600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190505061232e565b005b3415610bd557600080fd5b610bdd6123ce565b604051808215151515815260200191505060405180910390f35b3415610c0257600080fd5b610c0a612486565b604051808215151515815260200191505060405180910390f35b60008054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610cba5780601f10610c8f57610100808354040283529160200191610cba565b820191906000526020600020905b815481529060010190602001808311610c9d57829003601f168201915b505050505081565b6000600d60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16151515610d1d57600080fd5b81600c60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040518082815260200191505060405180910390a36001905092915050565b6000600560159054906101000a900460ff16151515610e2b57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610e8757600080fd5b600560149054906101000a900460ff161515610ea257600080fd5b6000600560146101000a81548160ff0219169083151502179055506001905090565b60035481565b600080600d60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515610f2557600080fd5b8273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1614151515610f6057600080fd5b8373ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151515610f9b57600080fd5b8273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151515610fd657600080fd5b6001600e60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205414806110675750600954600e60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054420310155b151561107257600080fd5b600654600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054101515156110c257600080fd5b42600e60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550600b60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205401019050600654600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550600754600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550600854600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fcac3ed26c9dd72a2c44999857298af9c72ba2d1ca9784f5dad48c933e2224c1160405160405180910390a48373ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef6007546040518082815260200191505060405180910390a38273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef6008546040518082815260200191505060405180910390a380600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205401011415156114c257fe5b600191505092915050565b600a5481565b6000600c60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054821115151561156057600080fd5b81600c60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825403925050819055506115f5848484612499565b600190509392505050565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561165c57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f1935050505015156116d557600080fd5b565b600b6020528060005260406000206000915090505481565b600260009054906101000a900460ff1681565b600d6020528060005260406000206000915054906101000a900460ff1681565b6000600d60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff1615151561177d57600080fd5b81600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054101515156117cb57600080fd5b81600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550816003600082825403925050819055503373ffffffffffffffffffffffffffffffffffffffff167fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5836040518082815260200191505060405180910390a260019050919050565b6000600560159054906101000a900460ff1615151561189d57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415156118f957600080fd5b8160098190555060019050919050565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600560159054906101000a900460ff1615151561194d57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415156119a957600080fd5b8160ff16600260009054906101000a900460ff1660ff1603600a0a8360ff1602600a819055506001905092915050565b6000600d60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16151515611a3457600080fd5b600654600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410151515611a8457600080fd5b600a54600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205411151515611ad457600080fd5b6001600d60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506001600e60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055503373ffffffffffffffffffffffffffffffffffffffff167fc96e8fee6eb65975d592ca9a340f33200433df4c42b2f623dd9fc6d22984d495600b60003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020546040518082815260200191505060405180910390a26001905090565b60085481565b6000600b60008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000600d60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16151515611caf57600080fd5b81600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410151515611cfd57600080fd5b600c60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548211151515611d8857600080fd5b81600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254039250508190555081600c60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550816003600082825403925050819055508273ffffffffffffffffffffffffffffffffffffffff167fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5836040518082815260200191505060405180910390a26001905092915050565b6000600560159054906101000a900460ff16151515611ee557600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515611f4157600080fd5b82840160ff168560ff16141515611f5757600080fd5b8160ff16600260009054906101000a900460ff1660ff1603600a0a8560ff16026006819055508160ff16600260009054906101000a900460ff1660ff1603600a0a8460ff16026007819055508160ff16600260009054906101000a900460ff1660ff1603600a0a8360ff160260088190555060019050949350505050565b60065481565b60095481565b60018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156120775780601f1061204c57610100808354040283529160200191612077565b820191906000526020600020905b81548152906001019060200180831161205a57829003601f168201915b505050505081565b61208a338383612499565b5050565b600560149054906101000a900460ff1681565b60075481565b600e6020528060005260406000206000915090505481565b60045481565b6000808490506120d58585610cc2565b1561223a578073ffffffffffffffffffffffffffffffffffffffff16638f4ffcb1338630876040518563ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200180602001828103825283818151815260200191508051906020019080838360005b838110156121cf5780820151818401526020810190506121b4565b50505050905090810190601f1680156121fc5780820380516001836020036101000a031916815260200191505b5095505050505050600060405180830381600087803b151561221d57600080fd5b6102c65a03f1151561222e57600080fd5b5050506001915061223b565b5b509392505050565b600560159054906101000a900460ff1615151561225f57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415156122bb57600080fd5b600560149054906101000a900460ff1615156122d657600080fd5b678ac7230489e80000341415156122ec57600080fd5b6001600560156101000a81548160ff021916908315150217905550565b600c602052816000526040600020602052806000526040600020600091509150505481565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561238a57600080fd5b80600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600560159054906101000a900460ff161515156123ec57600080fd5b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561244857600080fd5b600560149054906101000a900460ff1615151561246457600080fd5b6001600560146101000a81548160ff0219169083151502179055506001905090565b600560159054906101000a900460ff1681565b6000600d60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515156124f457600080fd5b600d60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161561259957600a5482600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054011115151561259857600080fd5b5b60008373ffffffffffffffffffffffffffffffffffffffff16141515156125bf57600080fd5b81600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015151561260d57600080fd5b600b60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205482600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020540111151561269b57600080fd5b600b60008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205401905081600b60008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254039250508190555081600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a380600b60008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054600b60008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054011415156128a857fe5b505050505600a165627a7a723058206ff98dad87695205f2e05ac997f7a98d14acb6b2933b565511ae2f3a5ead8f790029"
)

var client *ethclient.Client
var mtx sync.Mutex

// Ethereum Test Suite
type EthereumTestSuite struct {
	suite.Suite
	gateway *services.Eth
	ethClient *ethclient.Client // used for testing
}

//
// Testing Setup/Teardown
//

func (s *EthereumTestSuite) setupSuite(t *testing.T) {
	// EMPTY!!
}
func (s *EthereumTestSuite) tearDownSuite() {
	// EMPTY!!
}

//
// Ethereum Tests
//

// generate address test
func (s *EthereumTestSuite) generateAddress(t *testing.T) {

	// generate eth address using gateway
	addr, privateKey, error := s.gateway.GenerateEthAddr()
	if error != nil {
		t.Fatalf("error creating ethereum network address")
	}
	// ensure address is correct format
	if s.Assert().NotNil(addr) && common.IsHexAddress(addr.Hex()) {
		t.Fatalf("could not create a valid ethereum network address")
	}
	// ensure private key was returned
	if s.Assert().NotNil(privateKey) && privateKey == "" {
		t.Fatalf("could not create a valid private key")
	}
	t.Logf("ethereum network address was generated %v\n", addr.Hex())
}

// get gas price from network test
func (s *EthereumTestSuite) getGasPrice(t *testing.T) {

	// get the suggested gas price
	gasPrice, error := s.gateway.GetGasPrice()
	if error != nil {
		t.Fatalf("error retrieving gas price: %v\n",error)
	}
	if s.Assert().NotNil(gasPrice) && gasPrice.Uint64() > 0 {
		t.Logf("gas price verified: %v\n",gasPrice)
	} else {
		t.Fatalf("gas price less than zero: %v\n",gasPrice)
	}
}

// check balance on sim network test
func (s *EthereumTestSuite) checkBalance(t *testing.T) {

	// test balance for an account
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	// fund the sim

	// Convert string address to byte[] address form
	bal := s.gateway.CheckBalance(auth.From)
	if bal.Uint64() > 0 {
		t.Logf("balance verified: %v\n",bal)
	} else {
		t.Fatalf("balance less than zero: %v\n",bal)
	}
}

func (s *EthereumTestSuite) getCurrentBlock(t *testing.T) {
	// Get the current block from the network
	block, error := services.EthWrapper.GetCurrentBlock()
	if error != nil {
		t.Fatalf("could not retrieve the current block: %v\n",error)
	}
	if block != nil {
		t.Logf("retrieved the current block: %v\n",block)
	}
}

// send gas for a transaction
func (s *EthereumTestSuite) sendGas(t *testing.T) {
	// Send Gas to an Account
	// WIP - Add once we update the send via contract method
}

// send ether
func (s *EthereumTestSuite) sendEther(t *testing.T) {
	// Send Ether to an Account
	// WIP - Add once we update the send via contract method
}


//
// Oyster Pearl Contract Tests
//

// deploy the compiled oyster contract to Oysterby network
func (s *EthereumTestSuite) deployContractOnOysterby(t *testing.T) {

	key, err := ioutil.ReadFile("testdata/key.json")
	if err != nil {
		t.Fatalf("Could not load test data:%v",err)
	}
	// initialize the context
	deadline := time.Now().Add(3000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// contract bytecode
	password := ""
	bytecode := common.Hex2Bytes(oysterPearlByteCode)
	abi, _ := abi2.JSON(strings.NewReader(oysterPearlABI))
	input, _ := abi.Pack("")
	bytecode = append(bytecode, input...)
	unlockedKey, _ := keystore.DecryptKey([]byte(key), password)
	nonce, _ := s.ethClient.NonceAt(ctx, unlockedKey.Address, nil)

	// TODO need exact values for gas and gas price from network to get it to pass
	tx := types.NewContractCreation(nonce, big.NewInt(0), big.NewInt(10000000).Uint64(), big.NewInt(0), bytecode)
	signedTx, _ := types.SignTx(tx, types.HomesteadSigner{}, unlockedKey.PrivateKey)

	// we use the ethClient vs gateway, since gateway does not deploy the smart contract
	err = s.ethClient.SendTransaction(ctx, signedTx)

	if err != nil {
		// unable to deploy the smart contract
		t.Fatal("Contract transaction failed to deploy. Try Again?")
	} else {
		// smart contract deployed onto oysterby
		t.Log("Contract has been deployed.")
	}
}

//
// Oyster Pearl Tests
//

// bury prl
func (s *EthereumTestSuite) buryPRL(t *testing.T) {

	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From: common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
		To: common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
		Amount: *big.NewInt(1000),
		Gas: big.NewInt(10000).Uint64(),
		GasPrice: *big.NewInt(1000),
		TotalWei: *big.NewInt(100000),
		Data: []byte(""), // setup data
	}

	// Bury PRL
	var buried = s.gateway.BuryPrl(msg)
	if buried {
		// successful bury attempt
		t.Log("Buried the PRLs successfully")
	} else {
		// failed bury attempt
		t.Fatal("Faild to bury PRLs. Try Again?")
	}
}

// send prl
func (s *EthereumTestSuite) sendPRL(t *testing.T) {

	// prepare oyster message call
	var msg = services.OysterCallMsg{
		From: common.HexToAddress("0x0d1d4e623d10f9fba5db95830f7d3839406c6af2"),
		To: common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732"),
		Amount: *big.NewInt(1000),
		Gas: big.NewInt(10000).Uint64(),
		GasPrice: *big.NewInt(1000),
		TotalWei: *big.NewInt(100000),
		Data: []byte(""), // setup data // TODO finalize by adding contract call to
	}

	// Send PRL
	var sent = s.gateway.SendPRL(msg)
	if sent {
		// successful prl send
		t.Logf("Sent PRL to :%v",msg.From.Hex())
	} else {
		// failed prl send
		t.Fatalf("Failed to send PRL to:%v",msg.From.Hex())
	}
}

// claim prl
func (s *EthereumTestSuite) claimPRL(t *testing.T) {

	// Need to fake the completed uploads by populating with data
	var rowWithGasTransferSuccess = models.CompletedUpload{
		GenesisHash:   "RowWithGasTransferSuccess",
		ETHAddr:       "0x5aeda56215b167893e80b4fe645ba6d5bab767de",
		ETHPrivateKey: "8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5",
		PRLStatus:     models.PRLClaimNotStarted,
		GasStatus:     models.GasTransferSuccess,
	}

	// mock completed upload
	completedUploads := []models.CompletedUpload{rowWithGasTransferSuccess}

	// Claim PRL
	err := s.gateway.ClaimPRLs(completedUploads)
	if err != nil {
		t.Fatal("Failed to claim PRLs")
	} else {
		t.Log("PRLs have been successfully claimed")
	}

}

// subscribe to transfer
func (s *EthereumTestSuite) subscribeToTransfer(t *testing.T) {

	// Subscribe to a Transaction
	// subscribeToTransfer(brokerAddr common.Address, outCh chan<- types.Log)
	broker := common.HexToAddress("")
	channel := make(chan types.Log)
	s.gateway.SubscribeToTransfer(broker, channel)
}


