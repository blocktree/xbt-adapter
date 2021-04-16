/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwtester

import (
	"testing"
	"time"

	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
)

func testGetAssetsAccountBalance(tm *openw.WalletManager, walletID, accountID string) {
	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance)
}

func testGetAssetsAccountTokenBalance(tm *openw.WalletManager, walletID, accountID string, contract openwallet.SmartContract) {
	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("token balance:", balance.Balance)
}

func testCreateTransactionStep(tm *openw.WalletManager, walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	//err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	//if err != nil {
	//	log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	//	return nil, err
	//}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "test", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testCreateSummaryTransactionStep(
	tm *openw.WalletManager,
	walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract) ([]*openwallet.RawTransactionWithError, error) {

	rawTxArray, err := tm.CreateSummaryRawTransactionWithError(testApp, walletID, accountID, summaryAddress, minTransfer,
		retainedBalance, feeRate, start, limit, contract, nil)

	if err != nil {
		log.Error("CreateSummaryTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTxArray, nil
}

func testSignTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	log.Info("wait sign message : ", rawTx.Signatures[rawTx.Account.AccountID][0].Message)

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

/*
withdraw
wallet : VzoFQEUGHmFyxFLEwGyisVH1zFe87uGJDm
account : e6Z3XLX7dhUALhvL9rdmwRe7Z1NSYWddSG85ubRrbQy
1 address : xB029b2bc3302ddaF67953bF98F0C88EEFde7e5e9D

charge
wallet : WAii8ta4Tz1JhaVY29mvsm62vx9iygCrXd
account : H33u8H1s3iMsi4R1ENNYHJSjfARJXuzXidJV81jQ3ezh
1 address : xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B
*/
func TestTransfer(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "VzoFQEUGHmFyxFLEwGyisVH1zFe87uGJDm"
	accountID := "e6Z3XLX7dhUALhvL9rdmwRe7Z1NSYWddSG85ubRrbQy"
	to := "xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.1234", "", nil)
	if err != nil {
		return
	}

	log.Std.Info("rawTx: %+v", rawTx)

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}
}

/*
withdraw
wallet : WAii8ta4Tz1JhaVY29mvsm62vx9iygCrXd
account : 84hkCfouuL6HARaECgjS8RozovSTaZNgzeWYvToC7rhy
1 address : 16M6KBKqUFp5wyNeeG3DkmFZ951FAN2rA6z3wGHvNYfCyKU5
*/
func TestBatchTransfer(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "VzoFQEUGHmFyxFLEwGyisVH1zFe87uGJDm"
	accountID := "e6Z3XLX7dhUALhvL9rdmwRe7Z1NSYWddSG85ubRrbQy"

	toArr := make([]string, 0)
	toArr = append(toArr, "xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B")
	toArr = append(toArr, "xB1ee6747285807E072A335691ccf068Ae5e99E82A")
	toArr = append(toArr, "xB2A1285eDE7c98F9C7B2f8c09fe5E52bF636E33bB")
	toArr = append(toArr, "xB397Acc663f0315f0230827946f7f62f285325181")

	amountArr := make([]string, 0)
	amountArr = append(amountArr, "0.12341")
	amountArr = append(amountArr, "0.12342")
	amountArr = append(amountArr, "0.12343")
	amountArr = append(amountArr, "0.12344")

	for i := 0; i < len(amountArr); i++{
		to := toArr[i]
		amount := amountArr[i]
		testGetAssetsAccountBalance(tm, walletID, accountID)

		rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, amount, "", nil)
		if err != nil {
			return
		}

		log.Std.Info("rawTx: %+v", rawTx)

		_, err = testSignTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}

/*
charge
wallet : WAii8ta4Tz1JhaVY29mvsm62vx9iygCrXd
account : H33u8H1s3iMsi4R1ENNYHJSjfARJXuzXidJV81jQ3ezh
1 address : xB52c55E62d708CdE25Cec9B576F5bFDEcFB5C328B
*/
func TestSummary(t *testing.T) {
	tm := testInitWalletManager()

	walletID := "WAii8ta4Tz1JhaVY29mvsm62vx9iygCrXd"
	accountID := "H33u8H1s3iMsi4R1ENNYHJSjfARJXuzXidJV81jQ3ezh"
	summaryAddress := "xB029b2bc3302ddaF67953bF98F0C88EEFde7e5e9D"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	rawTxArray, err := testCreateSummaryTransactionStep(tm, walletID, accountID,
		summaryAddress, "0.11", "0", "",
		0, 100, nil)
	if err != nil {
		log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
		return
	}

	//执行汇总交易
	for _, rawTxWithErr := range rawTxArray {

		if rawTxWithErr.Error != nil {
			log.Error(rawTxWithErr.Error.Error())
			continue
		}

		_, err = testSignTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}
	}

}
