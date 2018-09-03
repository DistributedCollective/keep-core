package tecdsa

import (
	"github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/tecdsa/commitment"
	"github.com/keep-network/keep-core/pkg/tecdsa/curve"
	"github.com/keep-network/keep-core/pkg/tecdsa/zkp"
	"github.com/keep-network/paillier"
)

// SignRound1Message is a message produced by each signer as a result of
// executing the first round of T-ECDSA signing algorithm.
type SignRound1Message struct {
	signerID string

	secretKeyFactorShareCommitment *commitment.MultiTrapdoorCommitment // C_1i
}

// SignRound2Message is a message produced by each signer as a result of
// executing the second round of T-ECDSA signing algorithm.
type SignRound2Message struct {
	signerID string

	secretKeyFactorShare                *paillier.Cypher            // u_i = E(ρ_i)
	secretKeyMultipleShare              *paillier.Cypher            // v_i = E(ρ_i * x)
	secretKeyFactorShareDecommitmentKey *commitment.DecommitmentKey // D_1i

	secretKeyFactorProof *zkp.DsaPaillierSecretKeyFactorRangeProof // Π_1i
}

// isValid checks secret key random factor share and secret key multiple share
// against the zero knowledge proof shipped alongside them as well as validates
// commitment generated by signer in the first round.
func (msg *SignRound2Message) isValid(
	commitmentMasterPublicKey *bn256.G2, // h
	secretKeyFactorShareCommitment *commitment.MultiTrapdoorCommitment, // C_1i
	dsaSecretKey *paillier.Cypher, // E(x)
	zkpParams *zkp.PublicParameters,
) bool {
	commitmentValid := secretKeyFactorShareCommitment.Verify(
		commitmentMasterPublicKey,
		msg.secretKeyFactorShareDecommitmentKey,
		msg.secretKeyFactorShare.C.Bytes(),
		msg.secretKeyMultipleShare.C.Bytes(),
	)

	zkpValid := msg.secretKeyFactorProof.Verify(
		msg.secretKeyMultipleShare, dsaSecretKey, msg.secretKeyFactorShare, zkpParams,
	)

	return commitmentValid && zkpValid
}

// SignRound3Message is a message produced by each signer as a result of
// executing the third round of T-ECDSA signing algorithm.
type SignRound3Message struct {
	signerID string

	signatureFactorShareCommitment *commitment.MultiTrapdoorCommitment // C_2i
}

// SignRound4Message is a message produced by each signer as a result of
// executing the fourth round of T-ECDSA signing algorithm.
type SignRound4Message struct {
	signerID string

	signatureFactorPublicShare          *curve.Point                // r_i = g^{k_i}
	signatureUnmaskShare                *paillier.Cypher            // w_i = E(k_i * ρ + c_i * q)
	signatureFactorShareDecommitmentKey *commitment.DecommitmentKey // D_2i

	signatureFactorProof *zkp.EcdsaSignatureFactorRangeProof // Π_2i
}

// isValid checks the signature random multiple public share and signature
// unmask share against the zero knowledge proof shipped alongside them. It
// also validates commitment generated by the signer in the third round.
func (msg *SignRound4Message) isValid(
	commitmentMasterPublicKey *bn256.G2, // h
	signatureFactorShareCommitment *commitment.MultiTrapdoorCommitment, // C_2i
	secretKeyFactor *paillier.Cypher, // u = E(ρ)
	zkpParams *zkp.PublicParameters,
) bool {
	commitmentValid := signatureFactorShareCommitment.Verify(
		commitmentMasterPublicKey,
		msg.signatureFactorShareDecommitmentKey,
		msg.signatureFactorPublicShare.Bytes(),
		msg.signatureUnmaskShare.C.Bytes(),
	)

	zkpValid := msg.signatureFactorProof.Verify(
		msg.signatureFactorPublicShare,
		msg.signatureUnmaskShare,
		secretKeyFactor,
		zkpParams,
	)

	return commitmentValid && zkpValid
}

// SignRound5Message is a message produced by each signer as a result of
// executing the fifth round of T-ECDSA signing algorithm.
type SignRound5Message struct {
	signerID string

	signatureUnmaskPartialDecryption *paillier.PartialDecryption // TDec(w)
}

// SignRound6Message is a message produced by each signer as a result of
// executing the sixth round of T-ECDSA signing algorithm.
type SignRound6Message struct {
	signerID string

	signaturePartialDecryption *paillier.PartialDecryption // TDec(σ)
}