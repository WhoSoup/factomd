package config

const defaultConfig = `
; ------------------------------------------------------------------------------
; Configurations for factomd
; ------------------------------------------------------------------------------
; All settings are case insensitive and you can override specific settings in
; the factomd section with network-by-network settings by adding a 
; [factomd.NETWORKNAME] category, e.g.: [factomd.MAIN] or [factomd.fct_community_test]
; These settings will only take effect for that network
;
; All settings are case-insensitive with a command line equivalent of "--name=value",
; e.g.: "--blocktime=10m"
;
; Time-based variables allow semantic input between seconds (s)
; minutes (m), hours (h), days (d), defaulting to seconds:
;   180 = 180s = 3m
;   48h = 2d
;
[factomd]
; ---------------- GLOBAL ----------------
; The name of the network to connect to, such as MAIN, LOCAL, TEST, or fct_community_test
;network = MAIN

; The directory to keep factom data in. If left blank it defaults to ~/.factom/m2/ on *nix
; and %HOMEPATH%/.factom/m2/ on windows
;homeDir = 

; ---------------- CONSENSUS ----------------
; The time to build one directory block
;blockTime = 10m

; How long to wait for authority nodes each factom-minute before faulting them. Should be higher
; than a tenth of "blockTime"
;faultTimeout = 2m

; How long an audit node has to volunteer before moving to the next one
;roundTimeout = 30s

; Enable to force a node to always run as follower
;forceFollower = false

; The Oracle Chain governs the current exchange rate of Factoshi to EC
;oracleChain = 111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03

; The public key that validates entries to the Oracle chain
;oraclePublicKey = daf5815c2de603dbfa3e1e64f88a5cf06083307cf40da4a9b539c41832135b4a

; The identity of the node that will be the first federated server and sign the genesis block
;bootstrapIdentity = 38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9
; The public key of the bootstrap identity. Ed25519 key in hexadecimal
;bootstrapKey = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a

; Disable adding balance hashes to ACKs
;noBalanceHash = false

; Delay time for when to start processing requests for missing messages
;startDelay = 0s

; ---------------- IDENTITY ----------------
; The identity chain of this node
;identityChain =

; The private key of the identity used to sign messages. Ed25519 key in hexadecimal
;identityPrivateKey = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d

; The public key of the identity used to sign messages. Ed25519 key in hexadecimal
;identityPublicKey = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a

; The height at which to activate the identity (for brainswaps)
;identityActivationHeight = 0


; ---------------- WEB SERVICES ----------------
; The port at which to access the factomd API
;apiPort = 8088

; The mode of operation of the control panel
; Choices are: DISABLED | READONLY | READWRITE
;controlPanel = READONLY

; The web-port at which to access the control panel
;controlPanelPort = 8090

; The display name of the node on the control panel
;controlPanelName = 

; If enabled, the pprof server will accept connections outside of localhost
;pprofExpose = false

; Port for the pprof frontend
;pprofPort = 6060

; pprof memory profiling rate. 0 to disable, 1 for everything. default is 512kibi
;pprofMPR = 524288


; If TLS is enabled, the control panel and API will only be accessible via HTTPS. If you
; have a certificate, you can specify the location of the certificate and PEM key. 
; If you enable TLS without an existing certificate, factomd will generate a self-signed
; certificate inside HomeDir
;webTLS = false
;webTLSCertificate =
;webTLSKey =

; To include any additional ip addresses or hostnames in the self-signed certificate, add
; them in a comma-separated list. Note that localhost, 127.0.0.1, and ::1 are included by default
; Example: "exampledomain.abc,192.168.0.1,192.168.0.2"
;webTLSCertificateHosts = 

; If set, the control panel and API will require basic http authentication to use
;webUsername = 
;webPassword = 

; This sets the Cross-Origin Resource Sharing (CORS) header for the API and Walletd
; If left blank, CORS is disabled
;webCORS = 


; ---------------- DATABASE ----------------
; Which database architecture to use
; Choice of LDB | BOLT | MAP
;   LDB: LevelDB (default)
;   BOLT: BoltDB
;   MAP: in-memory only database
;dbType = LDB

; Set a unique identifier included in the path if you want run multiple databases in the same HomeDir
;dbSlug = 

; Sub-path relative to HomeDir to store Ldb files
;dbLdbPath = database/ldb

; Sub-path relative to HomeDir for BoltDB
;dbBoltPath = database/bolt

; If enabled, factomd will turn on the block extractor to export blocks to disk
;dbExportData = false

; Sub-path relative to HomeDir for exporting data
;dbExportDataPath = database/export/

; Sub-path relative to HomeDir for the block extractor
;dbDataStorePath = data/export

; Disable the use of the FastBoot file to cache block validation
;dbNoFastBoot = false

; Create a FastBoot entry every X blocks
;dbFastBootRate = 1000

; ---------------- P2P ----------------
; Disable the peer to peer network
;p2pDisable = false

; The filename suffix of the peers file which is added to the current network
;p2pPeerFileSuffix = "peers.json"

; The default port used for network connections
;p2pPort = 8108

; The URL of the seed file to use for bootstrapping
;p2pSeed =

; How many peers to broadcast messages to
;p2pFanout = 16

; A comma-separated list of peers that the node will always connect to in the format of "host:port"
; Example to add four special peers:
;   p2pSpecialPeers = "123.456.78.9:8108,97.86.54.32:8108,56.78.91.23:8108,hostname:8108"
;p2pSpecialPeers = 

; Which peers the node should allow.
; Choices:
;   NORMAL: allows all connections (default)
;   ACCEPT: the node accepts incoming connection but only dials to special peers
;   REFUSE: the node dials to special peers but refuses all incoming connections
;p2pConnectionPolicy = NORMAL

; How long peers have to send or receive a message before timing out
;p2pTimeout = 5m


; ---------------- LOGGING ----------------
; The level of messages to log. Setting includes all options to the right
; Choices:
;   DEBUG | INFO | NOTICE | WARNING | ERROR | CRITICAL | ALERT | EMERGENCY | NONE
;logLevel = ERROR

; The sub-path of HomeDir to store logs in
;logPath = database/Log

; If enabled, log files will be written in JSON
;logJson = false

; The URL of a logstash server to send logs to. Leave blank to disable
;logLogstash = 

; Specify a file to write a copy of StdOut to file
;logStdOut =

; Specify a file to write a copy of StdErr to file
;logStdErr = 

; A regular expression of which message logs to save in the current working directory
; For more details see https://factomize.com/forums/threads/logging-in-factomd.1766/
;logMessages = 

; Save DBStates to disk after being processed
; Files will be saved to dbLdbPath/<network>/dbstates/processed_dbstate_<height>.block
;logDBStates = false


; ---------------- SIMULATION ----------------
; Disable keyboard input to the console
;simNoInput = false

; How many simulated nodes to launch with a minimum of one
;simCount = 1

; The node to focus on at startup. The first node starts at 0
;simFocus = 0

; The network structure of the simulated network
; Choices: FILE | SQUARE | LONG | LOOPS | ALOT | ALOT+ | TREE | CIRCLES
;simNet = ALOT+

; The path to the sim node file for simNet=FILE
;simNetFile = 

; Simulated drop rate for packets. Number of messages to drop out of 1000
;simDropRate = 0

; Time offset between clocks in simulated nodes
;simTimeOffset = 0s

; If enabled, the node will keep track of recently sent messages that can be displayed
; in the console with the "m" command
;simRuntimeLog = false

; Pause the processing of entries in the processlist. Equivalent to the "W" command
;simWait = false


; ---------------- DEBUG ----------------
; The mode of the debug console.
; Choices:
;   OFF: no debug console (default)
;   LOCAL: only accepts connections from localhost and launches a terminal
;   ON: accepts remote connections
;debugConsole = OFF

; The port to launch the console server
;debugConsolePort = 8093

; The behavior of validating chain heads on boot
; Choices:
;   OFF: don't check at all
;   IGNORE: check but don't fix
;   ON: check and automatically fix invalid chain heads
;chainHeadFix = ON

; If enabled, all entries for one factom-minute will be handled by a VM index 0
; instead of being distributed over all VMs
;oneLeader = false

; Keep the node's DBState even if the signature doesn't match with the majority
;keepMismatch = false

; Force the height on the second pass sync. Set to -1 to disable, 0 to force a complete sync
;forceSync2Height = -1


; ---------------- JOURNALING ----------------
; Path to the journal file. Journaling disabled if left blank
;journalFile = 

; Whether to create a new journal or play back an existing journal
; Choices: CREATE | READ
;journalMode = READ

; Force the node to run the journal as a specific node type
; Choices:
;   AUTO: let node determine (default)
;   FOLLOWER: node is a follower
;   LEADER: node is a leader
;journalType = AUTO


; ---------------- PLUGINS ----------------
; In order for plugins to be enabled, the binaries have to be located inside this folder. 
; Leave blank to disable plugins
;pluginPath = 

; Enable torrent sync plugin 
;pluginTorrent = false

; If enabled, the node is an upload in the torrent network
;pluginTorrentUpload = false


; ------------------------------------------------------------------------------
; Configurations for factom-walletd
; ------------------------------------------------------------------------------
[Walletd]
; These are the username and password that factom-walletd requires
; This file is also used by factom-cli to determine what login to use
WalletRpcUser = 
WalletRpcPass =

; These define if the connection to the wallet should be encrypted, and if it is, what files
; are the secret key and the public certificate.  factom-cli uses the certificate specified here if TLS is enabled.
; To use default files and paths leave /full/path/to/... in place.
WalletTlsEnabled                      = false
WalletTlsPrivateKey                   = "/full/path/to/walletAPIpriv.key"
WalletTlsPublicCert                   = "/full/path/to/walletAPIpub.cert"

; This is where factom-walletd and factom-cli will find factomd to interact with the blockchain
; This value can also be updated to authorize an external ip or domain name when factomd creates a TLS cert
FactomdLocation                       = "localhost:8088"

; This is where factom-cli will find factom-walletd to create Factoid and Entry Credit transactions
; This value can also be updated to authorize an external ip or domain name when factom-walletd creates a TLS cert
WalletdLocation                       = "localhost:8089"

; Enables wallet database encryption on factom-walletd. If this option is enabled, an unencrypted database
; cannot exist. If an unencrypted database exists, the wallet will exit.
WalletEncrypted                       = false


[factomd.MAIN]
FERPublicKey: daf5815c2de603dbfa3e1e64f88a5cf06083307cf40da4a9b539c41832135b4a
blockTime: 10m
p2pPort: 8108
p2pSeed: https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt

[factomd.TEST]
p2pFERPublicKey: 1d75de249c2fc0384fb6701b30dc86b39dc72e5a47ba4f79ef250d39e21e7a4f
p2pPort: 8109
p2pSeed: https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt

[factomd.LOCAL]
p2pFERPublicKey: 3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29
p2pPort: 8110
p2pSeed: https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt

[factomd.fct_community_test]
p2pFERPublicKey: 58cfccaa48a101742845df3cecde6a9f38037030842d34d0eaa76867904705ae
blockTime: 10m
bootstrapIdentity: 8888882f5002ff95fce15d20ecb7e18ae6cc4d5849b372985d856b56e492ae0f
bootstrapKey: 58cfccaa48a101742845df3cecde6a9f38037030842d34d0eaa76867904705ae
p2pPort: 8110
p2pSeed: https://raw.githubusercontent.com/FactomProject/communitytestnet/master/seeds/testnetseeds.txt
`
