package config

import (
	"fmt"
	"strings"

	"github.com/go-ini/ini"
)

// *** DEPRECATED SUPPORT **
// This file holds all the functions to handle the old format of the ini and command line.
// It will potentially be removed at some point in the future, though there is no time scheduled
// as of yet.

func convertOldFlags(flags *Flags) {
	// helper function to move the value of flag a to flag b
	_move := func(a, b string) {
		if val, ok := flags.flags[a]; ok {
			flags.flags[b] = val
			delete(flags.flags, a)
		}
	}
	// helper function to move the value of flag a to flag b
	// and at the same time reverse the boolean value
	// if the previous default was "true"
	_reverseBool := func(a, b string) {
		if val, ok := flags.flags[a]; ok {
			if val == "false" {
				flags.flags[b] = "true"
			} // if it's not explicitly false, fall back to default
			delete(flags.flags, a)
		}
	}

	// ****************************
	// ALL FLAG NAMES ARE lowercase
	// ****************************

	// overwrite network with customnet
	if val, ok := flags.Get("network"); ok && val == "CUSTOM" {
		_move("customnet", "network")
	}
	_move("factomhome", "homedir")
	_move("blktime", "blocktime")

	_reverseBool("balancehash", "nobalancehash")

	_move("controlpanelsetting", "controlpanel")
	_move("nodename", "controlpanelname")
	_move("exposeprofiler", "pprofexpose")
	_move("logPort", "pprofport")
	_move("mpr", "pprofmpr")
	_move("tls", "webtls")
	_move("selfaddr", "webtlscertificatehosts")
	_move("rpcuser", "webusername")
	_move("rpcpass", "webpassword")
	_move("prefix", "dbslug")
	_reverseBool("fast", "dbnofastboot")
	_move("fastsaverate", "dbfastbootrate")
	_move("enablenet", "p2pdisable")
	_move("networkport", "p2pPort")
	_move("broadcastnum", "p2pfanout")
	_move("peers", "p2pspecialpeers")

	ex1, ok1 := flags.flags["exclusive"]
	ex2, ok2 := flags.flags["exclusive_in"]
	delete(flags.flags, "exclusive")
	delete(flags.flags, "exclusive_in")
	if ok1 || ok2 {
		switch {
		case ex1 == "true" && ex2 == "true":
			flags.flags["p2pconnectionpolicy"] = "REFUSE"
		case ex1 == "true":
			flags.flags["p2pconnectionpolicy"] = "ACCEPT"
		default:
			flags.flags["p2pconnectionpolicy"] = "NORMAL"
		}
	}

	_move("deadline", "p2ptimeout")
	_move("loglvl", "loglevel")
	_move("logjson", "logjson")
	if flags.flags["logstash"] == "true" {
		_move("logurl", "loglogstash")
	}
	delete(flags.flags, "logstash")
	delete(flags.flags, "logurl")

	_move("debuglog", "logmessages")
	_move("wrproc", "logdbstates")
	_reverseBool("sim_stdin", "simnoinput")
	_move("cnt", "simcount")
	_move("node", "simfocus")
	_move("net", "simnet")
	_move("fnet", "simnetfile")
	_move("drop", "simdroprate")
	_move("timedelta", "simtimeoffset")
	_move("runtimelog", "simruntimelog")
	_move("waitentries", "simwait")

	if val, ok := flags.flags["debugconsole"]; ok {
		var getPort bool
		if strings.Contains(val, "localhost") {
			flags.flags["debugconsole"] = "LOCAL"
			getPort = true
		} else if strings.Contains(val, "remote") {
			flags.flags["debugconsole"] = "ON"
			getPort = true
		}

		if getPort {
			parts := strings.Split(val, ":")
			if len(parts) != 2 {
				Shutdown(fmt.Errorf("invalid format of -debugconsole (deprecated format). needs to be \"(localhost|remote):port\""))
			}
			flags.flags["debugconsoleport"] = parts[1]
		}
	}

	if val, ok := flags.flags["checkheads"]; ok {
		if val == "true" {
			if flags.flags["fixheads"] == "false" {
				flags.flags["chainheadfix"] = "IGNORE"
			} else {
				flags.flags["chainheadfix"] = "ON"
			}
		} else {
			flags.flags["chainheadfix"] = "OFF"
		}

		delete(flags.flags, "checkheads")
		delete(flags.flags, "fixheads")
	}

	_move("rotate", "oneleader")
	_move("sync2", "forcesync2height")
	_move("plugin", "pluginpath")
	_move("tormanage", "plugintorrent")
	_move("torupload", "plugintorrentupload")

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
	// ini package converts string to lowercase automatically since case insensitive is set
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
