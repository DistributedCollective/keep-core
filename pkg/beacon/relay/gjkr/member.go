package gjkr

import (
	"math/big"
	"strconv"

	"github.com/keep-network/keep-core/pkg/beacon/relay/pedersen"
	"github.com/keep-network/keep-core/pkg/net/ephemeral"
)

// MemberID is a unique-in-group identifier of a member.
type MemberID uint32

type memberCore struct {
	// ID of this group member.
	ID MemberID

	// Group to which this member belongs.
	group *Group

	// DKG Protocol configuration parameters.
	protocolConfig *DKG

	// evidenceLog provides access to messages from earlier protocol phases
	// for the sake of compliant resolution
	evidenceLog evidenceLog
}

// EphemeralKeyPairGeneratingMember represents one member in a distributed key
// generating group performing ephemeral key pair generation. It has a full list
// of `memberIDs` that belong to its threshold group.
//
// Executes Phase 1 of the protocol.
type EphemeralKeyPairGeneratingMember struct {
	*memberCore
	// Ephemeral key pairs used to create symmetric keys,
	// generated individually for each other group member.
	ephemeralKeyPairs map[MemberID]*ephemeral.KeyPair
}

// SymmetricKeyGeneratingMember represents one member in a distributed key
// generating group performing ephemeral symmetric key generation.
//
// Executes Phase 2 of the protocol.
type SymmetricKeyGeneratingMember struct {
	*EphemeralKeyPairGeneratingMember

	// Symmetric keys used to encrypt confidential information,
	// generated individually for each other group member by ECDH'ing the
	// broadcasted ephemeral public key intended for this member and the
	// ephemeral private key generated for the other member.
	symmetricKeys map[MemberID]ephemeral.SymmetricKey
}

// CommittingMember represents one member in a distributed key generation group,
// after it has fully initialized ephemeral symmetric keys with all other group
// members.
//
// Executes Phase 3 of the protocol.
type CommittingMember struct {
	*SymmetricKeyGeneratingMember

	// Pedersen VSS scheme used to calculate commitments.
	vss *pedersen.VSS
	// Polynomial `a` coefficients generated by the member. Polynomial is of
	// degree `dishonestThreshold`, so the number of coefficients equals
	// `dishonestThreshold + 1`
	//
	// This is a private value and should not be exposed.
	secretCoefficients []*big.Int
	// Shares calculated by the current member for themself. They are defined as
	// `s_ii` and `t_ii` respectively across the protocol specification.
	//
	// These are private values and should not be exposed.
	selfSecretShareS, selfSecretShareT *big.Int
}

// CommitmentsVerifyingMember represents one member in a distributed key generation
// group, after it has received secret shares and commitments from other group
// members and it performs verification of received values.
//
// Executes Phase 4 of the protocol.
type CommitmentsVerifyingMember struct {
	*CommittingMember

	// Shares calculated for the current member by peer group members which passed
	// the validation.
	//
	// receivedValidSharesS are defined as `s_ji` and receivedValidSharesT are
	// defined as `t_ji` across the protocol specification.
	receivedValidSharesS, receivedValidSharesT map[MemberID]*big.Int
	// Commitments to coefficients received from peer group members which passed
	// the validation.
	receivedValidPeerCommitments map[MemberID][]*big.Int
}

// SharesJustifyingMember represents one member in a threshold key sharing group,
// after it completed secret shares and commitments verification and enters
// justification phase where it resolves invalid share accusations.
//
// Executes Phase 5 of the protocol.
type SharesJustifyingMember struct {
	*CommitmentsVerifyingMember
}

// QualifiedMember represents one member in a threshold key sharing group, after
// it completed secret shares justification. The member holds a share of group
// master private key.
//
// Executes Phase 6 of the protocol.
type QualifiedMember struct {
	*SharesJustifyingMember

	// Member's share of the secret master private key. It is denoted as `z_ik`
	// in protocol specification.
	// TODO: unsure if we need shareT `x'_i` field, it should be removed if not used in further steps
	masterPrivateKeyShare, shareT *big.Int
}

// SharingMember represents one member in a threshold key sharing group, after it
// has been qualified to the master private key sharing group. A member shares
// public values of it's polynomial coefficients with peer members.
//
// Executes Phase 7 and Phase 8 of the protocol.
type SharingMember struct {
	*QualifiedMember

	// Public values of each polynomial `a` coefficient defined in secretCoefficients
	// field. It is denoted as `A_ik` in protocol specification. The zeroth
	// public key share point `A_i0` is a member's public key share.
	publicKeySharePoints []*big.Int
	// Public key share points received from peer group members which passed the
	// validation. Defined as `A_jk` across the protocol documentation.
	receivedValidPeerPublicKeySharePoints map[MemberID][]*big.Int
}

// PointsJustifyingMember represents one member in a threshold key sharing group,
// after it completed public key share points verification and enters justification
// phase where it resolves public key share points accusations.
//
// Executes Phase 9 of the protocol.
type PointsJustifyingMember struct {
	*SharingMember
}

// ReconstructingMember represents one member in a threshold sharing group who
// is reconstructing individual private and public keys of disqualified group members.
//
// Executes Phase 11 of the protocol.
type ReconstructingMember struct {
	*PointsJustifyingMember // TODO Update this when all phases of protocol are ready

	// Disqualified members' individual private keys reconstructed from shares
	// revealed by other group members.
	// Stored as `<m, z_m>`, where:
	// - `m` is disqualified member's ID
	// - `z_m` is reconstructed individual private key of member `m`
	reconstructedIndividualPrivateKeys map[MemberID]*big.Int
	// Individual public keys calculated from reconstructed individual private keys.
	// Stored as `<m, y_m>`, where:
	// - `m` is disqualified member's ID
	// - `y_m` is reconstructed individual public key of member `m`
	reconstructedIndividualPublicKeys map[MemberID]*big.Int
}

// CombiningMember represents one member in a threshold sharing group who is
// combining individual public keys of group members to receive group public key.
//
// Executes Phase 12 of the protocol.
type CombiningMember struct {
	*ReconstructingMember

	// Group public key calculated from individual public keys of all group members.
	// Denoted as `Y` across the protocol specification.
	groupPublicKey *big.Int
}

// Int converts `MemberID` to `big.Int`
func (id MemberID) Int() *big.Int {
	return new(big.Int).SetUint64(uint64(id))
}

// InitializeSymmetricKeyGeneration performs a transition of the member state
// from phase 1 to phase 2. It returns a member instance ready to execute the
// next phase of the protocol.
func (ekgm *EphemeralKeyPairGeneratingMember) InitializeSymmetricKeyGeneration() *SymmetricKeyGeneratingMember {
	return &SymmetricKeyGeneratingMember{
		EphemeralKeyPairGeneratingMember: ekgm,
		symmetricKeys:                    make(map[MemberID]ephemeral.SymmetricKey),
	}
}

// InitializeCommitting returns a member to perform next protocol operations.
func (skgm *SymmetricKeyGeneratingMember) InitializeCommitting(vss *pedersen.VSS) *CommittingMember {
	return &CommittingMember{
		SymmetricKeyGeneratingMember: skgm,
		vss: vss,
	}
}

// InitializeCommitmentsVerification returns a member to perform next protocol operations.
func (cm *CommittingMember) InitializeCommitmentsVerification() *CommitmentsVerifyingMember {
	return &CommitmentsVerifyingMember{
		CommittingMember:             cm,
		receivedValidSharesS:         make(map[MemberID]*big.Int),
		receivedValidSharesT:         make(map[MemberID]*big.Int),
		receivedValidPeerCommitments: make(map[MemberID][]*big.Int),
	}
}

// InitializeSharesJustification returns a member to perform next protocol operations.
func (cvm *CommitmentsVerifyingMember) InitializeSharesJustification() *SharesJustifyingMember {
	return &SharesJustifyingMember{cvm}
}

// InitializeQualified returns a member to perform next protocol operations.
func (sjm *SharesJustifyingMember) InitializeQualified() *QualifiedMember {
	return &QualifiedMember{SharesJustifyingMember: sjm}
}

// InitializeSharing returns a member to perform next protocol operations.
func (qm *QualifiedMember) InitializeSharing() *SharingMember {
	return &SharingMember{
		QualifiedMember:                       qm,
		receivedValidPeerPublicKeySharePoints: make(map[MemberID][]*big.Int),
	}
}

// InitializePointsJustification returns a member to perform next protocol operations.
func (sm *SharingMember) InitializePointsJustification() *PointsJustifyingMember {
	return &PointsJustifyingMember{sm}
}

// InitializeReconstruction returns a member to perform next protocol operations.
func (pjm *PointsJustifyingMember) InitializeReconstruction() *ReconstructingMember {
	return &ReconstructingMember{
		PointsJustifyingMember:             pjm,
		reconstructedIndividualPrivateKeys: make(map[MemberID]*big.Int),
		reconstructedIndividualPublicKeys:  make(map[MemberID]*big.Int),
	}
}

// InitializeCombining returns a member to perform next protocol operations.
func (rm *ReconstructingMember) InitializeCombining() *CombiningMember {
	return &CombiningMember{ReconstructingMember: rm}
}

// individualPublicKey returns current member's individual public key.
// Individual public key is zeroth public key share point `A_i0`.
func (rm *ReconstructingMember) individualPublicKey() *big.Int {
	return rm.publicKeySharePoints[0]
}

// receivedValidPeerIndividualPublicKeys returns individual public keys received
// from peer members which passed the validation. Individual public key is zeroth
// public key share point `A_j0`.
func (sm *SharingMember) receivedValidPeerIndividualPublicKeys() []*big.Int {
	var receivedValidPeerIndividualPublicKeys []*big.Int

	for _, peerPublicKeySharePoints := range sm.receivedValidPeerPublicKeySharePoints {
		receivedValidPeerIndividualPublicKeys = append(
			receivedValidPeerIndividualPublicKeys,
			peerPublicKeySharePoints[0],
		)
	}
	return receivedValidPeerIndividualPublicKeys
}

// HexString converts `MemberID` to hex `string` representation.
func (id MemberID) HexString() string {
	return strconv.FormatInt(int64(id), 16)
}

// MemberIDFromHex returns a `MemberID` created from the hex `string`
// representation.
func MemberIDFromHex(hex string) (MemberID, error) {
	id, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, err
	}

	return MemberID(id), nil
}
