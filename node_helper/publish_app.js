import fs from "fs";
import {
  PublicKey,
  Keypair,
  Connection,
  SystemProgram,
  TransactionMessage,
  VersionedTransaction,
  TransactionInstruction,
} from "@solana/web3.js";
import * as borsh from "borsh";
import Instructions from "./instructions.js";
import State from "./state.js";

const APP_COUNTER = "APP_COUNTER";
const APP_STATE = "APP_STATE";

const mySplToken = new PublicKey(
  "5DYw4t2nJoSyhD9NDnPTveN7ZY4DwZyDXHTMJPdnqeZG"
);
const programId = new PublicKey("6bcSZLTvfu2ZaC7yhXfkaupFG315r4qWK8wqSQN5LRFT");

const keyPairBytes = JSON.parse(fs.readFileSync("keypair.json"));

const keyPair = Keypair.fromSecretKey(new Uint8Array(keyPairBytes));
const myAccount = new PublicKey("DDWnhLpBMVAXGrYNPZKu8wbQdt7iUXfc1fJVWmxTkCc");
const myAta = new PublicKey("45U4tSGsMgNASYn5NFX3remd5iRpXBLXLKfEWDwepyYN");

function bigIntToBuffer(bigInt) {
  const buffer = new ArrayBuffer(8); // Assuming maximum 8 bytes required
  const view = new DataView(buffer);

  // Assuming your BigInt is less than 2^64
  view.setBigUint64(0, BigInt.asUintN(64, bigInt), false); // Use true for little-endian, false for big-endian

  return buffer;
}

async function publish_app() {
  const connection = new Connection("http://127.0.0.1:8899", "confirmed");

  const addAppPayload = { rent_amount: BigInt(2), ipfs_hash: "test" };
  const encoded = borsh.serialize(Instructions.AddAppPayload, addAppPayload);
  const instructionData = [2, ...encoded];

  const [app_counter, _] = PublicKey.findProgramAddressSync(
    [Buffer.from(APP_COUNTER, "utf8")],
    programId
  );

  //const ata_account_info = await connection.getAccountInfo(myAta);

  const app_counter_info = await connection.getAccountInfo(app_counter);

  const app_counter_data = borsh.deserialize(
    State.VaultAppCounterState,
    app_counter_info.data
  );
  console.log(app_counter_data)
  console.log(bigIntToBuffer(app_counter_data.counter));
  const [app_state, __] = PublicKey.findProgramAddressSync(
    [Buffer.from(APP_STATE, "utf8"), bigIntToBuffer(app_counter_data.counter)],
    programId
  );

  const instructionKeys = [
    { pubkey: myAccount, isSigner: true },
    { pubkey: myAta },
    { pubkey: app_counter, isWritable: true },
    { pubkey: app_state, isWritable: true },
    { pubkey: SystemProgram.programId },
  ];

  // Create a TransactionInstruction
  const transactionInstruction = new TransactionInstruction({
    keys: instructionKeys,
    programId: programId,
    data: Buffer.from(instructionData),
  });
  const blockhash = (await connection.getLatestBlockhash("finalized"))
    .blockhash;
  const messageV0 = new TransactionMessage({
    payerKey: myAccount,
    recentBlockhash: blockhash,
    instructions: [transactionInstruction],
  }).compileToV0Message();
  const transaction = new VersionedTransaction(messageV0);
  transaction.sign([keyPair]);
  const signature = await connection.sendTransaction(transaction);
  console.log(signature);
}

publish_app();
