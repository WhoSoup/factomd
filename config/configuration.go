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
//			path: a valid filepath, either relative or absolute, with optional file
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
		Network                  string `def:"MAIN" f:"network"`
		HomeDir                  string `def:"" f:"path"`
		BlockTime                int    `def:"600" f:"time" min:"1"`
		FaultTimeout             int    `def:"60" f:"time"`
		RoundTimeout             int    `def:"30" f:"time"`
		ForceFollower            bool   `def:"false"`
		OracleChain              string `def:"111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03" f:"hex64"`
		OraclePublicKey          string `def:"daf5815c2de603dbfa3e1e64f88a5cf06083307cf40da4a9b539c41832135b4a" f:"hex64"`
		BootstrapIdentity        string `def:"38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9" f:"hex64"`
		BootstrapKey             string `def:"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a" f:"hex64"`
		NoBalanceHash            bool   `def:"false"`
		StartDelay               int    `def:"0" f:"time"`
		IdentityChain            string `def:""`
		IdentityPrivateKey       string `def:"4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d" f:"hex64"`
		IdentityPublicKey        string `def:"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a" f:"hex64"`
		IdentityActivationHeight int    `def:"0" min:"0"`

		APIPort           int    `def:"8088" min:"1" max:"65535"`
		ControlPanel      string `def:"READONLY" enum:"DISABLED,READONLY,READWRITE"`
		ControlPanelPort  int    `def:"8090" min:"1" max:"65535"`
		ControlPanelName  string `def:""`
		PprofExpose       bool   `def:"false"`
		PprofPort         int    `def:"8090" min:"1" max:"65535"`
		PprofMMR          int    `def:"524288" min:"0"`
		WebTLS            bool   `def:"false"`
		WebTLSCertificate string `def:"" f:"path"`
		WebTLSKey         string `def:"" f:"path"`
		WebTLSAddress     string `def:"" list:","`
		WebUsername       string `def:""`
		WebPassword       string `def:""`
		WebCORS           string `def:""`

		DbType           string `def:"LDB" enum:"LDB,BOLT,MAP"`
		DbSlug           string `def:"" f:"alpha"`
		DbLdbPath        string `def:"database/ldb" f:"path"`
		DbBoltPath       string `def:"database/bolt"`
		DbExportData     bool   `def:"false"`
		DbExportDataPath string `def:"database/export"`
		DbDataStorePath  string `def:"data/export"`
		DbNoFastBoot     bool   `def:"false"`
		DbFastBootRate   int    `def:"1000" min:"1"`

		P2PDisable        bool   `def:"false"`
		P2PPeerFileSuffix bool   `def:"false"`
		P2PPort           int    `def:"8108" min:"1"`
		P2PSeed           string `def:"" f:"url"`
		P2PFanout         int    `def:"16" min:"1"`
		P2PSpecialPeer    string `def:"" f:"ipport" list:","`
		P2PMode           string `def:"NORMAL" enum:"NORMAL,ACCEPT,REFUSE"`
		P2PTimeout        int    `def:"300" f:"time"`

		LogLevel    string `def:"ERROR" enum:"DEBUG,INFO,NOTICE,WARNING,ERROR,CRITICAL,ALERT,EMERGENCY,NONE"`
		LogPath     string `def:"database/Log" f:"path"`
		LogJSON     bool   `def:"false"`
		LogLogstash string `def:"" f:"url"`
		LogStdOut   string `def:"" f:"path"`
		LogStdErr   string `def:"" f:"path"`
		LogMessages string `def:""`
		LogDBStates bool   `def:"false"`

		SimNoInput    bool   `def:"false"`
		SimCount      int    `def:"1" min:"1"`
		SimFocus      int    `def:"0" min:"0"`
		SimNet        string `def:"LONG" enum:"FILE,SQUARE,LONG,LOOPS,ALOT,ALOT+,TREE,CIRCLES"`
		SimNetFile    string `def:"" f:"path"`
		SimDropRate   int    `def:"0" min:"0" max:"1000"`
		SimTimeOffset int    `def:"0" f:"time"`
		SimRuntimeLog bool   `def:"false"`
		SimWait       bool   `def:"false"`

		DebugConsole     string `def:"OFF" enum:"OFF,LOCAL,ON"`
		DebugConsolePort int    `def:"8093" min:"1" max:"65535"`
		ChainHeadCheck   bool   `def:"false"`
		ChainHeadFix     bool   `def:"false"`
		OneLeader        bool   `def:"false"`
		KeepMismatch     bool   `def:"false"`
		ForceSync2Height int    `def:"-1" min:"-1"`

		JournalFile string `def:"" f:"path"`
		JournalMode string `def:"CREATE" enum:"CREATE,READ"`
		JournalType string `def:"AUTO" enum:"AUTO,FOLLOWER,LEADER"`

		PluginPath          string `def:"" f:"path"`
		PluginTorrent       bool   `def:"false"`
		PluginTorrentUpload bool   `def:"false"`
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
func DefaultConfig() Config {
	var c Config

	err := apply(&c, func(cat reflect.StructField, field reflect.StructField, val reflect.Value) error {
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

func LoadConfig(path string) (Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return Config{}, err
	}
	return parseConfig(cfg)
}

func LoadString(data string) (Config, error) {
	cfg, err := ini.Load([]byte(data))
	if err != nil {
		return Config{}, err
	}
	return parseConfig(cfg)
}

func parseConfig(file *ini.File) (Config, error) {
	c := DefaultConfig()
	baseType := reflect.TypeOf(c)
	baseValue := reflect.ValueOf(&c)
	for i := 0; i < baseType.NumField(); i++ {
		cat := baseType.Field(i).Type
		catVal := baseValue.Elem().Field(i)
		for j := 0; j < cat.NumField(); j++ {
			f := cat.Field(j)
			v := catVal.Field(j)
			if def, ok := f.Tag.Lookup("def"); ok {
				err := set(v, def, f.Tag)
				if err != nil {
					panic(fmt.Sprintf("config.Config unable to set default for %s: %v", f.Name, err))
				}
			} else {
				panic(fmt.Sprintf("config.Config struct has no \"def\" tag for variable %s", f.Name))
			}
		}
	}
	return c, nil
}

func apply(cfg *Config, do func(category reflect.StructField, field reflect.StructField, val reflect.Value) error) error {
	baseType := reflect.TypeOf(*cfg)  // de-reference to get type of struct, not pointer
	baseValue := reflect.ValueOf(cfg) // value of the pointer so we can modify it
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
