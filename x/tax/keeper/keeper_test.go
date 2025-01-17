package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	nolusapp "github.com/Nolus-Protocol/nolus-core/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestAccount represents a client Account that can be used in unit tests.
type TestAccount struct {
	acc  authtypes.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type KeeperTestSuite struct {
	suite.Suite

	app         *nolusapp.App
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder
	anteHandler sdk.AnteHandler
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	tempDir := s.T().TempDir()
	s.app, s.ctx = nolusapp.CreateTestApp(isCheckTx, tempDir)
	s.ctx = s.ctx.WithBlockHeight(1)

	// set up TxConfig
	encodingConfig := simapp.MakeTestEncodingConfig()
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	s.Require().NoError(s.txBuilder.SetMsgs([]sdk.Msg{}...))

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   s.app.AccountKeeper,
			BankKeeper:      s.app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)
	s.Require().NoError(err)

	s.anteHandler = anteHandler
}

// CreateTestAccounts creates accounts.
func (s *KeeperTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount
	for i := 0; i < numAccs; i++ {
		priv, _, addr := sdktestutil.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestSuite) FundAcc(addr sdk.AccAddress, amounts sdk.Coins) {
	err := simapp.FundAccount(s.app.BankKeeper, s.ctx, addr, amounts)
	s.Require().NoError(err)
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *KeeperTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			s.txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
