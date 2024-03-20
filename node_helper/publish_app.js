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
  "G5KD3cp7YWQoRrCaWRJ7WyjuyWj1VDLhnorX4i3BuYTD"
);
const programId = new PublicKey("DmrxZbSZF84Pa9dnrg23FsvZffZpovfgi5pjqDKoDxom");

const keyPairBytes = JSON.parse(fs.readFileSync("keypair.json"));

const keyPair = Keypair.fromSecretKey(new Uint8Array(keyPairBytes));
const myAccount = new PublicKey("7j7qc7dtHomyQwcVTipUa1vCJxgFDjcBCKbFzdMY4j1U");
const myAta = new PublicKey("8RmdMAdraB3YXQ9Vu54zTJjMKdZoXFEMjEEm4jtCsAtW");

async function publish_app() {
  const connection = new Connection("http://127.0.0.1:8899", "confirmed");

  const addAppPayload = { rent_amount: BigInt(2), ipfs_hash: "test" };
  const encoded = borsh.serialize(Instructions.AddAppPayload, addAppPayload);
  const instructionData = [2, ...encoded];

  const [app_counter, _] = PublicKey.findProgramAddressSync(
    [Buffer.from(APP_COUNTER, "utf8")],
    programId
  );

  const ata_account_info = await connection.getAccountInfo(myAta);
  console.log(ata_account_info);
  const app_counter_info = await connection.getAccountInfo(app_counter);
  const app_counter_data = borsh.deserialize(
    State.VaultAppCounterState,
    app_counter_info.data
  );
  const [app_state, __] = PublicKey.findProgramAddressSync(
    [
      Buffer.from(APP_STATE, "utf8"),
      BigUint64Array.from(app_counter_data.counter).buffer,
    ],
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
