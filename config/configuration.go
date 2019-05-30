package config

import (
	"fmt"
	"reflect"
)

type Config struct {
	Factomd struct {
		Network                  string `default:"blah"`
		HomeDir                  string
		BlockTime                int `def:"500"`
		FaultTimeout             int
		RoundTimeout             int
		ForceFollower            bool
		OracleChain              string
		OraclePublicKey          string
		BootstrapIdentity        string
		BootstrapKey             string
		BalanceHash              string
		StartDelay               int
		IdentityChain            string
		IdentityPrivateKey       string
		IdentityPublicKey        string
		IdentityActivationHeight int

		ApiPort           int
		ControlPanel      string
		ControlPanelPort  int
		ControlPanelName  int
		PprofExpose       bool
		PprofPort         int
		PprofMMR          int
		WebTLS            bool
		WebTLSKey         string
		WebTLSCertificate string
		WebTLSAddress     string
		WebUsername       string
		WebPassword       string
		WebCORS           string

		DbType           string
		DbSlug           string
		DbLdbPath        string
		DbBoltPath       string
		DbExportData     bool
		DbExportDataPath string
		DbDataStorePath  string
		DbFastBoot       bool
		DbFastBootRate   int

		P2PEnable      bool
		P2PPeerFile    bool
		P2PPort        int
		P2PSeed        string
		P2PFanout      int
		P2PSpecialPeer []string
		P2PMode        string
		P2PTimeout     int

		LogLevel    string
		LogPath     string
		LogJson     bool
		LogLogstash string
		LogStdOut   string
		LogStdErr   string
		LogMessages string
		LogDBStates bool

		SimConsole    bool
		SimCount      int
		SimFocus      int
		SimNet        string
		SimNetFile    string
		SimDropRate   int
		SimTimeOffset int
		SimRuntimeLog bool
		SimWait       bool

		DebugConsole     string
		DebugConsolePort int
		ChainHeadCheck   bool
		ChainHeadFix     bool
		OneLeader        bool
		KeepMismatch     bool
		ForceSync2Height int

		JournalFile string
		JournalMode string
		JournalType string

		PluginPath          string
		PluginTorrent       bool
		PluginTorrentUpload bool
	}
	Walletd struct {
		WalletRpcUser       string
		WalletRpcPass       string
		WalletTlsEnabled    bool
		WalletTlsPrivateKey string
		WalletTlsPublicCert string
		FactomdLocation     string
		WalletdLocation     string
		WalletEncrypted     bool
	}
}

func DefaultConfig() Config {
	var c Config
	//fmt.Println(reflect.ValueOf(c).Field(0))
	//fmt.Println(reflect.TypeOf(c).Field(0))
	r := reflect.TypeOf(c)
	for i := 0; i < r.NumField(); i++ {
		cat := r.Field(i).Type
		for j := 0; j < cat.NumField(); j++ {
			f := cat.Field(j)
			fmt.Printf("%s %s\n", f.Name, f.Tag)
		}
	}
	return c
}
