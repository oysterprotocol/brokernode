package jobs

import (
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
	"github.com/pkg/errors"
	"gopkg.in/segmentio/analytics-go.v3"
	"math/big"
	"time"
)

/* CheckBetaPayments triggers the methods associated with checking beta payments */
func CheckBetaPayments(thresholdDuration time.Duration, PrometheusWrapper services.PrometheusService) {
	start := PrometheusWrapper.TimeNow()
	defer PrometheusWrapper.HistogramSeconds(PrometheusWrapper.HistogramCheckBetaPayments, start)

	CheckPaymentToBeta()

	HandleErrorTransactionsIfAlpha()

	HandleTimedOutTransactionsIfAlpha(thresholdDuration)
	/* make beta wait 3x as long as alpha so alpha has some chances to retry */
	HandleTimedOutBetaPaymentIfBeta(thresholdDuration * time.Duration(3))

	StartAndCheckGasReclaims()

	PurgeCompletedTransactions()
}

/* CheckPaymentToBeta checks whether the payment to the beta address has arrived */
func CheckPaymentToBeta() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending})

	for _, brokerTx := range brokerTxs {
		balance := EthWrapper.CheckPRLBalance(services.StringToAddress(brokerTx.ETHAddrBeta))
		expectedBalance := new(big.Int).Quo(brokerTx.GetTotalCostInWei(), big.NewInt(int64(2)))
		if balance.Int64() > 0 && balance.Int64() >= expectedBalance.Int64() {
			previousBetaPaymentStatus := brokerTx.PaymentStatus
			brokerTx.PaymentStatus = models.BrokerTxBetaPaymentConfirmed
			err := models.DB.Save(&brokerTx)
			if err != nil {
				oyster_utils.LogIfError(err, nil)
				brokerTx.PaymentStatus = previousBetaPaymentStatus
				continue
			}
			if brokerTx.Type == models.SessionTypeBeta {
				ReportGoodAlphaToDRS(brokerTx)
			}
			oyster_utils.LogToSegment("check_beta_payments: CheckPaymentToBeta - beta_confirmed",
				analytics.NewProperties().
					Set("beta_address", brokerTx.ETHAddrBeta).
					Set("alpha_address", brokerTx.ETHAddrAlpha))
		}
	}
}

/* HandleTimedOutBetaPaymentIfBeta would wrap calls that would report the alpha broker to the DRS */
func HandleTimedOutBetaPaymentIfBeta(thresholdDuration time.Duration) {
	thresholdTime := time.Now().Add(thresholdDuration)
	brokerTxs, _ := models.GetTransactionsBySessionTypesPaymentStatusesAndTime([]int{models.SessionTypeBeta},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentPending}, thresholdTime)

	for _, brokerTx := range brokerTxs {
		ReportBadAlphaToDRS(brokerTx)
	}
}

/* ReportBadAlphaToDRS is a stub of the method we will eventually use for beta to report alpha to the DRS */
func ReportBadAlphaToDRS(brokerTx models.BrokerBrokerTransaction) {
	/* TODO:  DRS reporting logic
	For now just send error to sentry and delete the transaction */

	if brokerTx.Type != models.SessionTypeBeta {
		return
	}

	fmt.Println("Alpha FAILED to send the prl to beta!")

	err := errors.New("alpha did not send PRL to beta for transaction with alpha address: " +
		brokerTx.ETHAddrAlpha + " and beta address: " +
		brokerTx.ETHAddrBeta)
	oyster_utils.LogIfError(err, nil)

	oyster_utils.LogToSegment("check_beta_payments: ReportBadAlphaToDRS - beta_payment_expired",
		analytics.NewProperties().
			Set("beta_address", brokerTx.ETHAddrBeta).
			Set("alpha_address", brokerTx.ETHAddrAlpha))

	models.DB.RawQuery("DELETE from broker_broker_transactions WHERE eth_addr_alpha = ? AND "+
		"eth_addr_beta = ?", brokerTx.ETHAddrAlpha, brokerTx.ETHAddrBeta).All(&[]models.BrokerBrokerTransaction{})
}

/* ReportGoodAlphaToDRS is a stub of the method we will eventually use for beta to report alpha to the DRS */
func ReportGoodAlphaToDRS(brokerTx models.BrokerBrokerTransaction) {
	/* TODO:  DRS is not ready yet but this is where beta would report that alpha
	sent the PRL */

	if brokerTx.Type != models.SessionTypeBeta {
		return
	}

	fmt.Println("Alpha sent the prl to beta!")

	oyster_utils.LogToSegment("check_beta_payments: ReportGoodAlphaToDRS - beta_payment_received",
		analytics.NewProperties().
			Set("beta_address", brokerTx.ETHAddrBeta).
			Set("alpha_address", brokerTx.ETHAddrAlpha))
}

/*StartAndCheckGasReclaims checks current gas reclaims and starts new ones*/
func StartAndCheckGasReclaims() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{models.BrokerTxBetaPaymentConfirmed, models.BrokerTxGasReclaimPending})

	for _, brokerTx := range brokerTxs {
		ethBalance := EthWrapper.CheckETHBalance(services.StringToAddress(brokerTx.ETHAddrAlpha))

		if ethBalance.Int64() > 0 {
			gasNeededToReclaimETH, err := EthWrapper.CalculateGasNeeded(services.GasLimitETHSend)
			if err != nil {
				fmt.Println("Could not calculate gas needed to retrieve ETH from " + brokerTx.ETHAddrAlpha +
					" in StartGasReclaims() in check_beta_payments")
				continue
			}
			if gasNeededToReclaimETH.Int64() > ethBalance.Int64() {
				fmt.Println("Not enough ETH to retrieve leftover ETH from " + brokerTx.ETHAddrAlpha +
					" in StartGasReclaims() in check_beta_payments, setting to success")
				/* won't be able to reclaim whatever is left, just set to confirmed so it will get purged*/
				brokerTx.PaymentStatus = models.BrokerTxGasReclaimConfirmed
				models.DB.ValidateAndUpdate(&brokerTx)
				continue
			}
			if brokerTx.PaymentStatus == models.BrokerTxGasReclaimPending {
				/* Gas payment still in progress, no need to send again */
				continue
			}

			reclaimGas(&brokerTx, ethBalance, gasNeededToReclaimETH)
		} else {
			fmt.Println("No extra gas to reclaim from " + brokerTx.ETHAddrAlpha + " in StartGasReclaims " +
				"in check_beta_payments")
			brokerTx.PaymentStatus = models.BrokerTxGasReclaimConfirmed
			models.DB.ValidateAndUpdate(&brokerTx)
		}
	}
}

/* reclaimGas attempts to take any leftover ETH from the alpha transaction address */
func reclaimGas(brokerTx *models.BrokerBrokerTransaction, ethBalance *big.Int, gasNeededToReclaimETH *big.Int) {
	gasToReclaim := new(big.Int).Sub(ethBalance, gasNeededToReclaimETH)

	privateKey, err := services.StringToPrivateKey(brokerTx.DecryptEthKey())

	_, _, _, err = EthWrapper.SendETH(
		services.StringToAddress(brokerTx.ETHAddrAlpha),
		privateKey,
		services.MainWalletAddress,
		gasToReclaim)

	if err != nil {
		fmt.Println("Could not reclaim leftover ETH from " + brokerTx.ETHAddrAlpha +
			" in StartGasReclaims in check_beta_payments")
	} else {
		fmt.Println("Reclaiming leftover ETH from " + brokerTx.ETHAddrAlpha + " in StartGasReclaims in " +
			"check_beta_payments")
		brokerTx.PaymentStatus = models.BrokerTxGasReclaimPending
		models.DB.ValidateAndUpdate(brokerTx)
	}
}

/* HandleTimedOutTransactionsIfAlpha simply stages old transactions to be tried again */
func HandleTimedOutTransactionsIfAlpha(thresholdDuration time.Duration) {
	thresholdTime := time.Now().Add(thresholdDuration)
	brokerTxs, _ := models.GetTransactionsBySessionTypesPaymentStatusesAndTime([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{
			models.BrokerTxGasPaymentPending,
			models.BrokerTxBetaPaymentPending,
			models.BrokerTxGasReclaimPending}, thresholdTime)

	for _, brokerTx := range brokerTxs {
		currentStatus := brokerTx.PaymentStatus
		brokerTx.PaymentStatus = models.PaymentStatus(int(currentStatus) - 1)
		err := models.DB.Save(&brokerTx)
		oyster_utils.LogIfError(err, nil)
	}
}

/* HandleErrorTransactionsIfAlpha simply stages error transactions to be tried again */
func HandleErrorTransactionsIfAlpha() {
	brokerTxs, _ := models.GetTransactionsBySessionTypesAndPaymentStatuses([]int{models.SessionTypeAlpha},
		[]models.PaymentStatus{
			models.BrokerTxGasPaymentError,
			models.BrokerTxBetaPaymentError,
			models.BrokerTxGasReclaimError})

	for _, brokerTx := range brokerTxs {
		currentStatus := brokerTx.PaymentStatus
		brokerTx.PaymentStatus = models.PaymentStatus(int(currentStatus) * -1)
		err := models.DB.Save(&brokerTx)
		oyster_utils.LogIfError(err, nil)
	}
}

/* PurgeCompletedTransactions wraps a call which will delete any brokerTxs whose gas has been reclaimed */
func PurgeCompletedTransactions() {
	models.DeleteCompletedBrokerTransactions()
}
