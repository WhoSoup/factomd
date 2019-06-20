package config

import (
	"fmt"
	"strings"

	"github.com/go-ini/ini"
)

func isOldFormat(file *ini.File) bool {
	sects := file.SectionStrings()
	for _, sec := range sects {
		if sec == "app" {
			return true
		}
	}
	return false
}

func convertOldFlags(flags *Flags) {
	_move := func(a, b string) {
		if val, ok := flags.flags[a]; ok {
			flags.flags[b] = val
			delete(flags.flags, a)
		}
	}
	_reverseBool := func(a, b string) {
		if val, ok := flags.flags[a]; ok {
			if val == "false" {
				flags.flags[b] = "true"
			}
			delete(flags.flags, a)
		}
	}

	// overwrite network with customnet
	if val, ok := flags.Get("network"); ok && val == "CUSTOM" {
		_move("customnet", "network")
	}
	_move("factomhome", "homedir")
	_move("blktime", "blockTime")

	_reverseBool("balancehash", "noBalanceHash")

	_move("controlpanelsetting", "controlPanel")
	_move("nodename", "controlPanelName")
	_move("exposeprofiler", "pprofExpose")
	_move("logPort", "pprofPort")
	_move("mpr", "pprofMPR")
	_move("tls", "webTLS")
	_move("selfaddr", "webTLSCertificateHosts")
	_move("rpcuser", "webUsername")
	_move("rpcpass", "webPassword")
	_move("prefix", "dbSlug")
	_reverseBool("fast", "dbNoFastBoot")
	_move("fastsaverate", "dbFastBootRate")
	_move("enablenet", "p2pDisable")
	_move("networkport", "p2pPort")
	_move("broadcastnum", "p2pFanout")
	_move("peers", "p2pSpecialPeers")

	ex1, ok1 := flags.flags["exclusive"]
	ex2, ok2 := flags.flags["exclusive_in"]
	delete(flags.flags, "exclusive")
	delete(flags.flags, "exclusive_in")
	if ok1 || ok2 {
		switch {
		case ex1 == "true" && ex2 == "true":
			flags.flags["p2pConnectionPolicy"] = "REFUSE"
		case ex1 == "true":
			flags.flags["p2pConnectionPolicy"] = "ACCEPT"
		default:
			flags.flags["p2pConnectionPolicy"] = "NORMAL"
		}
	}

	_move("deadline", "p2pTimeout")
	_move("loglvl", "logLevel")
	_move("logjson", "logJSON")
	if flags.flags["logstash"] == "true" {
		_move("logurl", "logLogstash")
	}
	delete(flags.flags, "logstash")
	delete(flags.flags, "logurl")

	_move("debuglog", "logMessages")
	_move("wrproc", "logDBStates")
	_reverseBool("sim_stdin", "simNoInput")
	_move("cnt", "simCount")
	_move("node", "simFocus")
	_move("net", "simNet")
	_move("fnet", "simNetFile")
	_move("drop", "simDropRate")
	_move("timedelta", "simTimeOffset")
	_move("runtimelog", "simRuntimeLog")
	_move("waitentries", "simWait")

	if val, ok := flags.flags["debugconsole"]; ok {
		var getPort bool
		if strings.Contains(val, "localhost") {
			flags.flags["debugConsole"] = "LOCAL"
			getPort = true
		} else if strings.Contains(val, "remote") {
			flags.flags["debugConsole"] = "ON"
			getPort = true
		}

		if getPort {
			parts := strings.Split(val, ":")
			if len(parts) != 2 {
				Shutdown(fmt.Errorf("invalid format of -debugconsole (deprecated format). needs to be \"(localhost|remote):port\""))
			}
			flags.flags["debugConsolePort"] = parts[1]
		}
	}

	if val, ok := flags.flags["checkheads"]; ok {
		if val == "true" {
			if flags.flags["fixheads"] == "false" {
				flags.flags["chainHeadFix"] = "IGNORE"
			} else {
				flags.flags["chainHeadFix"] = "ON"
			}
		} else {
			flags.flags["chainHeadFix"] = "OFF"
		}

		delete(flags.flags, "checkheads")
		delete(flags.flags, "fixheads")
	}

	_move("rotate", "oneLeader")
	_move("sync2", "forceSync2Height")
	_move("plugin", "pluginPath")
	_move("tormanage", "pluginTorrent")
	_move("torupload", "pluginTorrentUpload")

	// TODO journal

	delete(flags.flags, "clonedb")

}

func fromOldFormat(old *ini.File) *ini.File {
	cfg := ini.Empty()
	cfg.NewSections("Walletd", "Factomd")

	for _, k := range old.Section("Walletd").KeyStrings() {
		cfg.Section("Walletd").NewKey(k, old.Section("Walletd").Key(k).String())
	}

	n, _ := cfg.GetSection("Factomd")
	o, _ := cfg.GetSection("app")

	// helper function. sets new config[b] = old config[a], but only if setting exists
	_move := func(a, b string) {
		if o.HasKey(a) {
			n.NewKey(b, o.Key(a).String())
		}
	}

	// simple replacements
	_move("network", "network")
	_move("homedir", "homedir")
	_move("DirectoryBlockInSeconds", "blocktime")
	_move("ExchangeRateChainId", "oracleChain")
	_move("CustomBootstrapIdentity", "bootstrapIdentity")
	_move("CustomBootstrapKey", "boostrapKey")
	_move("IdentityChainID", "identityChain")
	_move("LocalServerPrivKey", "identityPrivateKey")
	_move("LocalServerPublicKey", "identityPublicKey")
	_move("ChangeAcksHeight", "identityActivationHeight")
	_move("port", "apiPort")
	_move("ControlPanelSetting", "controlPanel")
	_move("ControlPanelPort", "controlPanelPort")
	_move("FactomdTlsEnabled", "webTLS")
	_move("FactomdTlsPublicCert", "webTLSCertificate")
	_move("FactomdTlsPrivateKey", "webTLSKey")
	_move("FactomdRpcUser", "webUsername")
	_move("FactomdRpcPass", "webPassword")
	_move("CorsDomains", "webCORS")
	_move("DBType", "dbType")
	_move("LdbPath", "dbLdbPath")
	_move("BoltDBPath", "dbBoltPath")
	_move("ExportData", "dbExportData")
	_move("ExportDataSubPath", "dbExportDataPath")
	_move("FastBootSaveRate", "dbFastBootRate")
	_move("PeersFile", "p2pPeerFileSuffix")
	_move("logLevel", "logLevel")
	_move("LogPath", "logPath")

	// special replacements

	// the old config can't specify custom networks so we're forced to leave this blank
	// and have the value overwritten by the command line
	if n.Key("network").String() == "CUSTOM" { // new key
		n.DeleteKey("network")
	}

	if o.HasKey("nodemode") {
		n.NewKey("forceFollower", bts(o.Key("nodemode").String() == "FULL"))
	}

	switch n.Key("network").String() { // switch on the NEW key
	case "MAIN":
		_move("ExchangeRateAuthorityPublicKeyMainNet", "oraclePublicKey")
		_move("MainNetworkPort", "p2pPort")
		_move("MainSeedURL", "p2pSeed")
		_move("MainSpecialPeers", "p2pSpecialPeers")
	case "TEST":
		_move("ExchangeRateAuthorityPublicKeyTestNet", "oraclePublicKey")
		_move("TestNetworkPort", "p2pPort")
		_move("TestSeedURL", "p2pSeed")
		_move("TestSpecialPeers", "p2pSpecialPeers")
	case "LOCAL":
		_move("ExchangeRateAuthorityPublicKeyLocalNet", "oraclePublicKey")
		_move("LocalNetworkPort", "p2pPort")
		_move("LocalSeedURL", "p2pSeed")
		_move("LocalSpecialPeers", "p2pSpecialPeers")
	default:
		_move("ExchangeRateAuthorityPublicKey", "oraclePublicKey")
		_move("CustomNetworkPort", "p2pPort")
		_move("CustomSeedURL", "p2pSeed")
		_move("CustomSpecialPeers", "p2pSpecialPeers")
	}

	if o.HasKey("FastBoot") {
		n.NewKey("dbNoFastBoot", bts(o.Key("FastBoot").String() == "false")) // reversing bool
	}

	return cfg
}

// bool to string
func bts(b bool) string {
	return fmt.Sprintf("%v", b)
}
