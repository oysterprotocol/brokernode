// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package services

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// OysterPearlABI is the input ABI used to generate the binding from.
const OysterPearlABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"openSale\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payout\",\"type\":\"address\"},{\"name\":\"_fee\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"retentionMax\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdrawFunds\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"buried\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"epochSet\",\"type\":\"uint256\"}],\"name\":\"amendEpoch\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"director\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"retentionSet\",\"type\":\"uint8\"},{\"name\":\"accuracy\",\"type\":\"uint8\"}],\"name\":\"amendRetention\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bury\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"feeAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"burnFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"claimAmountSet\",\"type\":\"uint8\"},{\"name\":\"payAmountSet\",\"type\":\"uint8\"},{\"name\":\"feeAmountSet\",\"type\":\"uint8\"},{\"name\":\"accuracy\",\"type\":\"uint8\"}],\"name\":\"amendClaim\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"claimAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"epoch\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"saleClosed\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"payAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"claimed\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"funds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"approveAndCall\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"selfLock\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newDirector\",\"type\":\"address\"}],\"name\":\"transferDirector\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"closeSale\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"directorLock\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_target\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Bury\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_target\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_payout\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_fee\",\"type\":\"address\"}],\"name\":\"Claim\",\"type\":\"event\"}]"

// OysterPearlBin is the compiled bytecode used for deploying new contracts.
const OysterPearlBin = `0x60058054600160a060020a03191633600160a060020a031617905560c0604052600c60808190527f4f797374657220506561726c000000000000000000000000000000000000000060a09081526200005b916000919062000164565b506040805180820190915260038082527f50524c00000000000000000000000000000000000000000000000000000000006020909201918252620000a29160019162000164565b5060028054601260ff1990911617808255600580547401000000000000000000000000000000000000000060a060020a60ff02199091161760a860020a60ff0219811682556000600481815563017d784060ff958616600a90810a91820263047868c0830201627a120092909202919091016003819055600160a060020a039094168352600b60205260409092209290925593549092166000198101840a9182026006559181026007556008556301e13380600955810a602802905562000209565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10620001a757805160ff1916838001178555620001d7565b82800160010185558215620001d7579182015b82811115620001d7578251825591602001919060010190620001ba565b50620001e5929150620001e9565b5090565b6200020691905b80821115620001e55760008155600101620001f0565b90565b6114a180620002196000396000f3006080604052600436106101ab5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde03811461025a578063095ea7b3146102e4578063167ff46f1461031c57806318160ddd1461033157806321c0b3421461035857806322bb4f531461037f57806323b872dd1461039457806324600fc3146103be57806327e235e3146103d5578063313ce567146103f65780633f1199e61461042157806342966c6814610442578063549215a31461045a5780635af82abf146104725780635f5f2aef146104a357806361161aae146104c457806369e15404146104d957806370a08231146104ee57806379cc67901461050f5780637dbc9fba14610533578063830953ab14610560578063900cf0cf1461057557806395d89b411461058a578063a9059cbb1461059f578063b8c766b8146105c3578063c8705544146105d8578063c884ef83146105ed578063c89f2ce41461060e578063cae9ca5114610623578063d1e7e81f1461068c578063dd62ed3e14610694578063ddd41ef6146106bb578063ee55efee146106dc578063ffe2d77e146106f1575b60055460009060a060020a900460ff16156101c557600080fd5b66038d7ea4c680003410156101d957600080fd5b5060025460035434611388029160ff16600a0a631dcd650002908201111561020057600080fd5b6003805482019055600160a060020a033381166000818152600b602090815260409182902080548601905560048054340190558151858152915192933016926000805160206114568339815191529281900390910190a350005b34801561026657600080fd5b5061026f610706565b6040805160208082528351818301528351919283929083019185019080838360005b838110156102a9578181015183820152602001610291565b50505050905090810190601f1680156102d65780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156102f057600080fd5b50610308600160a060020a0360043516602435610794565b604080519115158252519081900360200190f35b34801561032857600080fd5b50610308610820565b34801561033d57600080fd5b50610346610891565b60408051918252519081900360200190f35b34801561036457600080fd5b50610308600160a060020a0360043581169060243516610897565b34801561038b57600080fd5b50610346610ada565b3480156103a057600080fd5b50610308600160a060020a0360043581169060243516604435610ae0565b3480156103ca57600080fd5b506103d3610b4d565b005b3480156103e157600080fd5b50610346600160a060020a0360043516610ba6565b34801561040257600080fd5b5061040b610bb8565b6040805160ff9092168252519081900360200190f35b34801561042d57600080fd5b50610308600160a060020a0360043516610bc1565b34801561044e57600080fd5b50610308600435610bd6565b34801561046657600080fd5b50610308600435610c86565b34801561047e57600080fd5b50610487610cc4565b60408051600160a060020a039092168252519081900360200190f35b3480156104af57600080fd5b5061030860ff60043581169060243516610cd3565b3480156104d057600080fd5b50610308610d29565b3480156104e557600080fd5b50610346610e0e565b3480156104fa57600080fd5b50610346600160a060020a0360043516610e14565b34801561051b57600080fd5b50610308600160a060020a0360043516602435610e2f565b34801561053f57600080fd5b5061030860ff60043581169060243581169060443581169060643516610f31565b34801561056c57600080fd5b50610346610fa6565b34801561058157600080fd5b50610346610fac565b34801561059657600080fd5b5061026f610fb2565b3480156105ab57600080fd5b506103d3600160a060020a036004351660243561100c565b3480156105cf57600080fd5b5061030861101b565b3480156105e457600080fd5b5061034661102b565b3480156105f957600080fd5b50610346600160a060020a0360043516611031565b34801561061a57600080fd5b50610346611043565b34801561062f57600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610308948235600160a060020a03169460248035953695946064949201919081908401838280828437509497506110499650505050505050565b6103d3611180565b3480156106a057600080fd5b50610346600160a060020a0360043581169060243516611205565b3480156106c757600080fd5b506103d3600160a060020a0360043516611222565b3480156106e857600080fd5b5061030861126c565b3480156106fd57600080fd5b506103086112e2565b6000805460408051602060026001851615610100026000190190941693909304601f8101849004840282018401909252818152929183018282801561078c5780601f106107615761010080835404028352916020019161078c565b820191906000526020600020905b81548152906001019060200180831161076f57829003601f168201915b505050505081565b600160a060020a0333166000908152600d602052604081205460ff16156107ba57600080fd5b600160a060020a033381166000818152600c6020908152604080832094881680845294825291829020869055815186815291517f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259281900390910190a350600192915050565b60055460009060a860020a900460ff161561083a57600080fd5b60055433600160a060020a0390811691161461085557600080fd5b60055460a060020a900460ff16151561086d57600080fd5b506005805474ff000000000000000000000000000000000000000019169055600190565b60035481565b600160a060020a0333166000908152600d6020526040812054819060ff1615156108c057600080fd5b600160a060020a0384811690841614156108d957600080fd5b83600160a060020a031633600160a060020a0316141515156108fa57600080fd5b82600160a060020a031633600160a060020a03161415151561091b57600080fd5b600160a060020a0333166000908152600e60205260409020546001148061095e5750600954600160a060020a0333166000908152600e6020526040902054420310155b151561096957600080fd5b600654600160a060020a0333166000908152600b6020526040902054101561099057600080fd5b50600160a060020a033381166000818152600e60209081526040808320429055868516808452600b909252808320805495891680855282852080548787528487208054600654810390915560075483540190925560085486885284540190935592519290910190950194919391927fcac3ed26c9dd72a2c44999857298af9c72ba2d1ca9784f5dad48c933e2224c1191a483600160a060020a031633600160a060020a03166000805160206114568339815191526007546040518082815260200191505060405180910390a382600160a060020a031633600160a060020a03166000805160206114568339815191526008546040518082815260200191505060405180910390a3600160a060020a038084166000908152600b602052604080822054878416835281832054339094168352912054909101018114610ad057fe5b5060019392505050565b600a5481565b600160a060020a038084166000908152600c6020908152604080832033909416835292905290812054821115610b1557600080fd5b600160a060020a038085166000908152600c602090815260408083203390941683529290522080548390039055610ad08484846112f2565b60055433600160a060020a03908116911614610b6857600080fd5b600554604051600160a060020a039182169130163180156108fc02916000818181858888f19350505050158015610ba3573d6000803e3d6000fd5b50565b600b6020526000908152604090205481565b60025460ff1681565b600d6020526000908152604090205460ff1681565b600160a060020a0333166000908152600d602052604081205460ff1615610bfc57600080fd5b600160a060020a0333166000908152600b6020526040902054821115610c2157600080fd5b600160a060020a0333166000818152600b602090815260409182902080548690039055600380548690039055815185815291517fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca59281900390910190a2506001919050565b60055460009060a860020a900460ff1615610ca057600080fd5b60055433600160a060020a03908116911614610cbb57600080fd5b50600955600190565b600554600160a060020a031681565b60055460009060a860020a900460ff1615610ced57600080fd5b60055433600160a060020a03908116911614610d0857600080fd5b5060025460ff91821690821603600a90810a92909116919091029055600190565b600160a060020a0333166000908152600d602052604081205460ff1615610d4f57600080fd5b600654600160a060020a0333166000908152600b60205260409020541015610d7657600080fd5b600a54600160a060020a0333166000908152600b60205260409020541115610d9d57600080fd5b600160a060020a0333166000818152600d60209081526040808320805460ff19166001908117909155600e835281842055600b82529182902054825190815291517fc96e8fee6eb65975d592ca9a340f33200433df4c42b2f623dd9fc6d22984d4959281900390910190a250600190565b60085481565b600160a060020a03166000908152600b602052604090205490565b600160a060020a0382166000908152600d602052604081205460ff1615610e5557600080fd5b600160a060020a0383166000908152600b6020526040902054821115610e7a57600080fd5b600160a060020a038084166000908152600c602090815260408083203390941683529290522054821115610ead57600080fd5b600160a060020a038084166000818152600b6020908152604080832080548890039055600c825280832033909516835293815290839020805486900390556003805486900390558251858152925191927fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5929081900390910190a250600192915050565b60055460009060a860020a900460ff1615610f4b57600080fd5b60055433600160a060020a03908116911614610f6657600080fd5b82840160ff168560ff16141515610f7c57600080fd5b5060025460ff91821690821603600a0a938116840260065591821683026007551602600855600190565b60065481565b60095481565b60018054604080516020600284861615610100026000190190941693909304601f8101849004840282018401909252818152929183018282801561078c5780601f106107615761010080835404028352916020019161078c565b6110173383836112f2565b5050565b60055460a060020a900460ff1681565b60075481565b600e6020526000908152604090205481565b60045481565b6000836110568185610794565b156111785780600160a060020a0316638f4ffcb1338630876040518563ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018085600160a060020a0316600160a060020a0316815260200184815260200183600160a060020a0316600160a060020a0316815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561110c5781810151838201526020016110f4565b50505050905090810190601f1680156111395780820380516001836020036101000a031916815260200191505b5095505050505050600060405180830381600087803b15801561115b57600080fd5b505af115801561116f573d6000803e3d6000fd5b50505050600191505b509392505050565b60055460a860020a900460ff161561119757600080fd5b60055433600160a060020a039081169116146111b257600080fd5b60055460a060020a900460ff1615156111ca57600080fd5b678ac7230489e8000034146111de57600080fd5b6005805475ff000000000000000000000000000000000000000000191660a860020a179055565b600c60209081526000928352604080842090915290825290205481565b60055433600160a060020a0390811691161461123d57600080fd5b6005805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b60055460009060a860020a900460ff161561128657600080fd5b60055433600160a060020a039081169116146112a157600080fd5b60055460a060020a900460ff16156112b857600080fd5b506005805474ff0000000000000000000000000000000000000000191660a060020a179055600190565b60055460a860020a900460ff1681565b600160a060020a0383166000908152600d602052604081205460ff161561131857600080fd5b600160a060020a0383166000908152600d602052604090205460ff161561136257600a54600160a060020a0384166000908152600b60205260409020548301111561136257600080fd5b600160a060020a038316151561137757600080fd5b600160a060020a0384166000908152600b602052604090205482111561139c57600080fd5b600160a060020a0383166000908152600b6020526040902054828101116113c257600080fd5b50600160a060020a038083166000818152600b6020908152604080832080549589168085528285208054898103909155948690528154880190915581518781529151939095019492600080516020611456833981519152929181900390910190a3600160a060020a038084166000908152600b602052604080822054928716825290205401811461144f57fe5b505050505600ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa165627a7a7230582037aa33fc8c6abe16a6686d38aed5bf4c9cea9e47ce402bf055ae9777bfd1cf4c0029`

// DeployOysterPearl deploys a new Ethereum contract, binding an instance of OysterPearl to it.
func DeployOysterPearl(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *OysterPearl, error) {
	parsed, err := abi.JSON(strings.NewReader(OysterPearlABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OysterPearlBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OysterPearl{OysterPearlCaller: OysterPearlCaller{contract: contract}, OysterPearlTransactor: OysterPearlTransactor{contract: contract}, OysterPearlFilterer: OysterPearlFilterer{contract: contract}}, nil
}

// OysterPearl is an auto generated Go binding around an Ethereum contract.
type OysterPearl struct {
	OysterPearlCaller     // Read-only binding to the contract
	OysterPearlTransactor // Write-only binding to the contract
	OysterPearlFilterer   // Log filterer for contract events
}

// OysterPearlCaller is an auto generated read-only Go binding around an Ethereum contract.
type OysterPearlCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OysterPearlTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OysterPearlTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OysterPearlFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OysterPearlFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OysterPearlSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OysterPearlSession struct {
	Contract     *OysterPearl      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OysterPearlCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OysterPearlCallerSession struct {
	Contract *OysterPearlCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// OysterPearlTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OysterPearlTransactorSession struct {
	Contract     *OysterPearlTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// OysterPearlRaw is an auto generated low-level Go binding around an Ethereum contract.
type OysterPearlRaw struct {
	Contract *OysterPearl // Generic contract binding to access the raw methods on
}

// OysterPearlCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OysterPearlCallerRaw struct {
	Contract *OysterPearlCaller // Generic read-only contract binding to access the raw methods on
}

// OysterPearlTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OysterPearlTransactorRaw struct {
	Contract *OysterPearlTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOysterPearl creates a new instance of OysterPearl, bound to a specific deployed contract.
func NewOysterPearl(address common.Address, backend bind.ContractBackend) (*OysterPearl, error) {
	contract, err := bindOysterPearl(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OysterPearl{OysterPearlCaller: OysterPearlCaller{contract: contract}, OysterPearlTransactor: OysterPearlTransactor{contract: contract}, OysterPearlFilterer: OysterPearlFilterer{contract: contract}}, nil
}

// NewOysterPearlCaller creates a new read-only instance of OysterPearl, bound to a specific deployed contract.
func NewOysterPearlCaller(address common.Address, caller bind.ContractCaller) (*OysterPearlCaller, error) {
	contract, err := bindOysterPearl(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OysterPearlCaller{contract: contract}, nil
}

// NewOysterPearlTransactor creates a new write-only instance of OysterPearl, bound to a specific deployed contract.
func NewOysterPearlTransactor(address common.Address, transactor bind.ContractTransactor) (*OysterPearlTransactor, error) {
	contract, err := bindOysterPearl(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OysterPearlTransactor{contract: contract}, nil
}

// NewOysterPearlFilterer creates a new log filterer instance of OysterPearl, bound to a specific deployed contract.
func NewOysterPearlFilterer(address common.Address, filterer bind.ContractFilterer) (*OysterPearlFilterer, error) {
	contract, err := bindOysterPearl(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OysterPearlFilterer{contract: contract}, nil
}

// bindOysterPearl binds a generic wrapper to an already deployed contract.
func bindOysterPearl(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OysterPearlABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OysterPearl *OysterPearlRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OysterPearl.Contract.OysterPearlCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OysterPearl *OysterPearlRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.Contract.OysterPearlTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OysterPearl *OysterPearlRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OysterPearl.Contract.OysterPearlTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OysterPearl *OysterPearlCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OysterPearl.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OysterPearl *OysterPearlTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OysterPearl *OysterPearlTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OysterPearl.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance( address,  address) constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "allowance", arg0, arg1)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance( address,  address) constant returns(uint256)
func (_OysterPearl *OysterPearlSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Allowance(&_OysterPearl.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance( address,  address) constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Allowance(&_OysterPearl.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_OysterPearl *OysterPearlCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_OysterPearl *OysterPearlSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.BalanceOf(&_OysterPearl.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_OysterPearl *OysterPearlCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.BalanceOf(&_OysterPearl.CallOpts, _owner)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) Balances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "balances", arg0)
	return *ret0, err
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_OysterPearl *OysterPearlSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Balances(&_OysterPearl.CallOpts, arg0)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Balances(&_OysterPearl.CallOpts, arg0)
}

// Buried is a free data retrieval call binding the contract method 0x3f1199e6.
//
// Solidity: function buried( address) constant returns(bool)
func (_OysterPearl *OysterPearlCaller) Buried(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "buried", arg0)
	return *ret0, err
}

// Buried is a free data retrieval call binding the contract method 0x3f1199e6.
//
// Solidity: function buried( address) constant returns(bool)
func (_OysterPearl *OysterPearlSession) Buried(arg0 common.Address) (bool, error) {
	return _OysterPearl.Contract.Buried(&_OysterPearl.CallOpts, arg0)
}

// Buried is a free data retrieval call binding the contract method 0x3f1199e6.
//
// Solidity: function buried( address) constant returns(bool)
func (_OysterPearl *OysterPearlCallerSession) Buried(arg0 common.Address) (bool, error) {
	return _OysterPearl.Contract.Buried(&_OysterPearl.CallOpts, arg0)
}

// ClaimAmount is a free data retrieval call binding the contract method 0x830953ab.
//
// Solidity: function claimAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) ClaimAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "claimAmount")
	return *ret0, err
}

// ClaimAmount is a free data retrieval call binding the contract method 0x830953ab.
//
// Solidity: function claimAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) ClaimAmount() (*big.Int, error) {
	return _OysterPearl.Contract.ClaimAmount(&_OysterPearl.CallOpts)
}

// ClaimAmount is a free data retrieval call binding the contract method 0x830953ab.
//
// Solidity: function claimAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) ClaimAmount() (*big.Int, error) {
	return _OysterPearl.Contract.ClaimAmount(&_OysterPearl.CallOpts)
}

// Claimed is a free data retrieval call binding the contract method 0xc884ef83.
//
// Solidity: function claimed( address) constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) Claimed(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "claimed", arg0)
	return *ret0, err
}

// Claimed is a free data retrieval call binding the contract method 0xc884ef83.
//
// Solidity: function claimed( address) constant returns(uint256)
func (_OysterPearl *OysterPearlSession) Claimed(arg0 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Claimed(&_OysterPearl.CallOpts, arg0)
}

// Claimed is a free data retrieval call binding the contract method 0xc884ef83.
//
// Solidity: function claimed( address) constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) Claimed(arg0 common.Address) (*big.Int, error) {
	return _OysterPearl.Contract.Claimed(&_OysterPearl.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_OysterPearl *OysterPearlCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_OysterPearl *OysterPearlSession) Decimals() (uint8, error) {
	return _OysterPearl.Contract.Decimals(&_OysterPearl.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_OysterPearl *OysterPearlCallerSession) Decimals() (uint8, error) {
	return _OysterPearl.Contract.Decimals(&_OysterPearl.CallOpts)
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_OysterPearl *OysterPearlCaller) Director(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "director")
	return *ret0, err
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_OysterPearl *OysterPearlSession) Director() (common.Address, error) {
	return _OysterPearl.Contract.Director(&_OysterPearl.CallOpts)
}

// Director is a free data retrieval call binding the contract method 0x5af82abf.
//
// Solidity: function director() constant returns(address)
func (_OysterPearl *OysterPearlCallerSession) Director() (common.Address, error) {
	return _OysterPearl.Contract.Director(&_OysterPearl.CallOpts)
}

// DirectorLock is a free data retrieval call binding the contract method 0xffe2d77e.
//
// Solidity: function directorLock() constant returns(bool)
func (_OysterPearl *OysterPearlCaller) DirectorLock(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "directorLock")
	return *ret0, err
}

// DirectorLock is a free data retrieval call binding the contract method 0xffe2d77e.
//
// Solidity: function directorLock() constant returns(bool)
func (_OysterPearl *OysterPearlSession) DirectorLock() (bool, error) {
	return _OysterPearl.Contract.DirectorLock(&_OysterPearl.CallOpts)
}

// DirectorLock is a free data retrieval call binding the contract method 0xffe2d77e.
//
// Solidity: function directorLock() constant returns(bool)
func (_OysterPearl *OysterPearlCallerSession) DirectorLock() (bool, error) {
	return _OysterPearl.Contract.DirectorLock(&_OysterPearl.CallOpts)
}

// Epoch is a free data retrieval call binding the contract method 0x900cf0cf.
//
// Solidity: function epoch() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) Epoch(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "epoch")
	return *ret0, err
}

// Epoch is a free data retrieval call binding the contract method 0x900cf0cf.
//
// Solidity: function epoch() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) Epoch() (*big.Int, error) {
	return _OysterPearl.Contract.Epoch(&_OysterPearl.CallOpts)
}

// Epoch is a free data retrieval call binding the contract method 0x900cf0cf.
//
// Solidity: function epoch() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) Epoch() (*big.Int, error) {
	return _OysterPearl.Contract.Epoch(&_OysterPearl.CallOpts)
}

// FeeAmount is a free data retrieval call binding the contract method 0x69e15404.
//
// Solidity: function feeAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) FeeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "feeAmount")
	return *ret0, err
}

// FeeAmount is a free data retrieval call binding the contract method 0x69e15404.
//
// Solidity: function feeAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) FeeAmount() (*big.Int, error) {
	return _OysterPearl.Contract.FeeAmount(&_OysterPearl.CallOpts)
}

// FeeAmount is a free data retrieval call binding the contract method 0x69e15404.
//
// Solidity: function feeAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) FeeAmount() (*big.Int, error) {
	return _OysterPearl.Contract.FeeAmount(&_OysterPearl.CallOpts)
}

// Funds is a free data retrieval call binding the contract method 0xc89f2ce4.
//
// Solidity: function funds() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) Funds(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "funds")
	return *ret0, err
}

// Funds is a free data retrieval call binding the contract method 0xc89f2ce4.
//
// Solidity: function funds() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) Funds() (*big.Int, error) {
	return _OysterPearl.Contract.Funds(&_OysterPearl.CallOpts)
}

// Funds is a free data retrieval call binding the contract method 0xc89f2ce4.
//
// Solidity: function funds() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) Funds() (*big.Int, error) {
	return _OysterPearl.Contract.Funds(&_OysterPearl.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_OysterPearl *OysterPearlCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_OysterPearl *OysterPearlSession) Name() (string, error) {
	return _OysterPearl.Contract.Name(&_OysterPearl.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_OysterPearl *OysterPearlCallerSession) Name() (string, error) {
	return _OysterPearl.Contract.Name(&_OysterPearl.CallOpts)
}

// PayAmount is a free data retrieval call binding the contract method 0xc8705544.
//
// Solidity: function payAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) PayAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "payAmount")
	return *ret0, err
}

// PayAmount is a free data retrieval call binding the contract method 0xc8705544.
//
// Solidity: function payAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) PayAmount() (*big.Int, error) {
	return _OysterPearl.Contract.PayAmount(&_OysterPearl.CallOpts)
}

// PayAmount is a free data retrieval call binding the contract method 0xc8705544.
//
// Solidity: function payAmount() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) PayAmount() (*big.Int, error) {
	return _OysterPearl.Contract.PayAmount(&_OysterPearl.CallOpts)
}

// RetentionMax is a free data retrieval call binding the contract method 0x22bb4f53.
//
// Solidity: function retentionMax() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) RetentionMax(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "retentionMax")
	return *ret0, err
}

// RetentionMax is a free data retrieval call binding the contract method 0x22bb4f53.
//
// Solidity: function retentionMax() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) RetentionMax() (*big.Int, error) {
	return _OysterPearl.Contract.RetentionMax(&_OysterPearl.CallOpts)
}

// RetentionMax is a free data retrieval call binding the contract method 0x22bb4f53.
//
// Solidity: function retentionMax() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) RetentionMax() (*big.Int, error) {
	return _OysterPearl.Contract.RetentionMax(&_OysterPearl.CallOpts)
}

// SaleClosed is a free data retrieval call binding the contract method 0xb8c766b8.
//
// Solidity: function saleClosed() constant returns(bool)
func (_OysterPearl *OysterPearlCaller) SaleClosed(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "saleClosed")
	return *ret0, err
}

// SaleClosed is a free data retrieval call binding the contract method 0xb8c766b8.
//
// Solidity: function saleClosed() constant returns(bool)
func (_OysterPearl *OysterPearlSession) SaleClosed() (bool, error) {
	return _OysterPearl.Contract.SaleClosed(&_OysterPearl.CallOpts)
}

// SaleClosed is a free data retrieval call binding the contract method 0xb8c766b8.
//
// Solidity: function saleClosed() constant returns(bool)
func (_OysterPearl *OysterPearlCallerSession) SaleClosed() (bool, error) {
	return _OysterPearl.Contract.SaleClosed(&_OysterPearl.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_OysterPearl *OysterPearlCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_OysterPearl *OysterPearlSession) Symbol() (string, error) {
	return _OysterPearl.Contract.Symbol(&_OysterPearl.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_OysterPearl *OysterPearlCallerSession) Symbol() (string, error) {
	return _OysterPearl.Contract.Symbol(&_OysterPearl.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_OysterPearl *OysterPearlCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _OysterPearl.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_OysterPearl *OysterPearlSession) TotalSupply() (*big.Int, error) {
	return _OysterPearl.Contract.TotalSupply(&_OysterPearl.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_OysterPearl *OysterPearlCallerSession) TotalSupply() (*big.Int, error) {
	return _OysterPearl.Contract.TotalSupply(&_OysterPearl.CallOpts)
}

// AmendClaim is a paid mutator transaction binding the contract method 0x7dbc9fba.
//
// Solidity: function amendClaim(claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) AmendClaim(opts *bind.TransactOpts, claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "amendClaim", claimAmountSet, payAmountSet, feeAmountSet, accuracy)
}

// AmendClaim is a paid mutator transaction binding the contract method 0x7dbc9fba.
//
// Solidity: function amendClaim(claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlSession) AmendClaim(claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendClaim(&_OysterPearl.TransactOpts, claimAmountSet, payAmountSet, feeAmountSet, accuracy)
}

// AmendClaim is a paid mutator transaction binding the contract method 0x7dbc9fba.
//
// Solidity: function amendClaim(claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) AmendClaim(claimAmountSet uint8, payAmountSet uint8, feeAmountSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendClaim(&_OysterPearl.TransactOpts, claimAmountSet, payAmountSet, feeAmountSet, accuracy)
}

// AmendEpoch is a paid mutator transaction binding the contract method 0x549215a3.
//
// Solidity: function amendEpoch(epochSet uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) AmendEpoch(opts *bind.TransactOpts, epochSet *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "amendEpoch", epochSet)
}

// AmendEpoch is a paid mutator transaction binding the contract method 0x549215a3.
//
// Solidity: function amendEpoch(epochSet uint256) returns(success bool)
func (_OysterPearl *OysterPearlSession) AmendEpoch(epochSet *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendEpoch(&_OysterPearl.TransactOpts, epochSet)
}

// AmendEpoch is a paid mutator transaction binding the contract method 0x549215a3.
//
// Solidity: function amendEpoch(epochSet uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) AmendEpoch(epochSet *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendEpoch(&_OysterPearl.TransactOpts, epochSet)
}

// AmendRetention is a paid mutator transaction binding the contract method 0x5f5f2aef.
//
// Solidity: function amendRetention(retentionSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) AmendRetention(opts *bind.TransactOpts, retentionSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "amendRetention", retentionSet, accuracy)
}

// AmendRetention is a paid mutator transaction binding the contract method 0x5f5f2aef.
//
// Solidity: function amendRetention(retentionSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlSession) AmendRetention(retentionSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendRetention(&_OysterPearl.TransactOpts, retentionSet, accuracy)
}

// AmendRetention is a paid mutator transaction binding the contract method 0x5f5f2aef.
//
// Solidity: function amendRetention(retentionSet uint8, accuracy uint8) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) AmendRetention(retentionSet uint8, accuracy uint8) (*types.Transaction, error) {
	return _OysterPearl.Contract.AmendRetention(&_OysterPearl.TransactOpts, retentionSet, accuracy)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Approve(&_OysterPearl.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Approve(&_OysterPearl.TransactOpts, _spender, _value)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(_spender address, _value uint256, _extraData bytes) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) ApproveAndCall(opts *bind.TransactOpts, _spender common.Address, _value *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "approveAndCall", _spender, _value, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(_spender address, _value uint256, _extraData bytes) returns(success bool)
func (_OysterPearl *OysterPearlSession) ApproveAndCall(_spender common.Address, _value *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _OysterPearl.Contract.ApproveAndCall(&_OysterPearl.TransactOpts, _spender, _value, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(_spender address, _value uint256, _extraData bytes) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) ApproveAndCall(_spender common.Address, _value *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _OysterPearl.Contract.ApproveAndCall(&_OysterPearl.TransactOpts, _spender, _value, _extraData)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(_value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) Burn(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "burn", _value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(_value uint256) returns(success bool)
func (_OysterPearl *OysterPearlSession) Burn(_value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Burn(&_OysterPearl.TransactOpts, _value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(_value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) Burn(_value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Burn(&_OysterPearl.TransactOpts, _value)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(_from address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) BurnFrom(opts *bind.TransactOpts, _from common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "burnFrom", _from, _value)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(_from address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlSession) BurnFrom(_from common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.BurnFrom(&_OysterPearl.TransactOpts, _from, _value)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(_from address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) BurnFrom(_from common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.BurnFrom(&_OysterPearl.TransactOpts, _from, _value)
}

// Bury is a paid mutator transaction binding the contract method 0x61161aae.
//
// Solidity: function bury() returns(success bool)
func (_OysterPearl *OysterPearlTransactor) Bury(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "bury")
}

// Bury is a paid mutator transaction binding the contract method 0x61161aae.
//
// Solidity: function bury() returns(success bool)
func (_OysterPearl *OysterPearlSession) Bury() (*types.Transaction, error) {
	return _OysterPearl.Contract.Bury(&_OysterPearl.TransactOpts)
}

// Bury is a paid mutator transaction binding the contract method 0x61161aae.
//
// Solidity: function bury() returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) Bury() (*types.Transaction, error) {
	return _OysterPearl.Contract.Bury(&_OysterPearl.TransactOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x21c0b342.
//
// Solidity: function claim(_payout address, _fee address) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) Claim(opts *bind.TransactOpts, _payout common.Address, _fee common.Address) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "claim", _payout, _fee)
}

// Claim is a paid mutator transaction binding the contract method 0x21c0b342.
//
// Solidity: function claim(_payout address, _fee address) returns(success bool)
func (_OysterPearl *OysterPearlSession) Claim(_payout common.Address, _fee common.Address) (*types.Transaction, error) {
	return _OysterPearl.Contract.Claim(&_OysterPearl.TransactOpts, _payout, _fee)
}

// Claim is a paid mutator transaction binding the contract method 0x21c0b342.
//
// Solidity: function claim(_payout address, _fee address) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) Claim(_payout common.Address, _fee common.Address) (*types.Transaction, error) {
	return _OysterPearl.Contract.Claim(&_OysterPearl.TransactOpts, _payout, _fee)
}

// CloseSale is a paid mutator transaction binding the contract method 0xee55efee.
//
// Solidity: function closeSale() returns(success bool)
func (_OysterPearl *OysterPearlTransactor) CloseSale(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "closeSale")
}

// CloseSale is a paid mutator transaction binding the contract method 0xee55efee.
//
// Solidity: function closeSale() returns(success bool)
func (_OysterPearl *OysterPearlSession) CloseSale() (*types.Transaction, error) {
	return _OysterPearl.Contract.CloseSale(&_OysterPearl.TransactOpts)
}

// CloseSale is a paid mutator transaction binding the contract method 0xee55efee.
//
// Solidity: function closeSale() returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) CloseSale() (*types.Transaction, error) {
	return _OysterPearl.Contract.CloseSale(&_OysterPearl.TransactOpts)
}

// OpenSale is a paid mutator transaction binding the contract method 0x167ff46f.
//
// Solidity: function openSale() returns(success bool)
func (_OysterPearl *OysterPearlTransactor) OpenSale(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "openSale")
}

// OpenSale is a paid mutator transaction binding the contract method 0x167ff46f.
//
// Solidity: function openSale() returns(success bool)
func (_OysterPearl *OysterPearlSession) OpenSale() (*types.Transaction, error) {
	return _OysterPearl.Contract.OpenSale(&_OysterPearl.TransactOpts)
}

// OpenSale is a paid mutator transaction binding the contract method 0x167ff46f.
//
// Solidity: function openSale() returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) OpenSale() (*types.Transaction, error) {
	return _OysterPearl.Contract.OpenSale(&_OysterPearl.TransactOpts)
}

// SelfLock is a paid mutator transaction binding the contract method 0xd1e7e81f.
//
// Solidity: function selfLock() returns()
func (_OysterPearl *OysterPearlTransactor) SelfLock(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "selfLock")
}

// SelfLock is a paid mutator transaction binding the contract method 0xd1e7e81f.
//
// Solidity: function selfLock() returns()
func (_OysterPearl *OysterPearlSession) SelfLock() (*types.Transaction, error) {
	return _OysterPearl.Contract.SelfLock(&_OysterPearl.TransactOpts)
}

// SelfLock is a paid mutator transaction binding the contract method 0xd1e7e81f.
//
// Solidity: function selfLock() returns()
func (_OysterPearl *OysterPearlTransactorSession) SelfLock() (*types.Transaction, error) {
	return _OysterPearl.Contract.SelfLock(&_OysterPearl.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_OysterPearl *OysterPearlTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_OysterPearl *OysterPearlSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Transfer(&_OysterPearl.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_OysterPearl *OysterPearlTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.Transfer(&_OysterPearl.TransactOpts, _to, _value)
}

// TransferDirector is a paid mutator transaction binding the contract method 0xddd41ef6.
//
// Solidity: function transferDirector(newDirector address) returns()
func (_OysterPearl *OysterPearlTransactor) TransferDirector(opts *bind.TransactOpts, newDirector common.Address) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "transferDirector", newDirector)
}

// TransferDirector is a paid mutator transaction binding the contract method 0xddd41ef6.
//
// Solidity: function transferDirector(newDirector address) returns()
func (_OysterPearl *OysterPearlSession) TransferDirector(newDirector common.Address) (*types.Transaction, error) {
	return _OysterPearl.Contract.TransferDirector(&_OysterPearl.TransactOpts, newDirector)
}

// TransferDirector is a paid mutator transaction binding the contract method 0xddd41ef6.
//
// Solidity: function transferDirector(newDirector address) returns()
func (_OysterPearl *OysterPearlTransactorSession) TransferDirector(newDirector common.Address) (*types.Transaction, error) {
	return _OysterPearl.Contract.TransferDirector(&_OysterPearl.TransactOpts, newDirector)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.TransferFrom(&_OysterPearl.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_OysterPearl *OysterPearlTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _OysterPearl.Contract.TransferFrom(&_OysterPearl.TransactOpts, _from, _to, _value)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x24600fc3.
//
// Solidity: function withdrawFunds() returns()
func (_OysterPearl *OysterPearlTransactor) WithdrawFunds(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OysterPearl.contract.Transact(opts, "withdrawFunds")
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x24600fc3.
//
// Solidity: function withdrawFunds() returns()
func (_OysterPearl *OysterPearlSession) WithdrawFunds() (*types.Transaction, error) {
	return _OysterPearl.Contract.WithdrawFunds(&_OysterPearl.TransactOpts)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x24600fc3.
//
// Solidity: function withdrawFunds() returns()
func (_OysterPearl *OysterPearlTransactorSession) WithdrawFunds() (*types.Transaction, error) {
	return _OysterPearl.Contract.WithdrawFunds(&_OysterPearl.TransactOpts)
}

// OysterPearlApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the OysterPearl contract.
type OysterPearlApprovalIterator struct {
	Event *OysterPearlApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OysterPearlApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OysterPearlApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OysterPearlApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OysterPearlApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OysterPearlApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OysterPearlApproval represents a Approval event raised by the OysterPearl contract.
type OysterPearlApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _spender indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*OysterPearlApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _OysterPearl.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &OysterPearlApprovalIterator{contract: _OysterPearl.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _spender indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *OysterPearlApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _OysterPearl.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OysterPearlApproval)
				if err := _OysterPearl.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// OysterPearlBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the OysterPearl contract.
type OysterPearlBurnIterator struct {
	Event *OysterPearlBurn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OysterPearlBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OysterPearlBurn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OysterPearlBurn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OysterPearlBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OysterPearlBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OysterPearlBurn represents a Burn event raised by the OysterPearl contract.
type OysterPearlBurn struct {
	From  common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: e Burn(_from indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) FilterBurn(opts *bind.FilterOpts, _from []common.Address) (*OysterPearlBurnIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _OysterPearl.contract.FilterLogs(opts, "Burn", _fromRule)
	if err != nil {
		return nil, err
	}
	return &OysterPearlBurnIterator{contract: _OysterPearl.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: e Burn(_from indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *OysterPearlBurn, _from []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _OysterPearl.contract.WatchLogs(opts, "Burn", _fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OysterPearlBurn)
				if err := _OysterPearl.contract.UnpackLog(event, "Burn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// OysterPearlBuryIterator is returned from FilterBury and is used to iterate over the raw logs and unpacked data for Bury events raised by the OysterPearl contract.
type OysterPearlBuryIterator struct {
	Event *OysterPearlBury // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OysterPearlBuryIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OysterPearlBury)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OysterPearlBury)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OysterPearlBuryIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OysterPearlBuryIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OysterPearlBury represents a Bury event raised by the OysterPearl contract.
type OysterPearlBury struct {
	Target common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBury is a free log retrieval operation binding the contract event 0xc96e8fee6eb65975d592ca9a340f33200433df4c42b2f623dd9fc6d22984d495.
//
// Solidity: e Bury(_target indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) FilterBury(opts *bind.FilterOpts, _target []common.Address) (*OysterPearlBuryIterator, error) {

	var _targetRule []interface{}
	for _, _targetItem := range _target {
		_targetRule = append(_targetRule, _targetItem)
	}

	logs, sub, err := _OysterPearl.contract.FilterLogs(opts, "Bury", _targetRule)
	if err != nil {
		return nil, err
	}
	return &OysterPearlBuryIterator{contract: _OysterPearl.contract, event: "Bury", logs: logs, sub: sub}, nil
}

// WatchBury is a free log subscription operation binding the contract event 0xc96e8fee6eb65975d592ca9a340f33200433df4c42b2f623dd9fc6d22984d495.
//
// Solidity: e Bury(_target indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) WatchBury(opts *bind.WatchOpts, sink chan<- *OysterPearlBury, _target []common.Address) (event.Subscription, error) {

	var _targetRule []interface{}
	for _, _targetItem := range _target {
		_targetRule = append(_targetRule, _targetItem)
	}

	logs, sub, err := _OysterPearl.contract.WatchLogs(opts, "Bury", _targetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OysterPearlBury)
				if err := _OysterPearl.contract.UnpackLog(event, "Bury", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// OysterPearlClaimIterator is returned from FilterClaim and is used to iterate over the raw logs and unpacked data for Claim events raised by the OysterPearl contract.
type OysterPearlClaimIterator struct {
	Event *OysterPearlClaim // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OysterPearlClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OysterPearlClaim)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OysterPearlClaim)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OysterPearlClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OysterPearlClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OysterPearlClaim represents a Claim event raised by the OysterPearl contract.
type OysterPearlClaim struct {
	Target common.Address
	Payout common.Address
	Fee    common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0xcac3ed26c9dd72a2c44999857298af9c72ba2d1ca9784f5dad48c933e2224c11.
//
// Solidity: e Claim(_target indexed address, _payout indexed address, _fee indexed address)
func (_OysterPearl *OysterPearlFilterer) FilterClaim(opts *bind.FilterOpts, _target []common.Address, _payout []common.Address, _fee []common.Address) (*OysterPearlClaimIterator, error) {

	var _targetRule []interface{}
	for _, _targetItem := range _target {
		_targetRule = append(_targetRule, _targetItem)
	}
	var _payoutRule []interface{}
	for _, _payoutItem := range _payout {
		_payoutRule = append(_payoutRule, _payoutItem)
	}
	var _feeRule []interface{}
	for _, _feeItem := range _fee {
		_feeRule = append(_feeRule, _feeItem)
	}

	logs, sub, err := _OysterPearl.contract.FilterLogs(opts, "Claim", _targetRule, _payoutRule, _feeRule)
	if err != nil {
		return nil, err
	}
	return &OysterPearlClaimIterator{contract: _OysterPearl.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0xcac3ed26c9dd72a2c44999857298af9c72ba2d1ca9784f5dad48c933e2224c11.
//
// Solidity: e Claim(_target indexed address, _payout indexed address, _fee indexed address)
func (_OysterPearl *OysterPearlFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *OysterPearlClaim, _target []common.Address, _payout []common.Address, _fee []common.Address) (event.Subscription, error) {

	var _targetRule []interface{}
	for _, _targetItem := range _target {
		_targetRule = append(_targetRule, _targetItem)
	}
	var _payoutRule []interface{}
	for _, _payoutItem := range _payout {
		_payoutRule = append(_payoutRule, _payoutItem)
	}
	var _feeRule []interface{}
	for _, _feeItem := range _fee {
		_feeRule = append(_feeRule, _feeItem)
	}

	logs, sub, err := _OysterPearl.contract.WatchLogs(opts, "Claim", _targetRule, _payoutRule, _feeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OysterPearlClaim)
				if err := _OysterPearl.contract.UnpackLog(event, "Claim", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// OysterPearlTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the OysterPearl contract.
type OysterPearlTransferIterator struct {
	Event *OysterPearlTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OysterPearlTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OysterPearlTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OysterPearlTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OysterPearlTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OysterPearlTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OysterPearlTransfer represents a Transfer event raised by the OysterPearl contract.
type OysterPearlTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*OysterPearlTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _OysterPearl.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &OysterPearlTransferIterator{contract: _OysterPearl.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _value uint256)
func (_OysterPearl *OysterPearlFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *OysterPearlTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _OysterPearl.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OysterPearlTransfer)
				if err := _OysterPearl.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// TokenRecipientABI is the input ABI used to generate the binding from.
const TokenRecipientABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"receiveApproval\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// TokenRecipientBin is the compiled bytecode used for deploying new contracts.
const TokenRecipientBin = `0x`

// DeployTokenRecipient deploys a new Ethereum contract, binding an instance of TokenRecipient to it.
func DeployTokenRecipient(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *TokenRecipient, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenRecipientABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TokenRecipientBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TokenRecipient{TokenRecipientCaller: TokenRecipientCaller{contract: contract}, TokenRecipientTransactor: TokenRecipientTransactor{contract: contract}, TokenRecipientFilterer: TokenRecipientFilterer{contract: contract}}, nil
}

// TokenRecipient is an auto generated Go binding around an Ethereum contract.
type TokenRecipient struct {
	TokenRecipientCaller     // Read-only binding to the contract
	TokenRecipientTransactor // Write-only binding to the contract
	TokenRecipientFilterer   // Log filterer for contract events
}

// TokenRecipientCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokenRecipientCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenRecipientTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokenRecipientTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenRecipientFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TokenRecipientFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenRecipientSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TokenRecipientSession struct {
	Contract     *TokenRecipient   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenRecipientCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TokenRecipientCallerSession struct {
	Contract *TokenRecipientCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// TokenRecipientTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TokenRecipientTransactorSession struct {
	Contract     *TokenRecipientTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// TokenRecipientRaw is an auto generated low-level Go binding around an Ethereum contract.
type TokenRecipientRaw struct {
	Contract *TokenRecipient // Generic contract binding to access the raw methods on
}

// TokenRecipientCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TokenRecipientCallerRaw struct {
	Contract *TokenRecipientCaller // Generic read-only contract binding to access the raw methods on
}

// TokenRecipientTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TokenRecipientTransactorRaw struct {
	Contract *TokenRecipientTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTokenRecipient creates a new instance of TokenRecipient, bound to a specific deployed contract.
func NewTokenRecipient(address common.Address, backend bind.ContractBackend) (*TokenRecipient, error) {
	contract, err := bindTokenRecipient(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TokenRecipient{TokenRecipientCaller: TokenRecipientCaller{contract: contract}, TokenRecipientTransactor: TokenRecipientTransactor{contract: contract}, TokenRecipientFilterer: TokenRecipientFilterer{contract: contract}}, nil
}

// NewTokenRecipientCaller creates a new read-only instance of TokenRecipient, bound to a specific deployed contract.
func NewTokenRecipientCaller(address common.Address, caller bind.ContractCaller) (*TokenRecipientCaller, error) {
	contract, err := bindTokenRecipient(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TokenRecipientCaller{contract: contract}, nil
}

// NewTokenRecipientTransactor creates a new write-only instance of TokenRecipient, bound to a specific deployed contract.
func NewTokenRecipientTransactor(address common.Address, transactor bind.ContractTransactor) (*TokenRecipientTransactor, error) {
	contract, err := bindTokenRecipient(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TokenRecipientTransactor{contract: contract}, nil
}

// NewTokenRecipientFilterer creates a new log filterer instance of TokenRecipient, bound to a specific deployed contract.
func NewTokenRecipientFilterer(address common.Address, filterer bind.ContractFilterer) (*TokenRecipientFilterer, error) {
	contract, err := bindTokenRecipient(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TokenRecipientFilterer{contract: contract}, nil
}

// bindTokenRecipient binds a generic wrapper to an already deployed contract.
func bindTokenRecipient(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenRecipientABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenRecipient *TokenRecipientRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _TokenRecipient.Contract.TokenRecipientCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenRecipient *TokenRecipientRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenRecipient.Contract.TokenRecipientTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenRecipient *TokenRecipientRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenRecipient.Contract.TokenRecipientTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenRecipient *TokenRecipientCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _TokenRecipient.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenRecipient *TokenRecipientTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenRecipient.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenRecipient *TokenRecipientTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenRecipient.Contract.contract.Transact(opts, method, params...)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(_from address, _value uint256, _token address, _extraData bytes) returns()
func (_TokenRecipient *TokenRecipientTransactor) ReceiveApproval(opts *bind.TransactOpts, _from common.Address, _value *big.Int, _token common.Address, _extraData []byte) (*types.Transaction, error) {
	return _TokenRecipient.contract.Transact(opts, "receiveApproval", _from, _value, _token, _extraData)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(_from address, _value uint256, _token address, _extraData bytes) returns()
func (_TokenRecipient *TokenRecipientSession) ReceiveApproval(_from common.Address, _value *big.Int, _token common.Address, _extraData []byte) (*types.Transaction, error) {
	return _TokenRecipient.Contract.ReceiveApproval(&_TokenRecipient.TransactOpts, _from, _value, _token, _extraData)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(_from address, _value uint256, _token address, _extraData bytes) returns()
func (_TokenRecipient *TokenRecipientTransactorSession) ReceiveApproval(_from common.Address, _value *big.Int, _token common.Address, _extraData []byte) (*types.Transaction, error) {
	return _TokenRecipient.Contract.ReceiveApproval(&_TokenRecipient.TransactOpts, _from, _value, _token, _extraData)
}
