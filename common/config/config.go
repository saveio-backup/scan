package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/saveio/themis/common/log"
	//"os"
)

//VERSION set this via ldflags
var VERSION = ""

var (
	ConfigDir   string
	Wallet1Addr = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
)

//default common parameter
const (
	WALLET_FILE              = "./wallet.dat"
	DEFAULT_CONFIG_FILE      = "./config.json"
	DEFAULT_LOG_DIR          = "./log/"
	DEFAULT_LOG_LEVEL        = 2 //INFO
	DEFAULT_DB_PATH          = "./TrackerLevelDB"
	DEFAULT_MAX_LOG_SIZE     = 20 * 1024 * 1024 //MB
	DEFAULT_WALLET_FILE_NAME = "./wallet.dat"
	DEFAULT_PWD              = "123"
)

//p2p relative

const (
	NETWORK_ID_MAIN_NET = 1
	NETWORK_ID_SOLO_NET = 2
)

// Tacker&DNS default port
const (
	DEFAULT_TRACKER_PORT   = 6369
	DEFAULT_TRACKER_FEE    = 0
	DEFAULT_P2P_PORT       = 6699
	DEFAULT_HOST           = "127.0.0.1:6699"
	DEFAULT_DNS_PORT       = 53
	DEFAULT_DNS_FEE        = 0
	DEFAULT_RPC_PORT       = uint(20448)
	DEFAULT_RPC_LOCAL_PORT = uint(20449)
	DEFAULT_REST_PORT      = uint(20445)
)

type CommonConfig struct {
	LogLevel     uint   `json:"LogLevel"`
	LogStderr    bool   `json:"LogStderr"`
	CommonDBPath string `json:"CommonDBPath"`
	ConfPath     string `json:"ConfPath"`

	ChainRestAddr uint `json:"ChainRestAddr"`
	ChainRpcAddr  uint `json:"ChainRpcAddr"`

	DBPath string `json:"DBPath"`

	DnsNodeMaxNum     int    `json:"DnsNodeMaxNum"`
	SeedInterval      int    `json:"SeedInterval"`
	DnsChannelDeposit uint64 `json:"DnsChannelDeposit"`

	WalletPwd string `json:"WalletPwd"`
	WalletDir string `json:"WalletDir"`

	P2PNATAddr string `json:"P2PNATAddr`
	NetworkId  uint32 `json:"NetworkID`
}

type FsConfig struct {
	FsRepoRoot string `json:"FsRepoRoot"`
	FsFileRoot string `json:"FsFileRoot"`
	FsType     int    `json:"FsType"`
}

type P2PConfig struct {
	Protocol  string   `json:"Protocol"`
	NetworkId uint32   `json:"NetworkId"`
	PublicIp  string   `json:"PublicIp"`
	PortBase  uint     `json:"PortBase"`
	SeedList  []string `json:"SeedList"`
}

type RpcConfig struct {
	EnableHttpJsonRpc bool
	HttpJsonPort      uint
	HttpLocalPort     uint
	RpcServer         string
}

type RestfulConfig struct {
	EnableHttpRestful bool
	HttpRestPort      uint
	HttpCertPath      string
	HttpKeyPath       string
}

// TrackerConfig config
type TrackerConfig struct {
	UdpPort   uint   // UDP server listen port
	Fee       uint64 // Service fee
	SeedLists []string
}

// DnsConfig dns config
type DnsConfig struct {
	AutoSetupDNSRegisterEnable bool
	AutoSetupDNSChannelsEnable bool
	Protocol                   string
	UdpPort                    uint // UDP server port
	InitDeposit                uint64
	ChannelDeposit             uint64
	Fee                        uint64 // Service fee
	IgnoreConnectDNSAddrs      []string
}

type ChannelConfig struct {
	ChannelPortOffset    int    `json:"ChannelPortOffset"`
	ChannelProtocol      string `json:"ChannelProtocol"`
	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`
	ChannelDBPath        string `json:"ChannelDBPath`
}

type DDNSConfig struct {
	CommonConfig  CommonConfig  `json:"Common"`
	P2PConfig     P2PConfig     `json:"P2P"`
	TrackerConfig TrackerConfig `json:"Tracker"`
	DnsConfig     DnsConfig     `json:"Dns"`
	ChannelConfig ChannelConfig `json:"Channel"`
	RpcConfig     RpcConfig     `json:"Rpc"`
	RestfulConfig RestfulConfig `json:"Restful"`
	FsConfig      FsConfig      `json:"Fs"`
}

func DefDDNSConfig() *DDNSConfig {
	return &DDNSConfig{
		CommonConfig: CommonConfig{
			LogLevel:     DEFAULT_LOG_LEVEL,
			LogStderr:    false,
			CommonDBPath: DEFAULT_DB_PATH,
		},
		P2PConfig: P2PConfig{
			NetworkId: NETWORK_ID_MAIN_NET,
			PortBase:  DEFAULT_P2P_PORT,
			SeedList:  nil,
		},
		DnsConfig: DnsConfig{
			AutoSetupDNSRegisterEnable: false,
			AutoSetupDNSChannelsEnable: true,
		},
		TrackerConfig: TrackerConfig{
			UdpPort:   DEFAULT_TRACKER_PORT,
			Fee:       DEFAULT_TRACKER_FEE,
			SeedLists: nil,
		},
		RpcConfig: RpcConfig{
			EnableHttpJsonRpc: true,
			HttpJsonPort:      DEFAULT_RPC_PORT,
			HttpLocalPort:     DEFAULT_RPC_LOCAL_PORT,
		},
		RestfulConfig: RestfulConfig{
			EnableHttpRestful: true,
			HttpRestPort:      DEFAULT_REST_PORT,
		},
		ChannelConfig: ChannelConfig{},
	}
}

//current default config
var DefaultConfig *DDNSConfig

func SetupDefaultConfig() {
	DefaultConfig = GenDefConfig()
}

func GenDefConfig() *DDNSConfig {
	var defConf *DDNSConfig

	err := GetJsonObjectFromFile("./config.json", &defConf)
	if err != nil {
		log.Fatalf("GetDefConfig err: ", err)
	}
	return defConf
}

func GetJsonObjectFromFile(filePath string, jsonObject interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(data, jsonObject)
	if err != nil {
		return fmt.Errorf("json.Unmarshal %s error:%s", data, err)
	}
	return nil
}

var DefaultDDnsConfig = DefDDNSConfig()

func Save() error {
	data, err := json.MarshalIndent(DefaultDDnsConfig, "", "  ")
	if err != nil {
		return err
	}
	//err = os.Remove(ConfigDir)
	//if err != nil {
	//	return err
	//}
	return ioutil.WriteFile(ConfigDir, data, 0666)
}
