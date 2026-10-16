# Singleton protocol v1

## 1. Goals
Give relying parties (RPs) three answers — *is this a real human?*, *have I seen this human
before?*, *does this human satisfy predicate P (e.g. over 16)?* — while revealing nothing else
to the RP, to other RPs, or to Singleton.

## 2. Parties
- **Holder** — a person with a device-held key.
- **Attester** — verifies a credential (mDL, ePassport, EUDI PID, in-person, vouching) once.
- **Registry** — keeps public keys of enrolled holders, keyed de-duplication hashes, threshold
  dates; seals epochs; anchors set roots on-chain.
- **Verifier** — checks presentations (Singleton's hosted verifier or an RP running the
  open-source verifier against the published sets).
- **RP** — the app asking.

## 3. Primitives
- Group 𝔾 = ristretto255 (prime order ℓ), generator G.
- `H_s(scope) = HashToGroup("singleton/scope/v1" ‖ scope)` (Elligator 2 over SHA-512).
- `H(·)` → scalar: SHA-512 with length-prefixed domain separation, reduced mod ℓ.
- Holder secret `x = H("singleton/holder-key/v1" ‖ seed)`, public `P = x·G`.

## 4. Enrollment
1. Holder generates `seed` on device (WebAssembly build of the same Go code as the verifier).
2. Holder presents a credential to an attester. The registry verifies the issuer signature.
3. Registry computes `doc = HMAC_k("doc" ‖ country ‖ type ‖ number)` and
   `person = HMAC_k("person" ‖ norm(given) ‖ norm(family) ‖ dob ‖ country)`; if either matches an
   active enrollment, enrollment is refused (one human, one pass). `k` lives in an HSM.
4. For each age threshold n ∈ {13, 16, 18, 21} the registry stores the crossing date, or a
   sentinel if already crossed — adults' dates of birth are never retained.
5. `P` is stored with assurance level (IAL2 for documents/in-person, IAL1 for vouching).

Vouching: a newcomer is enrolled at IAL1 after three presentations of `human_doc` in scope
`singleton:vouch` with challenge `vouch:P_new`. The voucher's key image in that scope rate-limits
vouches per pseudonym (5/month) without identifying the voucher.

## 5. Epochs and sets
Every epoch the registry snapshots predicate sets
`S_human, S_human_doc, S_age≥13/16/18/21, S_region∈{us,gb,au,eu}` and commits each with a Merkle
tree over sorted leaves `SHA-256(0x00 ‖ P)`, internal nodes `SHA-256(0x01 ‖ L ‖ R)`. Roots are
anchored by `HumanRegistry.anchorEpoch(epoch, setIds, roots, sizes)`. Verifiers accept the current
epoch and a small grace window.

## 6. Anonymity buckets
Ring signatures are O(n). v1 signs over a canonical bucket: holders are placed by
`B(P) = SHA-256("singleton/bucket/v1" ‖ P)`; a bucket at depth b is every member sharing the first
b bits of B(P). The prover uses the deepest b ≤ ⌊log₂(|S|/64)⌋ whose bucket has ≥ 32 members.
Verifiers rebuild the ring from the published set and reject buckets below the minimum. Because
B(P) does not depend on the epoch, repeated proofs over time can be intersected to at most the
bucket — never to the holder.

## 7. Presentation
For RP scope s (optionally `s#action`), challenge c (fresh per session), epoch e and each
requested predicate set S:

- Ring R = bucket of P in S at epoch e, sorted.
- Message m = "singleton/presentation/v1" ‖ s ‖ c ‖ e ‖ S ‖ b ‖ prefix.
- Key image I = x·H_s(s).
- bLSAG over R with base points (G, H_s(s)):
  - α ←$ ℤℓ; L_π = α·G; R_π = α·H_s; c_{π+1} = H(m, s, digest(R), I, L_π, R_π)
  - for i ≠ π: r_i ←$ ℤℓ; L_i = r_i·G + c_i·P_i; R_i = r_i·H_s + c_i·I; c_{i+1} = H(…, L_i, R_i)
  - r_π = α − c_π·x
  - σ = (I, c_0, r_0 … r_{n−1})

A presentation carries one σ per predicate. The verifier recomputes every ring, verifies every
σ, and requires **all key images equal** (one holder). The RP's pseudonym is
`N = SHA-256("singleton/nullifier/v1" ‖ I)`.

## 8. Properties (informal)
| Property | Argument |
|---|---|
| Soundness | bLSAG is a 1-of-n proof of knowledge of x with P_j = x·G and I = x·H_s; a non-member cannot sign. |
| Linkability in scope | I depends only on x and s: a holder cannot produce two different I for one scope (it would require two discrete logs relative to H_s). |
| Unlinkability across scopes | Given x·H_s1 and x·H_s2 for independent H_s1, H_s2, linking is DDH-hard. |
| Signer anonymity | Within the bucket, the transcript is simulatable; Singleton's verifier learns only the bucket. |
| Replay resistance | c is session-bound and single-use; m binds scope, challenge, epoch, set and bucket. |
| Splicing resistance | Predicate proofs from different holders have different I and are rejected. |

## 9. Delegation (Know-Your-Agent)
A holder authorises an agent key K_a inside scope s with statement
`D = {v, scope, agent, label, permissions, issuedAt, notAfter}` by presenting `human` with
challenge `SHA-256("singleton/delegation/v1" ‖ JSON(D))`. The RP stores (N, K_a, permissions).
Agent requests are Ed25519-signed and verified against K_a; revocation and expiry are enforced
server-side.

## 10. Recovery and revocation
- The 32-byte seed is split with Shamir's scheme over GF(2⁸) (3-of-5 by default). Recovery
  restores x, hence every pseudonym.
- A holder revokes with a Schnorr signature on `"singleton/revoke/v1:" ‖ P`. Attester fraud
  revokes every key that attester minted. Revoked keys leave the next epoch; a commitment to
  the revocation list is anchored on-chain.

## 11. Roadmap
- v2: zk-SNARK (Semaphore-style) membership over the whole set → O(1) proofs, full-set anonymity.
- Native wallets with secure-enclave keys; ISO 18013-7 online presentation of mDLs.
- Decentralised registry operators (threshold-signed epochs) and an open attester marketplace.
