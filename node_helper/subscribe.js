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
const USER_STATE = "USER_STATE";
const SUB_STATE = "SUB_STATE";

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

async function subscribe() {
  const connection = new Connection("http://127.0.0.1:8899", "confirmed");

  const startSubscriptionPayload = {
    max_rent: BigInt(1),
    app_id: BigInt(0),
    params_hash: "test",
  };
  const encoded = borsh.serialize(
    Instructions.StartSubscriptionPayload,
    startSubscriptionPayload
  );
  const instructionData = [4, ...encoded];

  const [user_state, _] = PublicKey.findProgramAddressSync(
    [Buffer.from(USER_STATE, "utf8"), myAccount.toBuffer()],
    programId
  );

  const user_state_info = await connection.getAccountInfo(user_state);

  const user_state_data = borsh.deserialize(
    State.VaultUserState,
    user_state_info.data
  );
  console.log(user_state_data);

  const [sub_state, __] = PublicKey.findProgramAddressSync(
    [
      Buffer.from(SUB_STATE, "utf8"),
      myAccount.toBuffer(),
      bigIntToBuffer(user_state_data.count),
    ],
    programId
  );

  console.log(sub_state);

  const [app_state, ___] = PublicKey.findProgramAddressSync(
    [Buffer.from(APP_STATE, "utf8"), bigIntToBuffer(BigInt(0))],
    programId
  );

  const instructionKeys = [
    { pubkey: myAccount, isSigner: true },
    { pubkey: user_state, isWritable: true },
    { pubkey: sub_state, isWritable: true },
    { pubkey: app_state },
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

subscribe();
