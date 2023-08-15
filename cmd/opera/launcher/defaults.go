package launcher

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	DefaultP2PPort  = 5050  // Default p2p port for listening
	DefaultHTTPPort = 18545 // Default TCP port for the HTTP RPC server
	DefaultWSPort   = 18546 // Default TCP port for the websocket RPC server
)

func overrideFlags() {
	utils.ListenPortFlag.Value = DefaultP2PPort
	utils.HTTPPortFlag.Value = DefaultHTTPPort
	utils.LegacyRPCPortFlag.Value = DefaultHTTPPort
	utils.WSPortFlag.Value = DefaultWSPort
}

// DefaultNodeConfig creates reasonable default configuration settings
func DefaultNodeConfig() node.Config {
	return node.Config{
		DataDir:             DefaultDataDir(),
		HTTPPort:            DefaultHTTPPort,
		HTTPTimeouts:        rpc.DefaultHTTPTimeouts,
		HTTPVirtualHosts:    []string{"localhost"},
		HTTPModules:         []string{"eth", "ftm", "dag", "abft", "web3"},
		WSModules:           []string{"eth", "ftm", "dag", "abft", "web3"},
		WSPort:              DefaultWSPort,
		GraphQLVirtualHosts: []string{"localhost"},
		P2P: p2p.Config{
			NoDiscovery: false, // enable discovery v4 by default
			DiscoveryV5: true,  // enable discovery v5 by default
			ListenAddr:  fmt.Sprintf(":%d", DefaultP2PPort),
			MaxPeers:    50,
			NAT:         nat.Any(),
		},
		Name:    clientIdentifier,
		Version: params.VersionWithCommit(gitCommit, gitDate),
		IPCPath: "opera.ipc",
	}
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Lachesis")
		case "windows":
			// We used to put everything in %HOME%\AppData\Roaming, but this caused
			// problems with non-typical setups. If this fallback location exists and
			// is non-empty, use it, otherwise DTRT and check %LOCALAPPDATA%.
			fallback := filepath.Join(home, "AppData", "Roaming", "Lachesis")
			appdata := windowsAppData()
			if appdata == "" || isNonEmptyDir(fallback) {
				return fallback
			}
			return filepath.Join(appdata, "Lachesis")
		default:
			return filepath.Join(home, ".opera")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func windowsAppData() string {
	v := os.Getenv("LOCALAPPDATA")
	if v == "" {
		// Windows XP and below don't have LocalAppData. Crash here because
		// we don't support Windows XP and undefining the variable will cause
		// other issues.
		panic("environment variable LocalAppData is undefined")
	}
	return v
}

func isNonEmptyDir(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	names, _ := f.Readdir(1)
	f.Close()
	return len(names) > 0
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
