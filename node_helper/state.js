export default{
    VaultMetaDataState: {
        struct: {
            is_initialized: 'bool',
            attestation_proof: 'string',
            vault_public_key: 'Pubkey'
        }
    },
    VaultAppCounterState: {
        struct: {
            is_initialized: 'bool',
            counter: 'u64'
        }
    },
    VaultAppState: {
        struct: {
            is_initialized: 'bool',
            ipfs_hash: 'string',
            rent: 'u64',
            creator_ata: 'Pubkey'
        }
    },
    VaultUserState: {
        struct: {
            is_initialized: 'bool',
            count: 'u64',
            balance: 'u64'
        }
    },
    VaultUserSubscriptionState: {
        struct: {
            id: 'u64',
            is_initialized: 'bool',
            closed: 'bool',
            app_id: 'u64',
            params_hash: 'string',
            max_rent: 'u64',
            is_assigned: 'bool',
            executor: 'Pubkey',
            bid_endtime: 'u64',
            rent: 'u64',
            nonce: 'u64',
            last_report_time: 'u64',
            restart: 'bool'
        }
    },
    VaultBidderState: {
        struct: {
            is_initialized: 'bool',
            nonce: 'u64'
        }
    }
};

