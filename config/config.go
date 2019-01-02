package config

//VERSION set this via ldflags
var VERSION = ""

//default common parameter
const (
	DEFAULT_INIT_DIR     = ".conf/"
	DEFAULT_LOG_DIR      = "./log"
	DEFAULT_LOG_LEVEL    = 1                //INFO
	DEFAULT_MAX_LOG_SIZE = 20 * 1024 * 1024 //MB
)

// Tacker&DNS default port
const (
	DEFAULT_TRACKER_PORT = 6369
	DEFAULT_TRACKER_FEE  = 0

	DEFAULT_DNS_PORT = 53
	DEFAULT_DNS_FEE  = 0
)

// TrackerConfig config
type TrackerConfig struct {
	UdpPort uint   // UDP server port
	Fee     uint64 // Service fee
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
}

func DefSeedsConfig() *CommonConfig {
	return &CommonConfig{
		LogLevel:  DEFAULT_LOG_LEVEL,
		LogStderr: false,
		Tracker: TrackerConfig{
			UdpPort: DEFAULT_TRACKER_PORT,
			Fee:     DEFAULT_TRACKER_FEE,
		},
		Dns: DnsConfig{
			UdpPort: DEFAULT_DNS_PORT,
			Fee:     DEFAULT_DNS_FEE,
		},
	}
}

//current default config
var DefaultConfig = DefSeedsConfig()
