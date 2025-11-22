package server

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"sync"
)

var (
	rng   *mrand.Rand
	rngMu sync.Mutex
)

func init() {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(fmt.Sprintf("failed to read random bytes: %v", err))
	}
	//nolint:gosec // G115: Integer overflow is fine for seeding
	seed := int64(binary.LittleEndian.Uint64(b[:]))
	//nolint:gosec // G404: Math/rand is seeded with crypto/rand, sufficient for port selection
	rng = mrand.New(mrand.NewSource(seed))
}

// GetRandomPort returns a random port in the given range [minPort, maxPort] that is not in the excluded map.
// Note: This does not check if the port is actually free on the network interface,
// as we rely on Docker's state (passed via excluded map) to determine availability on the host.
func GetRandomPort(minPort, maxPort int, excluded map[int]bool) (int, error) {
	if minPort > maxPort {
		return 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	// Create a list of ports to try
	count := maxPort - minPort + 1
	ports := make([]int, count)
	for i := 0; i < count; i++ {
		ports[i] = minPort + i
	}

	// Shuffle the ports
	rngMu.Lock()
	rng.Shuffle(len(ports), func(i, j int) {
		ports[i], ports[j] = ports[j], ports[i]
	})
	rngMu.Unlock()

	// Try ports until one works
	for _, port := range ports {
		// Skip if explicitly excluded
		if excluded != nil && excluded[port] {
			continue
		}

		return port, nil
	}

	return 0, fmt.Errorf("no available ports found in range %d-%d (all excluded)", minPort, maxPort)
}
