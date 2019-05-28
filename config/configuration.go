package config

type Config struct {
	App struct {
		Network           string
		Home              string
		Name              string
		Prefix            string
		BlockTime         int
		FaultTimeout      int
		RoundTimeout      int
		Mode              string // enum
		FERChain          string
		FERPublicKey      string
		BootstrapIdentity string
		BootstrapKey      string
		BalanceHash       bool
		StartDelay        int64
	}
	Identity struct {
		Chain            string
		PrivateKey       string
		PublicKey        string
		ActivationHeight uint32
	}
	Services struct {
		Port             int
		ControlPanel     string
		ControlPanelPort int
		TLS              bool
		TLSAddress       []string
		TLSKey           string
		TLSCertificate   string
		Username         string
		Password         string
		CORS             string
		PprofExpose      bool
		PprofPort        int
	}
	DB struct {
		Enable      bool
		PeerFile    bool
		Port        int
		Seed        string
		SpecialPeer string
		Exclusive   int
		Timeout     int
	}
	Log struct {
		Level    string
		Path     string
		Console  string
		Json     bool
		Logstash string
	}
	Walletd struct {
		Encryption     bool
		Username       string
		Password       string
		TLS            bool
		TLSKey         string
		TLSCertificate string
	}
	Remote struct {
		Factomd string
		Walletd string
	}
	Sim struct {
		Focus      int
		Count      int
		Net        string
		File       string
		DropRate   int
		TimeOffset int
	}
	Debug struct {
		Console           string
		StdIn             bool
		StdOutLog         string
		ErrOutLog         string
		MsgLog            string
		CheckChainHeads   bool
		FixChainHeads     bool
		WaitEntries       bool
		RunetimeLog       bool
		OneLeader         bool
		KeepMismatch      bool
		MemoryProfileRate int
		ForceSync2Height  int
		SaveDBStates      bool
	}
	Journal struct {
		Write bool
		File  string
		Mode  string
	}
	Plugins struct {
		Path          string
		TorrentSync   bool
		TorrentUpload bool
	}
}
