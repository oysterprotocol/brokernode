package jobs

import (
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/oysterprotocol/brokernode/utils/eth_gateway"
	"gopkg.in/segmentio/analytics-go.v3"
	"math/big"
)

/* CheckAlphaPayments handles the operations around checking the status of payments to alpha and
initiating the payment to beta */
func CheckAlphaPayments(PrometheusWrapper services.PrometheusService) {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramCheckAlphaPayments, start)

	CheckPaymentToAlpha()
	SendGasToAlphaTransactionAddress()
	CheckGasPayments()
	SendPaymentToBeta()
}

/* CheckPaymentToAlpha checks whether the PRL payment has arrived to alpha, and if this is true and the
session is a beta session, it will set the beta payment status to pending */
func CheckPaymentToAlpha() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{},
		[]models.PaymentStatus{models.BrokerTxAlphaPaymentPending})

	for _, brokerTx := range brokerTxs {
		balance := EthWrapper.CheckPRLBalance(eth_gateway.StringToAddress(brokerTx.ETHAddrAlpha))
		if balance.Int64() > 0 && balance.Int64() >= brokerTx.GetTotalCostInWei().Int64() {
			previousPaymentStatus := brokerTx.PaymentStatus

			if brokerTx.Type == models.SessionTypeAlpha {
				brokerTx.PaymentStatus = models.BrokerTxAlphaPaymentConfirmed
			} else {
				/* A beta broker does not care about the gas transfer, only that it ultimately receives its share of
				the PRL.  So once beta sees that the alpha payment has arrived it just starts waiting for its PRL. */
				brokerTx.PaymentStatus = models.BrokerTxBetaPaymentPending
			}
			err := models.DB.Save(&brokerTx)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				brokerTx.PaymentStatus = previousPaymentStatus
				continue
			}

			models.SetUploadSessionToPaid(brokerTx)
			oyster_utils.LogToSegment("check_alpha_payments: CheckPaymentToAlpha - alpha_confirmed",
				analytics.NewProperties().
					Set("beta_address", brokerTx.ETHAddrBeta).
					Set("alpha_address", brokerTx.ETHAddrAlpha))
		}
	}
}

/* SendGasToAlphaTransactionAddress gets the transactions for which the alpha address has received payment but the
gas has not been sent, and initiates sending the gas */
func SendGasToAlphaTransactionAddress() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxAlphaPaymentConfirmed})

	for _, brokerTx := range brokerTxs {
		hasEnoughGas, gasToSend, err := addressHasEnoughGas(brokerTx.ETHAddrAlpha)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		if hasEnoughGas {
			brokerTx.PaymentStatus = models.BrokerTxGasPaymentConfirmed
			err = models.DB.Save(&brokerTx)
			continue
		}

		_, _, _, err = EthWrapper.SendETH(
			eth_gateway.MainWalletAddress,
			eth_gateway.MainWalletPrivateKey,
			eth_gateway.StringToAddress(brokerTx.ETHAddrAlpha),
			gasToSend)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		previousPaymentStatus := brokerTx.PaymentStatus
		brokerTx.PaymentStatus = models.BrokerTxGasPaymentPending

		err = models.DB.Save(&brokerTx)
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			brokerTx.PaymentStatus = previousPaymentStatus
			continue
		}
	}
}

/* CheckGasPayments checks the status of gas payments to the alpha
transaction address that are currently in progress */
func CheckGasPayments() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxGasPaymentPending})

	for _, brokerTx := range brokerTxs {

		hasEnoughGas, _, err := addressHasEnoughGas(brokerTx.ETHAddrAlpha)

		if err != nil {
			oyster_utils.LogIfError(err, nil)
			continue
		}

		if hasEnoughGas {
			brokerTx.PaymentStatus = models.BrokerTxGasPaymentConfirmed
			err = models.DB.Save(&brokerTx)
			continue
		}
	}
}

/* addressHasEnoughGas will be called on the alpha address to determine if it has enough
gas to send the PRL to the beta address */
func addressHasEnoughGas(address string) (bool, *big.Int, error) {
	gasBalance := EthWrapper.CheckETHBalance(eth_gateway.StringToAddress(address))

	gasNeeded, err := EthWrapper.CalculateGasNeeded(eth_gateway.GasLimitPRLSend)
	if err != nil {
		oyster_utils.LogIfError(err, nil)
		return false, big.NewInt(0), err
	}

	gasToSend := new(big.Int).Sub(gasNeeded, gasBalance)

	if gasToSend.Int64() <= 0 {
		return true, big.NewInt(0), nil
	}
	return false, gasToSend, nil
}

/* SendPaymentToBeta gets the transactions for which the alpha address has received payment but the beta address
has not, and calls a method on those transactions to start the beta payment */
func SendPaymentToBeta() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxGasPaymentConfirmed})

	for _, brokerTx := range brokerTxs {
		balance := EthWrapper.CheckPRLBalance(eth_gateway.StringToAddress(brokerTx.ETHAddrAlpha))
		checkAndSendHalfPrlToBeta(brokerTx, balance)
	}
}

/* checkAndSendHalfPrlToBeta checks whether beta has already received the transaction, and
if not, sends it half the PRL and marks beta payment status as pending */
func checkAndSendHalfPrlToBeta(brokerTx models.BrokerBrokerTransaction, balance *big.Int) {
	if brokerTx.Type != models.SessionTypeAlpha ||
		brokerTx.PaymentStatus != models.BrokerTxGasPaymentConfirmed ||
		brokerTx.ETHAddrBeta == "" {
		return
	}

	betaAddr := eth_gateway.StringToAddress(brokerTx.ETHAddrBeta)
	betaBalance := EthWrapper.CheckPRLBalance(betaAddr)
	if betaBalance.Int64() > 0 {
		brokerTx.PaymentStatus = models.BrokerTxBetaPaymentConfirmed
		err := models.DB.Save(&brokerTx)
		oyster_utils.LogIfError(err, nil)
		return
	}

	var splitAmount big.Int
	splitAmount.Set(balance)
	splitAmount.Div(balance, big.NewInt(2))

	privateKey, err := eth_gateway.StringToPrivateKey(brokerTx.DecryptEthKey())
	if err != nil {
		oyster_utils.LogIfError(err, nil)
	}

	callMsg, _ := EthWrapper.CreateSendPRLMessage(
		eth_gateway.StringToAddress(brokerTx.ETHAddrAlpha),
		privateKey,
		eth_gateway.StringToAddress(brokerTx.ETHAddrBeta), splitAmount)

	sendSuccess, _, _ := EthWrapper.SendPRLFromOyster(callMsg)

	if sendSuccess {
		brokerTx.PaymentStatus = models.BrokerTxBetaPaymentPending
		err := models.DB.Save(&brokerTx)
		oyster_utils.LogIfError(err, nil)

		oyster_utils.LogToSegment("check_alpha_payments: CheckAndSendHalfPrlToBeta - beta_transaction_started",
			analytics.NewProperties().
				Set("beta_address", brokerTx.ETHAddrBeta).
				Set("alpha_address", brokerTx.ETHAddrAlpha))
	}
}
