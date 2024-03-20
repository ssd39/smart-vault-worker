const GetInstructionObj = (obj) => {
  return Object.assign({}, obj).struct;
};

export default {
  GetInstructionObj,
  InitPayload: {
    struct: {
      vault_public_key: "Pubkey",
      attestation_proof: "string",
    },
  },
  JoinPayload: {
    struct: {
      attestation_proof: "string",
      transit_key: "Pubkey",
      p2p_connection: "string",
    },
  },
  AddAppPayload: {
    struct: {
      rent_amount: "u64",
      ipfs_hash: "string",
    },
  },
  TopUpPayload: {
    struct: {
      amount: "u64",
    },
  },
  StartSubscriptionPayload: {
    struct: {
      max_rent: "u64",
      app_id: "u64",
      params_hash: "string",
    },
  },
  BidPayload: {
    struct: {
      _signature: "string",
      bid_amount: "u64",
    },
  },
  ClaimBidPayload: {
    struct: {
      _signature: "string",
    },
  },
  ReportWorkPayload: {
    struct: {
      nonce: "u64",
      _signature: "string",
    },
  },
};
