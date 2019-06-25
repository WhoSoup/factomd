// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factomd/config"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/wsapi"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

type FactomNode struct {
	Index    int
	State    *state.State
	Peers    []interfaces.IPeer
	MLog     *MsgLog
	P2PIndex int
}

var fnodes []*FactomNode

var networkpattern string
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var p2pNetwork *p2p.Controller
var logPort int

func GetFnodes() []*FactomNode {
	return fnodes
}

func init() {
	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General
}

func NetStart(s *state.State, cfg config.Config) {

	s.PortNumber = cfg.Factomd.APIPort
	s.ControlPanelPort = cfg.Factomd.ControlPanelPort
	logPort = cfg.Factomd.PprofPort
	messages.AckBalanceHash = !cfg.Factomd.NoBalanceHash

	// Must add the prefix before loading the configuration.
	s.AddPrefix(cfg.Factomd.DBSlug)

	var networkID p2p.NetworkID
	s.Network = cfg.Factomd.Network
	switch cfg.Factomd.Network {
	case "MAIN":
		s.DirectoryBlockInSeconds = 600
		networkID = p2p.MainNet
	case "LOCAL":
		fmt.Println("Running on the local network, use local coinbase constants")
		constants.SetLocalCoinBaseConstants()
		networkID = p2p.LocalNet
	case "TEST":
		networkID = p2p.TestNet
	default: // custom
		s.Network = "CUSTOM"
		s.CustomNetworkID = primitives.Sha([]byte(cfg.Factomd.Network)).Bytes()[:4]
		networkID = p2p.NetworkID(binary.BigEndian.Uint32(s.CustomNetworkID))
		fmt.Println("Running on the custom network, use custom coinbase constants")
		constants.SetCustomCoinBaseConstants()
	}

	globals.Params.NetworkName = s.Network

	s.LogPath = cfg.SlugPath(cfg.Factomd.LogPath)
	s.LdbPath = cfg.SlugPath(cfg.Factomd.DBLdbPath)
	s.BoltDBPath = cfg.SlugPath(cfg.Factomd.DBBoltPath)

	s.LogLevel = strings.ToLower(cfg.Factomd.LogLevel)
	if cfg.Factomd.ForceFollower {
		s.NodeMode = "FULL"
	} else {
		s.NodeMode = "SERVER"
	}

	s.DBType = cfg.Factomd.DBType
	s.ExportData = cfg.Factomd.DBExportData // bool
	s.ExportDataSubpath = cfg.Factomd.DBExportDataPath

	//s.MainNetworkPort = cfg.Factomd.MainNetworkPort
	s.PeersFile = cfg.Factomd.P2PPeerFileSuffix
	//s.MainSeedURL = cfg.Factomd.MainSeedURL
	//s.MainSpecialPeers = cfg.Factomd.MainSpecialPeers
	//s.TestNetworkPort = cfg.Factomd.TestNetworkPort
	//s.TestSeedURL = cfg.Factomd.TestSeedURL
	//s.TestSpecialPeers = cfg.Factomd.TestSpecialPeers
	s.CustomBootstrapIdentity = cfg.Factomd.BootstrapIdentity
	s.CustomBootstrapKey = cfg.Factomd.BootstrapKey
	//s.LocalNetworkPort = cfg.Factomd.LocalNetworkPort
	//s.LocalSeedURL = cfg.Factomd.LocalSeedURL
	//s.LocalSpecialPeers = cfg.Factomd.LocalSpecialPeers
	s.LocalServerPrivKey = cfg.Factomd.IdentityPrivateKey
	//s.CustomNetworkPort = cfg.Factomd.CustomNetworkPort
	//s.CustomSeedURL = cfg.Factomd.CustomSeedURL
	//s.CustomSpecialPeers = cfg.Factomd.CustomSpecialPeers
	//.FactoshisPerEC = cfg.Factomd.ExchangeRate
	s.DirectoryBlockInSeconds = cfg.Factomd.BlockTime
	s.PortNumber = cfg.Factomd.APIPort
	s.ControlPanelPort = cfg.Factomd.ControlPanelPort
	s.RpcUser = cfg.Factomd.WebUsername
	s.RpcPass = cfg.Factomd.WebPassword
	s.StateSaverStruct.FastBoot = !cfg.Factomd.DBNoFastBoot
	//s.StateSaverStruct.FastBootLocation = cfg.Factomd.FastBootLocation
	s.FastBoot = !cfg.Factomd.DBNoFastBoot
	//s.FastBootLocation = cfg.Factomd.FastBootLocation

	// to test run curl -H "Origin: http://anotherexample.com" -H "Access-Control-Request-Method: POST" /
	//     -H "Access-Control-Request-Headers: X-Requested-With" -X POST /
	//     --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "heights"}' -H 'content-type:text/plain;'  /
	//     --verbose http://localhost:8088/v2

	// while the config file has http://anotherexample.com in parameter CorsDomains the response should contain the string
	// < Access-Control-Allow-Origin: http://anotherexample.com

	if len(cfg.Factomd.WebCORS) > 0 {
		domains := strings.Split(cfg.Factomd.WebCORS, ",")
		s.CorsDomains = make([]string, len(domains))
		for _, domain := range domains {
			s.CorsDomains = append(s.CorsDomains, strings.Trim(domain, " "))
		}
	}
	s.FactomdTLSEnable = cfg.Factomd.WebTLS

	cert := cfg.Factomd.WebTLSCertificate
	if cert == "" {
		cert = cfg.HomePath("factomdAPIpub.cert")
	}

	key := cfg.Factomd.WebTLSKey
	if key == "" {
		key = cfg.HomePath("factomdAPIpriv.key")
	}

	s.SetTLSCertificate(cert, key)

	externalIP := strings.Split(cfg.Walletd.FactomdLocation, ":")[0]
	if externalIP != "localhost" {
		s.FactomdLocations = externalIP
	}

	switch cfg.Factomd.ControlPanel {
	case "DISABLED":
		s.ControlPanelSetting = 0
	case "READWRITE":
		s.ControlPanelSetting = 2
	case "READONLY":
		fallthrough
	default:
		s.ControlPanelSetting = 1
	}

	s.FERChainId = cfg.Factomd.OracleChain
	s.ExchangeRateAuthorityPublicKey = cfg.Factomd.OraclePublicKey

	identity, err := primitives.HexToHash(cfg.Factomd.IdentityChain)
	if err != nil {
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
		s.LogPrintf("AckChange", "Bad IdentityChainID  in config \"%v\"", cfg.Factomd.IdentityChain)
		s.LogPrintf("AckChange", "Default2 IdentityChainID \"%v\"", s.IdentityChainID.String())
	} else {
		s.IdentityChainID = identity
		s.LogPrintf("AckChange", "Load IdentityChainID \"%v\"", s.IdentityChainID.String())
	}

	s.JournalFile = s.LogPath + "/journal0" + ".log"

	s.OneLeader = cfg.Factomd.OneLeader
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(cfg.Factomd.SimTimeOffset) * 1000)
	s.StartDelayLimit = int64(cfg.Factomd.StartDelay) * 1000
	s.Journaling = cfg.Factomd.JournalFile != ""
	s.FactomdVersion = FactomdVersion
	s.EFactory = new(electionMsgs.ElectionsFactory)

	log.SetOutput(os.Stdout)
	switch strings.ToLower(cfg.Factomd.LogLevel) {
	case "none":
		log.SetOutput(ioutil.Discard)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	}

	if cfg.Factomd.LogJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Set the wait for entries flag
	s.WaitForEntries = cfg.Factomd.SimWait

	s.FaultTimeout = 9999999 //todo: Old Fault Mechanism -- remove

	s.FastSaveRate = cfg.Factomd.DBFastBootRate

	s.CheckChainHeads.CheckChainHeads = (cfg.Factomd.ChainHeadFix != "OFF")
	s.CheckChainHeads.Fix = (cfg.Factomd.ChainHeadFix == "ON")

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", cfg.Factomd.SimFocus)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		if !cfg.Factomd.P2PDisable {
			p2pNetwork.NetworkStop()
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	if cfg.Factomd.JournalFile != "" && cfg.Factomd.DBType != "MAP" {
		cfg.Factomd.DBType = "MAP"
	}

	if cfg.Factomd.ForceFollower {
		s.NodeMode = "FULL"
		leadID := primitives.Sha([]byte(s.Prefix + "FNode0"))
		if s.IdentityChainID.IsSameAs(leadID) {
			s.SetIdentityChainID(primitives.Sha([]byte(time.Now().String()))) // Make sure this node is NOT a leader
		}
	}

	s.KeepMismatch = cfg.Factomd.KeepMismatch

	s.UseLogstash = cfg.Factomd.LogLogstash != ""
	s.LogstashURL = cfg.Factomd.LogLogstash

	go StartProfiler(cfg.Factomd.PprofMPR, cfg.Factomd.PprofExpose)

	s.AddPrefix(cfg.Factomd.DBSlug)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(cfg.Factomd.SimDropRate)

	if cfg.Factomd.ForceSync2Height >= 0 {
		s.EntryDBHeightComplete = uint32(cfg.Factomd.ForceSync2Height)
		s.LogPrintf("EntrySync", "NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
	} else {
		height, err := s.DB.FetchDatabaseEntryHeight()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("ERROR: %v", err))
		} else {
			s.EntryDBHeightComplete = height
			s.LogPrintf("EntrySync", "NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
		}
	}

	mLog.Init(cfg.Factomd.SimRuntimeLog, cfg.Factomd.SimCount)

	setupFirstAuthority(s)

	/*os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Build", Build))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Node name", s.NodeName))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "balancehash", messages.AckBalanceHash))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "FNode 0 Salt", s.Salt.String()[:16]))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", p.EnableNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "waitentries", p.WaitEntries))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", p.ListenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", p.Prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", p.Cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "FastSaveRate", p.FastSaveRate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net spec", pnet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", p.DropRate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", p.Journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", p.Db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", p.CloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", p.Peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", p.Exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive_in", p.ExclusiveIn))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", p.BlkTime))
	//os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", p.FaultTimeout)) // TODO old fault timeout mechanism to be removed
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", p.RuntimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", p.Rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", p.TimeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", p.KeepMismatch))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "startDelay", p.StartDelay))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	os.Stderr.WriteString(fmt.Sprintf("%20s %x (%s)\n", "customnet", p.CustomNet, p.CustomNetName))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "deadline (ms)", p.Deadline))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "tls", s.FactomdTLSEnable))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "selfaddr", s.FactomdLocations))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "rpcuser", s.RpcUser))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "corsdomains", s.CorsDomains))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Start 2nd Sync at ht", s.EntryDBHeightComplete))

	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", elections.FaultTimeout))*/

	if "" == s.RpcPass {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is blank"))
	} else {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is set"))
	}
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "TCP port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "pprof port", logPort))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "Control Panel port", s.ControlPanelPort))

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make p.cnt Factom nodes
	for i := 0; i < cfg.Factomd.SimCount; i++ {
		makeServer(s) // We clone s to make all of our servers
	}
	// Modify Identities of new nodes
	if len(fnodes) > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Setup the Skeleton Identity & Registration
	for i := range fnodes {
		fnodes[i].State.IntiateNetworkSkeletonIdentity()
		fnodes[i].State.InitiateNetworkIdentityRegistration()
	}

	// Start the P2P network

	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
	p2p.NetworkDeadline = time.Duration(cfg.Factomd.P2PTimeout) * time.Millisecond

	if !cfg.Factomd.P2PDisable {

		nodeName := fnodes[0].State.FactomNodeName
		ci := p2p.ControllerInit{
			NodeName:                 nodeName,
			Port:                     string(cfg.Factomd.P2PPort),
			PeersFile:                s.PeersFile,
			Network:                  networkID,
			Exclusive:                cfg.Factomd.P2PConnectionPolicy != "NORMAL",
			ExclusiveIn:              cfg.Factomd.P2PConnectionPolicy == "REFUSE",
			SeedURL:                  cfg.Factomd.P2PSeed,
			ConfigPeers:              cfg.Factomd.P2PSpecialPeers,
			CmdLinePeers:             "",
			ConnectionMetricsChannel: connectionMetricsChannel,
		}
		p2pNetwork = new(p2p.Controller).Init(ci)
		fnodes[0].State.NetworkController = p2pNetwork
		p2pNetwork.StartNetwork()
		p2pProxy = new(P2PProxy).Init(nodeName, "P2P Network").(*P2PProxy)
		p2pProxy.FromNetwork = p2pNetwork.FromNetwork
		p2pProxy.ToNetwork = p2pNetwork.ToNetwork

		fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
		p2pProxy.StartProxy()

		go networkHousekeeping() // This goroutine executes once a second to keep the proxy apprised of the network status.
	}

	networkpattern = strings.ToLower(cfg.Factomd.SimNet)

	switch networkpattern {
	case "file":
		file, err := os.Open(cfg.Factomd.SimNetFile)
		if err != nil {
			panic(fmt.Sprintf("File network.txt failed to open: %s", err.Error()))
		} else if file == nil {
			panic(fmt.Sprint("File network.txt failed to open, and we got a file of <nil>"))
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var a, b int
			var s string
			fmt.Sscanf(scanner.Text(), "%d %s %d", &a, &s, &b)
			if s == "--" {
				AddSimPeer(fnodes, a, b)
			}
		}
	case "square":
		side := int(math.Sqrt(float64(cfg.Factomd.SimCount)))

		for i := 0; i < side; i++ {
			AddSimPeer(fnodes, i*side, (i+1)*side-1)
			AddSimPeer(fnodes, i, side*(side-1)+i)
			for j := 0; j < side; j++ {
				if j < side-1 {
					AddSimPeer(fnodes, i*side+j, i*side+j+1)
				}
				AddSimPeer(fnodes, i*side+j, ((i+1)*side)+j)
			}
		}
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < cfg.Factomd.SimCount; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		// Make long into a circle
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < cfg.Factomd.SimCount; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < cfg.Factomd.SimCount; i += 17 {
			AddSimPeer(fnodes, i%cfg.Factomd.SimCount, (i+5)%cfg.Factomd.SimCount)
		}
		for i := 0; (i+13)*2 < cfg.Factomd.SimCount; i += 13 {
			AddSimPeer(fnodes, i%cfg.Factomd.SimCount, (i+7)%cfg.Factomd.SimCount)
		}
	case "alot":
		n := len(fnodes)
		for i := 0; i < n; i++ {
			AddSimPeer(fnodes, i, (i+1)%n)
			AddSimPeer(fnodes, i, (i+5)%n)
			AddSimPeer(fnodes, i, (i+7)%n)
		}

	case "alot+":
		n := len(fnodes)
		for i := 0; i < n; i++ {
			AddSimPeer(fnodes, i, (i+1)%n)
			AddSimPeer(fnodes, i, (i+5)%n)
			AddSimPeer(fnodes, i, (i+7)%n)
			AddSimPeer(fnodes, i, (i+13)%n)
		}

	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddSimPeer(fnodes, index, row)
				AddSimPeer(fnodes, index, row+1)
				row++
				index++
				if index >= len(fnodes) {
					break treeloop
				}
			}
			row += 1
		}
	case "circles":
		circleSize := 7
		index := 0
		for {
			AddSimPeer(fnodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddSimPeer(fnodes, i, i+1)
			}
			index += circleSize

			AddSimPeer(fnodes, index, index-circleSize/3)
			AddSimPeer(fnodes, index+2, index-circleSize-circleSize*2/3-1)
			AddSimPeer(fnodes, index+3, index-(2*circleSize)-circleSize*2/3)
			AddSimPeer(fnodes, index+5, index-(3*circleSize)-circleSize*2/3+1)

			if index >= len(fnodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < cfg.Factomd.SimCount; i++ {
			AddSimPeer(fnodes, i-1, i)
		}

	}

	var colors []string = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	if len(fnodes) > 2 {
		for i, s := range fnodes {
			fmt.Printf("%d {color:#%v, shape:dot, label:%v}\n", i, colors[i%len(colors)], s.State.FactomNodeName)
		}
		fmt.Printf("Paste the network info above into http://arborjs.org/halfviz to visualize the network\n")
	}
	// Initiate dbstate plugin if enabled. Only does so for first node,
	// any more nodes on sim control will use default method
	fnodes[0].State.SetTorrentUploader(cfg.Factomd.PluginTorrentUpload)
	if cfg.Factomd.PluginTorrent {
		fnodes[0].State.SetUseTorrent(true)
		manager, err := LaunchDBStateManagePlugin(cfg.Factomd.PluginPath, fnodes[0].State.InMsgQueue(), fnodes[0].State, fnodes[0].State.GetServerPrivateKey(), cfg.Factomd.PprofMPR)
		if err != nil {
			panic("Encountered an error while trying to use torrent DBState manager: " + err.Error())
		}
		fnodes[0].State.DBStateManager = manager
	} else {
		fnodes[0].State.SetUseTorrent(false)
	}

	if cfg.Factomd.JournalFile != "" {
		go LoadJournal(s, cfg.Factomd.JournalFile)
		startServers(false)
	} else {
		startServers(true)
	}

	// Start the webserver
	wsapi.Start(fnodes[0].State)
	if fnodes[0].State.DebugExec() && messages.CheckFileName("graphData.txt") {
		go printGraphData("graphData.txt", 30)
	}

	// Start prometheus on port
	launchPrometheus(9876)
	// Start Package's prometheus
	state.RegisterPrometheus()
	p2p.RegisterPrometheus()
	leveldb.RegisterPrometheus()
	RegisterPrometheus()

	go controlPanel.ServeControlPanel(fnodes[0].State.ControlPanelChannel, fnodes[0].State, connectionMetricsChannel, p2pNetwork, Build, s.FactomNodeName)

	go SimControl(cfg.Factomd.SimFocus, !cfg.Factomd.SimNoInput)

}

func printGraphData(filename string, period int) {
	downscale := int64(1)
	messages.LogPrintf(filename, "\t%9s\t%9s\t%9s\t%9s\t%9s\t%9s", "Dbh-:-min", "Node", "ProcessCnt", "ListPCnt", "UpdateState", "SleepCnt")
	for {
		for _, f := range fnodes {
			s := f.State
			messages.LogPrintf(filename, "\t%9s\t%9s\t%9d\t%9d\t%9d\t%9d", fmt.Sprintf("%d-:-%d", s.LLeaderHeight, s.CurrentMinute), s.FactomNodeName, s.StateProcessCnt/downscale, s.ProcessListProcessCnt/downscale, s.StateUpdateState/downscale, s.ValidatorLoopSleepCnt/downscale)
		}
		time.Sleep(time.Duration(period) * time.Second)
	} // for ever ...
}

//**********************************************************************
// Functions that access variables in this method to set up Factom Nodes
// and start the servers.
//**********************************************************************
func makeServer(s *state.State) *FactomNode {
	// All other states are clones of the first state.  Which this routine
	// gets passed to it.
	newState := s

	if len(fnodes) > 0 {
		newState = s.Clone(len(fnodes)).(*state.State)
		newState.EFactory = new(electionMsgs.ElectionsFactory) // not an elegant place but before we let the messages hit the state
		time.Sleep(10 * time.Millisecond)
		newState.Init()
		newState.EFactory = new(electionMsgs.ElectionsFactory)
	}

	fnode := new(FactomNode)
	fnode.State = newState
	fnodes = append(fnodes, fnode)
	fnode.MLog = mLog

	return fnode
}

func startServers(load bool) {
	for i, fnode := range fnodes {
		if i > 0 {
			fnode.State.Init()
		}
		go NetworkProcessorNet(fnode)
		if load {
			go state.LoadDatabase(fnode.State)
		}
		go fnode.State.GoSyncEntries()
		go Timer(fnode.State)
		go elections.Run(fnode.State)
		go fnode.State.ValidatorLoop()
	}
}

func setupFirstAuthority(s *state.State) {
	if len(s.IdentityControl.Authorities) > 0 {
		//Don't initialize first authority if we are loading during fast boot
		//And there are already authorities present
		return
	}

	s.IdentityControl.SetBootstrapIdentity(s.GetNetworkBootStrapIdentity(), s.GetNetworkBootStrapKey())
}

func networkHousekeeping() {
	for {
		time.Sleep(1 * time.Second)
		p2pProxy.SetWeight(p2pNetwork.GetNumberOfConnections())
	}
}
