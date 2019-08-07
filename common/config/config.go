package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

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

type BaseConfig struct {
	BaseDir            string `json:"BaseDir"`
	LogLevel           uint   `json:"LogLevel"`
	LogStderr          bool   `json:"LogStderr"`
	NetworkId          uint32 `json:"NetworkId"`
	PublicIP           string `json:"PublicIP"`
	PortBase           uint   `json:"PortBase"`
	LocalRpcPortOffset uint   `json:"LocalRpcPortOffset"`
	EnableLocalRpc     bool   `json:"EnableLocalRpc"`
	JsonRpcPortOffset  uint   `json:"JsonRpcPortOffset"`
	EnableJsonRpc      bool   `json:"EnableJsonRpc"`
	HttpRestPortOffset uint   `json:"HttpRestPortOffset"`
	HttpCertPath       string `json:"HttpCertPath"`
	HttpKeyPath        string `json:"HttpKeyPath"`
	EnableRest         bool   `json:"EnableRest"`

	ChainRpcAddr  string `json:"ChainRpcAddr"`
	ChainRestAddr string `json:"ChainRestAddr"`

	ChannelPortOffset    uint   `json:"ChannelPortOffset"`
	ChannelProtocol      string `json:"ChannelProtocol"`
	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`
	ChannelSettleTimeout string `json:"ChannelSettleTimeout"`

	TrackerPortOffset  uint     `json:"TrackerPortOffset"`
	TrackerFee         int      `json:"TrackerFee`
	TrackerSeedList    []string `json:"TrackerSeedList"`
	NATProxyServerAddr string   `json:"NATProxyServerAddr"`

	WalletPwd string `json:"WalletPwd"`
	WalletDir string `json:"WalletDir"`

	DBPath string `json:"DBPath"`

	DnsNodeMaxNum              int      `json:"DnsNodeMaxNum"`
	SeedInterval               int      `json:"SeedInterval"`
	DnsChannelDeposit          uint64   `json:"DnsChannelDeposit"`
	AutoSetupDNSRegisterEnable bool     `json:"AutoSetupDNSRegisterEnable"`
	AutoSetupDNSChannelsEnable bool     `json:"AutoSetupDNSChannelsEnable"`
	Protocol                   string   `json:"Protocol"`
	UdpPort                    uint     `json:"UdpPort"`
	InitDeposit                uint64   `json:"InitDeposit"`
	ChannelDeposit             uint64   `json:"ChannelDeposit"`
	Fee                        uint64   `json:"Fee"`
	IgnoreConnectDNSAddrs      []string `json:"IgnoreConnectDNSAddrs"`
}

type FsConfig struct {
	FsRepoRoot string `json:"FsRepoRoot"`
	FsFileRoot string `json:"FsFileRoot"`
	FsType     int    `json:"FsType"`
}

type DDNSConfig struct {
	Base     BaseConfig `json:"Base"`
	FsConfig FsConfig   `json:"Fs"`
}

func DefDDNSConfig() *DDNSConfig {
	return &DDNSConfig{
		Base: BaseConfig{
			BaseDir:                    DEFAULT_DB_PATH,
			LogLevel:                   DEFAULT_LOG_LEVEL,
			LogStderr:                  false,
			NetworkId:                  NETWORK_ID_MAIN_NET,
			PublicIP:                   "127.0.0.1",
			EnableJsonRpc:              true,
			JsonRpcPortOffset:          DEFAULT_RPC_PORT,
			EnableLocalRpc:             false,
			LocalRpcPortOffset:         DEFAULT_RPC_LOCAL_PORT,
			EnableRest:                 true,
			HttpRestPortOffset:         DEFAULT_REST_PORT,
			TrackerPortOffset:          DEFAULT_TRACKER_PORT,
			TrackerFee:                 DEFAULT_TRACKER_FEE,
			TrackerSeedList:            nil,
			AutoSetupDNSRegisterEnable: false,
			AutoSetupDNSChannelsEnable: false,
			UdpPort:                    DEFAULT_P2P_PORT,
		},
	}
}

//current default config
var Parameters *DDNSConfig
var curUsrWalAddr string

func SetupDefaultConfig() {
	Parameters = GenDefConfig()
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

func SetCurrentUserWalletAddress(addr string) {
	curUsrWalAddr = addr
}

// DspDBPath. dsp database path
func DspDBPath() string {
	return filepath.Join(Parameters.Base.BaseDir, Parameters.Base.DBPath, curUsrWalAddr, "dsp")
}

// TrackerDBPath. tracker database path
func TrackerDBPath() string {
	return filepath.Join(Parameters.Base.BaseDir, Parameters.Base.DBPath, curUsrWalAddr, "tracker")
}

// ChannelDBPath. channel database path
func ChannelDBPath() string {
	return filepath.Join(Parameters.Base.BaseDir, Parameters.Base.DBPath, curUsrWalAddr, "channel")
}
