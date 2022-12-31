package helper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/stovegi/stove-helper/pkg/rand/csharp"
	"github.com/stovegi/stove-helper/pkg/rand/mt19937"
)

type KeyStore struct {
	sync.Mutex
	keyMap map[uint16][]byte
}

func (s *Service) xor(p []byte) {
	s.keyStore.Lock()
	key := s.keyStore.keyMap[binary.BigEndian.Uint16(p)]
	if key == nil {
		seed := s.cfg.Seed
		if s.serverSeed == 0 {
			s.serverSeed = s.cfg.ServerSeed
		}
		if seed != 0 {
			seed, key = bruteforce(seed, s.serverSeed, p)
		}
		if seed == 0 || key == nil {
			seed, key = bruteforce(s.sentMs, s.serverSeed, p)
			if key == nil {
				s.keyStore.Unlock()
				return
			}
		}
		if s.cfg.Seed == 0 {
			s.cfg.Seed = seed
		}
		fmt.Fprintf(s.file, "- seed: %d\n  serverSeed: %d\n", seed, s.serverSeed)
		s.keyStore.keyMap[binary.BigEndian.Uint16(p)] = key
	}
	s.keyStore.Unlock()
	if key != nil {
		xor(p, key)
	}
}

func xor(p, key []byte) {
	for i := 0; i < len(p); i++ {
		p[i] ^= key[i%4096]
	}
}

// bruteforce is a method to find the encryption seed of the client.
func bruteforce(ms, seed uint64, p []byte) (uint64, []byte) {
	r1 := csharp.NewRand()
	r2 := mt19937.NewRand()
	v := binary.BigEndian.Uint64(p)
	for i := uint64(0); i < 1000; i++ {
		r1.Seed(int64(ms + i))
		for j := uint64(0); j < 1000; j++ {
			s := r1.Uint64()
			r2.Seed(int64(s ^ seed))
			r2.Seed(int64(r2.Uint64()))
			r2.Uint64()
			if (v^r2.Uint64())&0xFFFF0000FF00FFFF == 0x4567000000000000 {
				log.Info().Uint64("#seed", ms+i).Uint64("depth", j).Msg("Found seed")
				return ms + i, mt19937.NewKeyBlock(s ^ seed).Key()
			}
			if i != 0 && (i > 100 || i+j > 100) {
				break
			}
		}
		r1.Seed(int64(ms - i - 1))
		for j := uint64(0); j < 1000; j++ {
			s := r1.Uint64()
			r2.Seed(int64(s ^ seed))
			r2.Seed(int64(r2.Uint64()))
			r2.Uint64()
			if (v^r2.Uint64())&0xFFFF0000FF00FFFF == 0x4567000000000000 {
				log.Info().Uint64("#seed", ms-i-1).Uint64("depth", j).Msg("Found seed")
				return ms - i - 1, mt19937.NewKeyBlock(s ^ seed).Key()
			}
			if i+1 > 100 || i+j+1 > 100 {
				break
			}
		}
	}
	return 0, nil
}

// clientwind is a method to find the encryption key of the client,
// which needs the cleartext of the WindSeedClientNotify packet.
func clientwind(f *os.File, p []byte) []byte {
	// not tested, use with caution
	f.Seek(0, 0)
	data, _ := io.ReadAll(f)
	c := make([]byte, 2*4096) // 2 blocks
	for i := 0; i < 4096; i++ {
		copy(c, data[i:])
		for j := 0; j < 2*4096; j++ {
			c[j] = c[j] ^ p[4096+j]
		}
		if !bytes.Equal(c[:4096], c[4096:]) {
			continue
		}
		if c[0] == p[0]^0x45 && c[1] == p[1]^0x67 && c[4] == p[4] && c[6] == p[6] && c[7] == p[7] {
			key := make([]byte, 4096)
			copy(key, c)
			log.Info().Bytes("#key", key).Msg("Found key")
			return key
		}
	}
	return nil
}

// searchvmem is another method to find the encryption key of the client,
// which needs the permission to read the memory of the process.
func searchvmem(f *os.File, p []byte) []byte {
	// not tested, use with caution
	f.Seek(0, 0)
	var n int64
	c := make([]byte, 4<<20) // 4 MiB
	for {
		m, err := f.Read(c[:cap(c)])
		if err != nil {
			break
		}
		c = c[:m]
		n += int64(m)
		for i := 0; i < 4<<20; i += 16 {
			if c[i] == p[0]^0x45 && c[i+1] == p[1]^0x67 && c[i+4] == p[4] && c[i+6] == p[6] && c[i+7] == p[7] {
				key := make([]byte, 4096)
				copy(key, c[i:])
				log.Info().Bytes("#key", key).Msg("Found key")
				return key
			}
		}
	}
	return nil
}
