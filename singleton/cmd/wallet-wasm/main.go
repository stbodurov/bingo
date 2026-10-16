//go:build js && wasm

// Command wallet-wasm exposes Singleton's proof core to the browser wallet. The holder's
// seed is generated and used only inside the page; the server never receives it.
//
// Build: GOOS=js GOARCH=wasm go build -o web/static/wasm/wallet.wasm ./cmd/wallet-wasm
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"syscall/js"

	"github.com/stbodurov/bingo/singleton/internal/proof"
)

func result(v any, err error) any {
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	b, _ := json.Marshal(v)
	return map[string]any{"ok": string(b)}
}

func keyFrom(seedHex string) (*proof.KeyPair, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil || len(seed) != 32 {
		return nil, proof.ErrDecode
	}
	return proof.KeyPairFromSeed(seed), nil
}

type bucketIn struct {
	Bits    int      `json:"bits"`
	Prefix  string   `json:"prefix"`
	Members []string `json:"members"`
}

func main() {
	api := map[string]any{
		"newSeed": js.FuncOf(func(this js.Value, args []js.Value) any {
			b := make([]byte, 32)
			_, _ = rand.Read(b)
			return hex.EncodeToString(b)
		}),
		"publicKey": js.FuncOf(func(this js.Value, args []js.Value) any {
			k, err := keyFrom(args[0].String())
			if err != nil {
				return result(nil, err)
			}
			return result(map[string]string{"pubkey": k.PublicHex(), "fingerprint": proof.Fingerprint(k.Public)}, nil)
		}),
		"nullifier": js.FuncOf(func(this js.Value, args []js.Value) any {
			k, err := keyFrom(args[0].String())
			if err != nil {
				return result(nil, err)
			}
			return result(proof.Nullifier(k.KeyImage(args[1].String())), nil)
		}),
		// prove(seed, scope, challenge, epoch, bucketsJSON, orderJSON, minRing)
		"prove": js.FuncOf(func(this js.Value, args []js.Value) any {
			k, err := keyFrom(args[0].String())
			if err != nil {
				return result(nil, err)
			}
			var raw map[string]bucketIn
			if err := json.Unmarshal([]byte(args[4].String()), &raw); err != nil {
				return result(nil, err)
			}
			var order []string
			if err := json.Unmarshal([]byte(args[5].String()), &order); err != nil {
				return result(nil, err)
			}
			buckets := map[string]proof.BucketData{}
			for name, b := range raw {
				bd := proof.BucketData{Bits: b.Bits, Prefix: b.Prefix}
				for _, m := range b.Members {
					mb, err := hex.DecodeString(m)
					if err != nil || len(mb) != 32 {
						return result(nil, proof.ErrDecode)
					}
					bd.Members = append(bd.Members, mb)
				}
				buckets[name] = bd
			}
			p, err := proof.ProveBuckets(k, args[1].String(), args[2].String(), int64(args[3].Int()), buckets, order, args[6].Int())
			return result(p, err)
		}),
		"schnorr": js.FuncOf(func(this js.Value, args []js.Value) any {
			k, err := keyFrom(args[0].String())
			if err != nil {
				return result(nil, err)
			}
			return result(k.SignSchnorr([]byte(args[1].String())).String(), nil)
		}),
		"split": js.FuncOf(func(this js.Value, args []js.Value) any {
			seed, err := hex.DecodeString(args[0].String())
			if err != nil {
				return result(nil, err)
			}
			shares, err := proof.Split(seed, args[1].Int(), args[2].Int())
			if err != nil {
				return result(nil, err)
			}
			out := make([]string, len(shares))
			for i, s := range shares {
				out[i] = s.String()
			}
			return result(out, nil)
		}),
		"combine": js.FuncOf(func(this js.Value, args []js.Value) any {
			var in []string
			if err := json.Unmarshal([]byte(args[0].String()), &in); err != nil {
				return result(nil, err)
			}
			var shares []proof.Share
			for _, s := range in {
				sh, err := proof.ParseShare(s)
				if err != nil {
					return result(nil, err)
				}
				shares = append(shares, sh)
			}
			seed, err := proof.Combine(shares)
			if err != nil {
				return result(nil, err)
			}
			return result(hex.EncodeToString(seed), nil)
		}),
		// Agent keys are Ed25519 so any server language can verify agent requests.
		"agentKey": js.FuncOf(func(this js.Value, args []js.Value) any {
			pub, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return result(nil, err)
			}
			return result(map[string]string{"public": hex.EncodeToString(pub), "private": hex.EncodeToString(priv.Seed())}, nil)
		}),
		"agentSign": js.FuncOf(func(this js.Value, args []js.Value) any {
			seed, err := hex.DecodeString(args[0].String())
			if err != nil || len(seed) != ed25519.SeedSize {
				return result(nil, proof.ErrDecode)
			}
			sig := ed25519.Sign(ed25519.NewKeyFromSeed(seed), []byte(args[1].String()))
			return result(base64.RawURLEncoding.EncodeToString(sig), nil)
		}),
		"delegationChallenge": js.FuncOf(func(this js.Value, args []js.Value) any {
			var d proof.Delegation
			if err := json.Unmarshal([]byte(args[0].String()), &d); err != nil {
				return result(nil, err)
			}
			return result(d.Challenge(), nil)
		}),
	}
	js.Global().Set("SingletonCore", js.ValueOf(api))
	js.Global().Call("dispatchEvent", js.Global().Get("CustomEvent").New("singleton-core-ready"))
	select {}
}
