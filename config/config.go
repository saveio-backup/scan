package config

//VERSION set this via ldflags
var VERSION = ""
var (
	TRACKER_DB_PATH = "./TrackerLevelDB"
	WALLET_FILE     = "./wallet.dat"
	WALLET_PWD      = "123"
)
//test
var (
	Host="192.168.1.1:11"
	Wallet1Addr = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
)

//default common parameter
const (
	DEFAULT_INIT_DIR         = ".conf/"
	DEFAULT_LOG_DIR          = "./log/"
	DEFAULT_LOG_LEVEL        = 2                //INFO
	DEFAULT_MAX_LOG_SIZE     = 20 * 1024 * 1024 //MB
	DEFAULT_WALLET_FILE_NAME = "./wallet.dat"
	DEFAULT_PWD              = "123"
)

// Tacker&DNS default port
const (
	DEFAULT_TRACKER_PORT = 6369
	DEFAULT_TRACKER_FEE  = 0
	DEFAULT_SYN_PORT     = 6699
	DEFAULT_HOST         = "127.0.0.1:6699"
	DEFAULT_DNS_PORT     = 53
	DEFAULT_DNS_FEE      = 0
	DEFAULT_RPC_PORT                        = uint(20338)
	DEFAULT_RPC_LOCAL_PORT                  = uint(20339)
	DEFAULT_REST_PORT                       = uint(20335)

)

type RpcConfig struct {
	EnableHttpJsonRpc bool
	HttpJsonPort      uint
	HttpLocalPort     uint
}

type RestfulConfig struct {
	EnableHttpRestful bool
	HttpRestPort      uint
	HttpCertPath      string
	HttpKeyPath       string
}

// TrackerConfig config
type TrackerConfig struct {
	UdpPort   uint   // UDP server port
	Fee       uint64 // Service fee
	SeedLists []string
	SyncPort  uint //tracker sync msg
}

// DnsConfig dns config
type DnsConfig struct {
	UdpPort uint   // UDP server port
	Fee     uint64 // Service fee
}

type CommonConfig struct {
	LogLevel  uint
	LogStderr bool
	Tracker   TrackerConfig
	Dns       DnsConfig
	Rpc       RpcConfig
	Restful   RestfulConfig
}

func DefSeedsConfig() *CommonConfig {
	return &CommonConfig{
		LogLevel:  DEFAULT_LOG_LEVEL,
		LogStderr: false,
		Tracker: TrackerConfig{
			UdpPort:   DEFAULT_TRACKER_PORT,
			SyncPort:  DEFAULT_SYN_PORT,
			Fee:       DEFAULT_TRACKER_FEE,
			SeedLists: nil,
		},
		Dns: DnsConfig{
			UdpPort: DEFAULT_DNS_PORT,
			Fee:     DEFAULT_DNS_FEE,
		},
		Rpc: RpcConfig{
			EnableHttpJsonRpc: true,
			HttpJsonPort:      DEFAULT_RPC_PORT,
			HttpLocalPort:     DEFAULT_RPC_LOCAL_PORT,
		},
		Restful: RestfulConfig{
			EnableHttpRestful: true,
			HttpRestPort:      DEFAULT_REST_PORT,
		},
	}
}

//current default config
var DefaultConfig = DefSeedsConfig()
