package config

import (
	"fmt"
	"reflect"

	"github.com/go-ini/ini"
)

// Config holds all the possible variables that are configurable in the config file and
// settable by command line parameters. Default values and conversion methods are defined
// via tags and variable types
//
// All variables have a "def" tag which holds the default value, which should be writable
// to the variable type. e.g. an int with def:"foo" will panic.
//
// The "f" tag specifies a function that is used to convert user input to the variable type.
// They are defined in converter.go.
//
// The "enum" tag specifies a comma-separated list of strings that the variable can hold.
//
// Variable Types:
// 	string
// 		special tags
//			enum: comma-separated list of allowed values. case insensitive.
//			list: the value is a list separated by the character in the "list" tag. values
//				  will be checked recursively, e.g.: `f:"ipport" list=","`
//		functions (tag "f")
//			network: a valid factom network name
//			hex64: a 32 byte / 64 character long hexadecimal string
//			alpha: alpha-numerical (a-z, 0-9) values only
//			ipport: net addresses in the form <hostname/ip>:<port>
//
//	int
//		special tags:
//			min: optional minimum value
//			max: optional maximum value
// 		functions (tag "f")
//			time: a time duration in seconds, with optional modifiers of ^\d+(s|m|h|d)?$
//				  for seconds, minutes, hours, days converted to seconds
//
//	bool
//		nothing special
//
type Config struct {
	Factomd struct {
		Network                  string `def:"MAIN" f:"network" short:"n" hint:"The name of the network to connect to, such as MAIN, LOCAL, TEST, or fct_community_test"`
		HomeDir                  string `def:"" short:"h" hint:"The path to the working directory where factom will place all files"`
		BlockTime                int    `def:"10m" f:"time" min:"1" short:"b" hint:"The time it takes to build one directory block"`
		FaultTimeout             int    `def:"1m" f:"time" hint:"How long to wait before federated servers are considered inactive"`
		RoundTimeout             int    `def:"30s" f:"time" hint:"How long an election round lasts"`
		ForceFollower            bool   `def:"false" hint:"Force the node to run as a follower"`
		OracleChain              string `def:"111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03" f:"hex64" hint:"The chain id containing the oracle data"`
		OraclePublicKey          string `def:"daf5815c2de603dbfa3e1e64f88a5cf06083307cf40da4a9b539c41832135b4a" f:"hex64" hint:"The public key to verify oracle data"`
		BootstrapIdentity        string `def:"38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9" f:"hex64" hint:"The identity of the node that will be the first federated server and sign the genesis block"`
		BootstrapKey             string `def:"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a" f:"hex64" hint:"The public key of the bootstrap identity. Ed25519 key in hexadecimal"`
		NoBalanceHash            bool   `def:"false" hint:"Don't add balance hashes to ACKs"`
		StartDelay               int    `def:"0" f:"time" hint:"Delay time for when to start processing requests for missing messages"`
		IdentityChain            string `def:"" hint:"The identity chain of the node"`
		IdentityPrivateKey       string `def:"4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d" f:"hex64" hint:"The private key of the identity used to sign messages. Ed25519 key in hexadecimal"`
		IdentityPublicKey        string `def:"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a" f:"hex64" hint:"The public key of the identity used to sign messages. Ed25519 key in hexadecimal"`
		IdentityActivationHeight int    `def:"0" min:"0" hint:"The height at which to activate the identity (for brainswaps)"`

		APIPort                int    `def:"8088" min:"1" max:"65535" hint:"The port at which to access the factomd API"`
		ControlPanel           string `def:"READONLY" enum:"DISABLED,READONLY,READWRITE" hint:"The mode of operation of the control panel"`
		ControlPanelPort       int    `def:"8090" min:"1" max:"65535" hint:"The web-port at which to access the control panel"`
		ControlPanelName       string `def:"" hint:"The display name of the node on the control panel"`
		PprofExpose            bool   `def:"false" hint:"If enabled, the pprof server will accept connections outside of localhost"`
		PprofPort              int    `def:"8090" min:"1" max:"65535" hint:"Port for the pprof frontend"`
		PprofMPR               int    `def:"524288" min:"0" hint:"pprof memory profiling rate. 0 to disable, 1 for everything. default is 512kibi"`
		WebTLS                 bool   `def:"false" hint:"If TLS is enabled, the control panel and API will only be accessible via HTTPS. If you have a certificate, you can specify the location of the certificate and PEM key. If you enable TLS without an existing certificate, factomd will generate a self-signed certificate inside HomeDir"`
		WebTLSCertificate      string `def:"" hint:"Path to the certificate"`
		WebTLSKey              string `def:"" hint:"Path to the PEM key file"`
		WebTLSCertificateHosts string `def:"" list:"," hint:"To include any additional ip addresses or hostnames in the self-signed certificate, add them in a comma-separated list. Note that localhost, 127.0.0.1, and ::1 are included by default.\nExample: \"exampledomain.abc,192.168.0.1,192.168.0.2\""`
		WebUsername            string `def:"" hint:"If set, the control panel and API will require basic http authentication to use"`
		WebPassword            string `def:"" hint:"The password for web authentication"`
		WebCORS                string `def:"" hint:"This sets the Cross-Origin Resource Sharing (CORS) header for the API and Walletd. If left blank, CORS is disabled"`

		DBType           string `def:"LDB" enum:"LDB,BOLT,MAP" short:"db" hint:"Which database architecture to use"`
		DBSlug           string `def:"" f:"alpha" hint:"Set a unique identifier included in the path if you want run multiple databases in the same HomeDir"`
		DBLdbPath        string `def:"database/ldb" hint:"Sub-path relative to HomeDir to store Ldb files"`
		DBBoltPath       string `def:"database/bolt" hint:"Sub-path relative to HomeDir for BoltDB"`
		DBExportData     bool   `def:"false" hint:"If enabled, factomd will turn on the block extractor to export blocks to disk"`
		DBExportDataPath string `def:"database/export" hint:"Sub-path relative to HomeDir for exporting data"`
		DBDataStorePath  string `def:"data/export" hint:"Sub-path relative to HomeDir for the block extractor"`
		DBNoFastBoot     bool   `def:"false" short:"fb" hint:"Disable the use of the FastBoot file to cache block validation"`
		DBFastBootRate   int    `def:"1000" min:"1" hint:"Create a FastBoot entry every X blocks"`

		P2PDisable          bool   `def:"false" hint:"Disable the peer to peer network"`
		P2PPeerFileSuffix   string `def:"peers.json" hint:"The filename suffix of the peers file which is added to the current network"`
		P2PPort             int    `def:"8108" min:"1" hint:"The default port used for network connections"`
		P2PSeed             string `def:"" f:"url" hint:"The URL of the seed file to use for bootstrapping"`
		P2PFanout           int    `def:"16" min:"1" hint:"How many peers to broadcast messages to"`
		P2PSpecialPeers     string `def:"" f:"ipport" list:"," short:"p" hint:"A comma-separated list of peers that the node will always connect to in the format of \"host:port\"\nExample to add four special peers: \"123.456.78.9:8108,97.86.54.32:8108,56.78.91.23:8108,hostname:8108\""`
		P2PConnectionPolicy string `def:"NORMAL" enum:"NORMAL,ACCEPT,REFUSE" hint:"Which peers the node should allow.\n  NORMAL: allows all connections\n  ACCEPT: the node accepts incoming connection but only dials to special peers\n  REFUSE: the node dials to special peers but refuses all incoming connections\n"`
		P2PTimeout          int    `def:"5m" f:"time" hint:"How long peers have to send or receive a message before timing out"`

		LogLevel    string `def:"ERROR" enum:"DEBUG,INFO,NOTICE,WARNING,ERROR,CRITICAL,ALERT,EMERGENCY,NONE" short:"l" hint:"The level of messages to log. Setting includes all options to the right"`
		LogPath     string `def:"database/Log" hint:"The sub-path of HomeDir to store logs in"`
		LogJSON     bool   `def:"false" hint:"If enabled, log files will be written in JSON"`
		LogLogstash string `def:"" f:"url" hint:"The URL of a logstash server to send logs to. Leave blank to disable"`
		LogStdOut   string `def:"" hint:"Specify a file to write a copy of StdOut to file"`
		LogStdErr   string `def:"" hint:"Specify a file to write a copy of StdErr to file"`
		LogMessages string `def:"" short:"m" hint:"A regular expression of which message logs to save in the current working directory"`
		LogDBStates bool   `def:"false" hint:"Save DBStates to disk after being processed"`

		SimNoInput    bool   `def:"false" hint:"Disable keyboard input to the console"`
		SimCount      int    `def:"1" min:"1" short:"sc" hint:"How many simulated nodes to launch with a minimum of one"`
		SimFocus      int    `def:"0" min:"0" hint:"The node to focus on at startup. The first node starts at 0"`
		SimNet        string `def:"LONG" enum:"FILE,SQUARE,LONG,LOOPS,ALOT,ALOT+,TREE,CIRCLES" short:"sn" hint:"The network structure of the simulated network"`
		SimNetFile    string `def:"" hint:"The path to the sim node file for simNet=FILE"`
		SimDropRate   int    `def:"0" min:"0" max:"1000" hint:"Simulated drop rate for packets. Number of messages to drop out of 1000"`
		SimTimeOffset int    `def:"0" f:"time" hint:"Time offset between clocks in simulated nodes"`
		SimRuntimeLog bool   `def:"false" hint:"If enabled, the node will keep track of recently sent messages that can be displayed in the console with the \"m\" command"`
		SimWait       bool   `def:"false" hint:"Pause the processing of entries in the processlist. Equivalent to the \"W\" command"`

		DebugConsole     string `def:"OFF" enum:"OFF,LOCAL,ON" hint:"The mode of the debug console.\n  OFF: no debug console\n  LOCAL: only accepts connections from localhost and launches a terminal\n  ON: accepts remote connections\n"`
		DebugConsolePort int    `def:"8093" min:"1" max:"65535" hint:"The port to launch the console server"`
		ChainHeadFix     bool   `def:"ON" enum:"OFF,IGNORE,ON" hint:"The behavior of validating chain heads on boot\n  OFF: don't check at all\n  IGNORE: check but don't fix\n  ON: check and automatically fix invalid chain heads\n"`
		OneLeader        bool   `def:"false" hint:"If enabled, all entries for one factom-minute will be handled by a VM index 0 instead of being distributed over all VMs"`
		KeepMismatch     bool   `def:"false" hint:"Keep the node's DBState even if the signature doesn't match with the majority"`
		ForceSync2Height int    `def:"-1" min:"-1" hint:"Force the height on the second pass sync. Set to -1 to disable, 0 to force a complete sync"`

		JournalFile string `def:"" hint:"Path to the journal file. Journaling disabled if left blank"`
		JournalMode string `def:"CREATE" enum:"CREATE,READ" hint:"Whether to create a new journal or play back an existing journal"`
		JournalType string `def:"AUTO" enum:"AUTO,FOLLOWER,LEADER" hint:"Force the node to run the journal as a specific node type"`

		PluginPath          string `def:"" hint:"In order for plugins to be enabled, the binaries have to be located inside this folder. Leave blank to disable plugins"`
		PluginTorrent       bool   `def:"false" hint:"Enable torrent sync plugin "`
		PluginTorrentUpload bool   `def:"false" hint:"If enabled, the node is an upload in the torrent network"`
	}
	Walletd struct {
		WalletRpcUser       string `def:""`
		WalletRpcPass       string `def:""`
		WalletTlsEnabled    bool   `def:"false"`
		WalletTlsPrivateKey string `def:""`
		WalletTlsPublicCert string `def:""`
		FactomdLocation     string `def:"localhost:8088"`
		WalletdLocation     string `def:"localhost:8089"`
		WalletEncrypted     bool   `def:"false"`
	}
}

// DefaultConfig populates the default values of the config struct using the default tag
func DefaultConfig() *Config {
	c := new(Config)

	err := c.walk(func(cat reflect.StructField, field reflect.StructField, val reflect.Value) error {
		if def, ok := field.Tag.Lookup("def"); ok {
			err := set(val, def, field.Tag)
			if err != nil {
				return fmt.Errorf("config.Config unable to set default for %s: %v", field.Name, err)
			}
			return nil
		}
		return fmt.Errorf("config.Config struct has no \"def\" tag for variable %s", field.Name)
	})

	if err != nil {
		panic(err)
	}

	return c
}

func (c *Config) walk(do func(category reflect.StructField, field reflect.StructField, val reflect.Value) error) error {
	baseType := reflect.TypeOf(*c)  // de-reference to get type of struct, not pointer
	baseValue := reflect.ValueOf(c) // value of the pointer so we can modify it
	for i := 0; i < baseType.NumField(); i++ {
		cat := baseType.Field(i)
		catVal := baseValue.Elem().Field(i)
		for j := 0; j < cat.Type.NumField(); j++ {
			f := cat.Type.Field(j)
			v := catVal.Field(j)

			err := do(cat, f, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) addConfig(file *ini.File, network string) error {
	if isOldFormat(file) {
		file = fromOldFormat(file)
	}

	err := c.walk(func(category reflect.StructField, field reflect.StructField, val reflect.Value) error {
		section := fmt.Sprintf("%s.%s", category.Name, network) // ini package automatically handles inheritance
		if file.Section(section).HasKey(field.Name) {
			err := set(val, file.Section(section).Key(field.Name).String(), field.Tag)
			if err != nil {
				return fmt.Errorf("invalid value for \"%s.%s\" in the config file:\n%v", section, field.Name, err)
			}
		}

		// default is already set
		return nil
	})

	return err
}

func (c *Config) addFlags(flags *Flags) error {
	convertOldFlags(flags)

	return c.walk(func(category reflect.StructField, field reflect.StructField, val reflect.Value) error {
		// there's a short tag
		if short, ok := field.Tag.Lookup("short"); ok {
			if short, ok = flags.Get(short); ok {
				err := set(val, short, field.Tag)
				if err != nil {
					return fmt.Errorf("invalid input for command line option \"-%s\":\n%v", short, err)
				}
				return nil
			}
		}

		// normal command line
		if long, ok := flags.Get(field.Name); ok {
			err := set(val, long, field.Tag)
			if err != nil {
				return fmt.Errorf("invalid input for command line option \"-%s\":\n%v", field.Name, err)
			}
			return nil
		}

		// default is already set
		return nil
	})
}
