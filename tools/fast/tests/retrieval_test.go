package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ipfs/go-ipfs-files"
	logging "github.com/ipfs/go-log"

	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-filecoin/protocol/storage/storagedeal"
	th "github.com/filecoin-project/go-filecoin/testhelpers"
	"github.com/filecoin-project/go-filecoin/tools/fast"
	"github.com/filecoin-project/go-filecoin/tools/fast/series"
	localplugin "github.com/filecoin-project/go-filecoin/tools/iptb-plugins/filecoin/local"
)

func init() {
	// Enabling debug logging provides a lot of insight into what commands are
	// being executed
	logging.SetDebugLogging()
}

// TestRetrieval exercises storing and retreiving with the filecoin protocols
func TestRetrieval(t *testing.T) {
	// Set the series global sleep delay to 5 seconds
	series.GlobalSleepDelay = time.Second * 5
	var sectorSize int64 = 1016

	// This test should run in 20 block times, with 60 seconds for sealing, and no longer
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(20*series.GlobalSleepDelay).Add(60*time.Second))
	defer cancel()

	require := require.New(t)

	// Create a directory for the test using the test name (mostly for FAST)
	dir, err := ioutil.TempDir("", t.Name())
	require.NoError(err)

	// Create an environment that includes a genesis block with 1MM FIL
	env, err := fast.NewEnvironmentMemoryGenesis(big.NewInt(1000000), dir)
	require.NoError(err)

	// Teardown will shutdown all running processes the environment knows about
	// and cleanup anything the evironment setup. This includes the directory
	// the environment was created to use.
	defer func() {
		require.NoError(env.Teardown(ctx))
	}()

	// Setup options for nodes.
	options := make(map[string]string)
	options[localplugin.AttrLogJSON] = "0"                               // Disable JSON logs
	options[localplugin.AttrLogLevel] = "4"                              // Set log level to Info
	options[localplugin.AttrUseSmallSectors] = "true"                    // Enable small sectors
	options[localplugin.AttrFilecoinBinary] = th.MustGetFilecoinBinary() // Enable small sectors

	genesisURI := env.GenesisCar()
	genesisMiner, err := env.GenesisMiner()
	require.NoError(err)

	fastenvOpts := fast.EnvironmentOpts{
		InitOpts:   []fast.ProcessInitOption{fast.POGenesisFile(genesisURI)},
		DaemonOpts: []fast.ProcessDaemonOption{fast.POBlockTime(series.GlobalSleepDelay)},
	}

	// Setup nodes used for the test
	genesis, err := env.NewProcess(ctx, localplugin.PluginName, options, fastenvOpts)
	require.NoError(err)

	miner, err := env.NewProcess(ctx, localplugin.PluginName, options, fastenvOpts)
	require.NoError(err)

	client, err := env.NewProcess(ctx, localplugin.PluginName, options, fastenvOpts)
	require.NoError(err)

	// Start setting up the nodes
	// Setup Genesis
	err = series.SetupGenesisNode(ctx, genesis, genesisMiner.Address, files.NewReaderFile(genesisMiner.Owner))
	require.NoError(err)

	// Start Miner
	err = series.InitAndStart(ctx, miner)
	require.NoError(err)

	// Start Client
	err = series.InitAndStart(ctx, client)
	require.NoError(err)

	// Connect everything to the genesis node so it can issue filecoin when needed
	err = series.Connect(ctx, genesis, miner)
	require.NoError(err)

	err = series.Connect(ctx, genesis, client)
	require.NoError(err)

	// Everyone needs FIL to deal with gas costs and make sure their wallets
	// exists (sending FIL to a wallet addr creates it)
	err = series.SendFilecoinDefaults(ctx, genesis, miner, 1000)
	require.NoError(err)

	err = series.SendFilecoinDefaults(ctx, genesis, client, 1000)
	require.NoError(err)

	RunRetrievalTest(ctx, t, miner, client, sectorSize)
}

// TestRetrieval exercises storing and retreiving with the filecoin protocols
func TestNetworkRetrieval(t *testing.T) {
	t.SkipNow()

	// Set the series global sleep delay to 30 seconds
	series.GlobalSleepDelay = time.Second * 30
	var sectorSize int64 = 266338304

	// This test should run in and hour and no longer
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add((60*60)*time.Second))
	defer cancel()

	require := require.New(t)

	// Create a directory for the test using the test name (mostly for FAST)
	dir, err := ioutil.TempDir("", t.Name())
	require.NoError(err)

	// Create an environment that includes a genesis block with 1MM FIL
	env, err := fast.NewEnvironmentDevnet("nightly", dir)
	require.NoError(err)

	// Teardown will shutdown all running processes the environment knows about
	// and cleanup anything the evironment setup. This includes the directory
	// the environment was created to use.
	defer env.Teardown(ctx)

	// Setup options for nodes.
	options := make(map[string]string)
	options[localplugin.AttrLogJSON] = "0"                               // Disable JSON logs
	options[localplugin.AttrLogLevel] = "4"                              // Set log level to Info
	options[localplugin.AttrUseSmallSectors] = "false"                   // Enable small sectors
	options[localplugin.AttrFilecoinBinary] = th.MustGetFilecoinBinary() // Enable small sectors

	genesisURI := env.GenesisCar()

	fastenvOpts := fast.EnvironmentOpts{
		InitOpts:   []fast.ProcessInitOption{fast.POGenesisFile(genesisURI), fast.PODevnetTest()},
		DaemonOpts: []fast.ProcessDaemonOption{},
	}

	miner, err := env.NewProcess(ctx, localplugin.PluginName, options, fastenvOpts)
	require.NoError(err)

	client, err := env.NewProcess(ctx, localplugin.PluginName, options, fastenvOpts)
	require.NoError(err)

	// Start Miner
	err = series.InitAndStart(ctx, miner)
	require.NoError(err)

	// Start Client
	err = series.InitAndStart(ctx, client)
	require.NoError(err)

	// Everyone needs FIL to deal with gas costs and make sure their wallets
	// exists (sending FIL to a wallet addr creates it)
	err = fast.GetFunds(ctx, env, miner)
	require.NoError(err)

	err = fast.GetFunds(ctx, env, client)
	require.NoError(err)

	RunRetrievalTest(ctx, t, miner, client, sectorSize)
}

func RunRetrievalTest(ctx context.Context, t *testing.T, miner, client *fast.Filecoin, sectorSize int64) {
	require := require.New(t)

	pledge := uint64(10)                    // sectors
	collateral := big.NewInt(int64(pledge)) // FIL
	price := big.NewFloat(0.000000001)      // price per byte/block
	expiry := big.NewInt(24 * 60 * 60 / 30) // ~24 hours

	// Create a miner on the miner node
	ask, err := series.CreateMinerWithAsk(ctx, miner, pledge, collateral, price, expiry)
	require.NoError(err)

	// Connect the client and the miner
	err = series.Connect(ctx, client, miner)
	require.NoError(err)

	// Store some data with the miner with the given ask, returns the cid for
	// the imported data, and the deal which was created
	var data bytes.Buffer
	dataReader := io.LimitReader(rand.Reader, sectorSize)
	dataReader = io.TeeReader(dataReader, &data)
	dcid, deal, err := series.ImportAndStore(ctx, client, ask, files.NewReaderFile(dataReader))
	require.NoError(err)

	// Wait for the deal to be posted
	err = series.WaitForDealState(ctx, client, deal, storagedeal.Posted)
	require.NoError(err)

	// Retrieve the stored piece of data
	reader, err := client.RetrievalClientRetrievePiece(ctx, dcid, ask.Miner)
	require.NoError(err)

	// Verify that it's all the same
	retrievedData, err := ioutil.ReadAll(reader)
	require.NoError(err)
	require.Equal(data.Bytes(), retrievedData)
}
