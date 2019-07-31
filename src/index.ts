import http from "http";
import crypto from "crypto";

const PORT: number = Number(process.env.port) || 5555; // server port
const KEY: string = process.env.secret || ""; // write your secret key here

/**
 * judge if two buffer string same
 * @param a buffer string
 * @param b buffer string
 */
function eq(a: Buffer, b: Buffer): boolean {
  if (a.length !== b.length) {
    return false;
  }

  let res = 0;
  for (let i: number = 0; i < a.length; ++i) {
    res |= a[i] ^ b[i];
  }
  return res === 0;
}


/**
 * check data in buffer
 * @param sign get the data from request header
 * @param data get the response data to check the header
 */
function bufferEqual(sign: any, data: any): boolean {
  return eq(
    Buffer.from(sign),
    Buffer.from(
      "sha1=".concat(crypto
        .createHmac("sha1", KEY)
        .update(data)
        .digest("hex"))
    )
  );
}


// create server
http
  .createServer(function (req, res) {
    req.on("data", async function (data: Buffer) {
      try {
        if (bufferEqual(req.headers["x-hub-signature"], data)) {
          let tmp = JSON.parse(data.toString());
          console.log(tmp['repository']['clone_url']);
        }
        else {
          // FIXME: need to log in the server
        }
      } catch (error) {
        console.log(error)
      }
    });
    res.end();
  })
  .listen(PORT);
