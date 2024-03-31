import * as web3 from "@solana/web3.js";
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
function findFirstDifferenceIndex(arr1, arr2) {
  const length = Math.min(arr1.length, arr2.length);

  for (let i = 0; i < length; i++) {
    if (arr1[i] !== arr2[i]) {
      return i;
    }
  }

  // If all bytes are the same up to the minimum length, check for extra bytes
  if (arr1.length !== arr2.length) {
    return Math.min(arr1.length, arr2.length);
  }

  // If arrays are identical, return -1
  return -1;
}

let ins = web3.Ed25519Program.createInstructionWithPublicKey({
  publicKey: new Uint8Array([
    77, 180, 223, 144, 109, 40, 103, 13, 103, 238, 122, 152, 192, 202, 9, 35,
    178, 49, 54, 206, 182, 82, 44, 72, 183, 155, 123, 219, 178, 160, 123, 57,
  ]),
  signature: new Uint8Array([
    167, 22, 112, 179, 189, 62, 145, 193, 69, 178, 213, 113, 197, 78, 15, 224,
    118, 253, 109, 52, 198, 25, 90, 65, 183, 149, 255, 52, 252, 31, 90, 125, 67,
    3, 89, 110, 182, 240, 199, 62, 213, 142, 54, 55, 9, 52, 34, 61, 78, 115, 69,
    43, 1, 189, 121, 39, 23, 92, 168, 85, 152, 103, 227, 10,
  ]),
  message: new Uint8Array([
    77, 180, 223, 144, 109, 40, 103, 13, 103, 238, 122, 152, 192, 202, 9, 35,
    178, 49, 54, 206, 182, 82, 44, 72, 183, 155, 123, 219, 178, 160, 123, 57, 0,
    0, 0, 0, 0, 0, 0, 0,
  ]),
});
let x = new Uint8Array([
  1, 0, 48, 0, 255, 255, 16, 0, 255, 255, 112, 0, 40, 0, 255, 255, 77, 180, 223,
  144, 109, 40, 103, 13, 103, 238, 122, 152, 192, 202, 9, 35, 178, 49, 54, 206,
  182, 82, 44, 72, 183, 155, 123, 219, 178, 160, 123, 57, 167, 22, 112, 179,
  189, 62, 145, 193, 69, 178, 213, 113, 197, 78, 15, 224, 118, 253, 109, 52,
  198, 25, 90, 65, 183, 149, 255, 52, 252, 31, 90, 125, 67, 3, 89, 110, 182,
  240, 199, 62, 213, 142, 54, 55, 9, 52, 34, 61, 78, 115, 69, 43, 1, 189, 121,
  39, 23, 92, 168, 85, 152, 103, 227, 10, 77, 180, 223, 144, 109, 40, 103, 13,
  103, 238, 122, 152, 192, 202, 9, 35, 178, 49, 54, 206, 182, 82, 44, 72, 183,
  155, 123, 219, 178, 160, 123, 57, 0, 0, 0, 0, 0, 0, 0, 0,
]);
let y = new Uint8Array(ins.data);

const keyPairBytes = JSON.parse(fs.readFileSync("keypair.json"));

const keyPair = Keypair.fromSecretKey(new Uint8Array(keyPairBytes));
const myAccount = new PublicKey("DDWnhLpBMVAXGrYNPZKu8wbQdt7iUXfc1fJVWmxTkCc");

function bigIntToBuffer(bigInt) {
  const buffer = new ArrayBuffer(8); // Assuming maximum 8 bytes required
  const view = new DataView(buffer);

  // Assuming your BigInt is less than 2^64
  view.setBigUint64(0, BigInt.asUintN(64, bigInt), false); // Use true for little-endian, false for big-endian

  return buffer;
}

async function ed2599test() {
  if (findFirstDifferenceIndex(x, y) == -1) {
    const connection = new Connection("http://127.0.0.1:8899", "confirmed");
    console.log(keyPair)
    const newIns = web3.Ed25519Program.createInstructionWithPrivateKey({
      message: new Uint8Array([77,180,223,144,109,40,103,13,103,238,122,152,192,202,9,35,178,49,54,206,182,82,44,72,183,155,123,219,178,160,123,57,0,0,0,0,0,0,0,0]),
      privateKey: keyPair.secretKey,
    });
    const blockhash = (await connection.getLatestBlockhash("finalized"))
      .blockhash;
    const messageV0 = new TransactionMessage({
      payerKey: myAccount,
      recentBlockhash: blockhash,
      instructions: [newIns],
    }).compileToV0Message();
    const transaction = new VersionedTransaction(messageV0);
    transaction.sign([keyPair]);
    const signature = await connection.sendTransaction(transaction);
    console.log(signature);
  } else {
    console.log("no luck");
  }
}

ed2599test();
