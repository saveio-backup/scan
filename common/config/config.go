package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/saveio/edge/common"
	"github.com/saveio/scan/cmd/flags"
	chainFlags "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
	"github.com/urfave/cli"
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
	DEFAULT_CONFIG_FILE      = "/scan_config.json"
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
	DisableChain  bool   `json:"DisableChain"`

	ChannelNetworkId     uint32 `json:"ChannelNetworkId"`
	ChannelProtocol      string `json:"ChannelProtocol"`
	ChannelPortOffset    uint   `json:"ChannelPortOffset"`
	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`
	ChannelSettleTimeout string `json:"ChannelSettleTimeout"`

	DnsNetworkId      uint32 `json:"DnsNetworkId"`
	DnsProtocol       string `json:"DnsProtocol"`
	DnsPortOffset     uint   `json:"DnsPortOffset"`
	DnsGovernDeposit  uint64 `json:"DnsGovernDeposit"`
	DnsChannelDeposit uint64 `json:"DnsChannelDeposit"`

	TrackerPortOffset        uint     `json:"TrackerPortOffset"`
	TrackerFee               int      `json:"TrackerFee`
	TrackerSeedList          []string `json:"TrackerSeedList"`
	TrackerPeerValidDuration string   `json:"TrackerPeerValidDuration"`
	NATProxyServerAddr       string   `json:"NATProxyServerAddr"`

	WalletPwd string `json:"WalletPwd"`
	WalletDir string `json:"WalletDir"`

	DBPath string `json:"DBPath"`

	AutoSetupDNSRegisterEnable bool     `json:"AutoSetupDNSRegisterEnable"`
	DnsNodeMaxNum              int      `json:"DnsNodeMaxNum"`
	SeedInterval               int      `json:"SeedInterval"`
	Fee                        uint64   `json:"Fee"`
	IgnoreConnectDNSAddrs      []string `json:"IgnoreConnectDNSAddrs"`

	DumpMemory bool `json:"DumpMemory"`
}

type FsConfig struct {
	FsRepoRoot string `json:"FsRepoRoot"`
	FsFileRoot string `json:"FsFileRoot"`
	FsType     int    `json:"FsType"`
}

type ScanConfig struct {
	Base     BaseConfig `json:"Base"`
	FsConfig FsConfig   `json:"Fs"`
}

func TestConfig() *ScanConfig {
	return &ScanConfig{
		Base: BaseConfig{
			BaseDir:                    DEFAULT_DB_PATH,
			LogLevel:                   DEFAULT_LOG_LEVEL,
			LogStderr:                  false,
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
		},
	}
}

//current default config
var Parameters = DefaultConfig()
var configDir string
var curUsrWalAddr string

func DefaultConfig() *ScanConfig {
	configDir = "." + DEFAULT_CONFIG_FILE
	existed := common.FileExisted(configDir)
	if !existed {
		log.Debugf("%s config is not existed", configDir)
		return TestConfig()
	}
	cfg := &ScanConfig{}
	GetJsonObjectFromFile(configDir, cfg)
	return cfg
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

func setConfigByCommandParams(scanConfig *ScanConfig, ctx *cli.Context) {
	///////////////////// protocol setting ///////////////////////////
	if ctx.GlobalIsSet(flags.GetFlagName(flags.LogLevelFlag)) {
		scanConfig.Base.LogLevel = ctx.Uint(flags.GetFlagName(flags.LogLevelFlag))
	}
	if ctx.GlobalIsSet(flags.GetFlagName(chainFlags.AccountPassFlag)) {
		scanConfig.Base.WalletPwd = ctx.String(chainFlags.GetFlagName(chainFlags.AccountPassFlag))
	}
}

func SetScanConfig(ctx *cli.Context) {
	setConfigByCommandParams(Parameters, ctx)
}

func Init(ctx *cli.Context) {
	if ctx.GlobalIsSet(flags.GetFlagName(flags.ScanConfigFlag)) {
		scanConfigStr := ctx.String(flags.GetFlagName(flags.ScanConfigFlag))
		f, err := os.Stat(scanConfigStr)
		if err != nil {
			log.Fatalf(fmt.Sprintf("scan config not found, %v\n", err))
		}
		if f.IsDir() {
			configDir = scanConfigStr + DEFAULT_CONFIG_FILE
		} else {
			configDir = scanConfigStr
		}
	} else {
		configDir = "." + DEFAULT_CONFIG_FILE
	}
	log.Infof("scan config path %s", configDir)
	existed := common.FileExisted(configDir)
	if !existed {
		log.Infof("config file is not exist: %s, use default config", configDir)
		return
	}
	GetJsonObjectFromFile(configDir, Parameters)
	SetScanConfig(ctx)
}

func Save() error {
	data, err := json.MarshalIndent(Parameters, "", "  ")
	if err != nil {
		return err
	}
	err = os.Remove(ConfigDir)
	if err != nil {
		return err
	}
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

func EndpointDBPath() string {
	return filepath.Join(Parameters.Base.BaseDir, Parameters.Base.DBPath, curUsrWalAddr, "endpoint")
}

// ChannelDBPath. channel database path
func ChannelDBPath() string {
	return filepath.Join(Parameters.Base.BaseDir, Parameters.Base.DBPath, curUsrWalAddr, "channel")
}
