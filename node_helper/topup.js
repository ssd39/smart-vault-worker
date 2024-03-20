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
import { getAssociatedTokenAddressSync } from "@solana/spl-token";
import State from "./state.js";

const APP_COUNTER = "APP_COUNTER";
const APP_STATE = "APP_STATE";
const USER_STATE = "USER_STATE";
const TREASURY_STATE = "TREASURY_STATE";

const mySplToken = new PublicKey(
  "5DYw4t2nJoSyhD9NDnPTveN7ZY4DwZyDXHTMJPdnqeZG"
);
const programId = new PublicKey("BJzXMqLoic3YKFJVkFr3PTfL8Myopdo6rUHHKoVirYga");

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

  const topupPayload = { amount: BigInt(10) };
  const encoded = borsh.serialize(Instructions.TopUpPayload, topupPayload);
  const instructionData = [3, ...encoded];

  const [user_state, _] = PublicKey.findProgramAddressSync(
    [Buffer.from(USER_STATE, "utf8"), myAccount.toBuffer()],
    programId
  );

  const [programm_treasury, __] = PublicKey.findProgramAddressSync(
    [Buffer.from(TREASURY_STATE, "utf8")],
    programId
  );
  const program_ata = getAssociatedTokenAddressSync(
    mySplToken,
    programm_treasury,
    true
  );

  const instructionKeys = [
    { pubkey: myAccount, isSigner: true },
    { pubkey: myAta },
    { pubkey: user_state, isWritable: true },
    { pubkey: programm_treasury },
    { pubkey: program_ata},
    { pubkey: mySplToken},
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
