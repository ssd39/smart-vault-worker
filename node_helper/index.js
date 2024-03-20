import * as web3 from '@solana/web3.js'
import fs from 'fs'

// 6pJUGwn9Jq55wdEVUhAL3iGTFPQEnH61Q99wCKLZCf4f
const keypair = web3.Keypair.generate()
console.log(`PublicKey: ${keypair.publicKey.toBase58()}`)
fs.writeFileSync("./.walletKey", keypair.secretKey)

fs.writeFileSync("./keypair.json", JSON.stringify(keypair))