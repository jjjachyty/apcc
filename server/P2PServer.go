package server

import (
	"context"
	"sync"

	"apcchis.com/apcc/common"
	discovery "github.com/libp2p/go-libp2p-discovery"
	host "github.com/libp2p/go-libp2p-host"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

var stm inet.Stream

func Start() {

	// 当前节点IP地址

	// 挖矿节点设置
	// if len(minerAdd) > 0 {

	// 	miningAddress = minerAdd
	// }
	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all logruss with:
	// golog.SetAlllogruss(gologging.INFO) // Change to DEBUG for extra info

	// Parse options from the command line

	// secio := flag.Bool("secio", false, "enable secio")
	// seed := flag.Int64("seed", 0, "set random seed for id generation")

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(common.ListenPort, true, int64(common.ListenPort))
	if err != nil {
		logrus.Error(err)
	}

	logrus.Debug("listening for connections")
	// Set a stream handler on host A. /p2p/1.0.0 is
	// a user-defined protocol name.
	ha.SetStreamHandler(protocol.ID(common.ProtocolID), handleStream)
	stm = KadDht(context.Background(), ha)

	// The following code extracts target's peer ID from the
	// given multiaddress
	if nil != stm {

		// Create a thread to read and write data.
		sendVersion(common.ServerNodeAddrs[0])

		go readData()
	}

	select {}

}

func KadDht(ctx context.Context, host host.Host) inet.Stream {
	var stm inet.Stream
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := libp2pdht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	logrus.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, addrString := range common.ServerNodeAddrs {
		peerAddr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			logrus.Error("格式化BootstrapPeers地址出错")
		}

		peerinfo, _ := peerstore.InfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				// logrus.Warning(err)
			} else {
				logrus.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	logrus.Info("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, common.Ns)
	logrus.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	logrus.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, common.Ns)
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}
		logrus.Debug("Found peer:", peer)

		logrus.Debug("Connecting to:", peer)
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(common.ProtocolID))

		if err != nil {
			logrus.Warning("Connection failed:", err)
			continue
		} else {
			// rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			// go writeData(rw)
			// go readData(rw)
			stm = stream
			logrus.Infof("Connected to:%v", peer)
			break
		}

	}
	return stm
}
